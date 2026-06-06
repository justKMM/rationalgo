package algorand

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"

	"github.com/algonode/go-algorand-hd-wallet/XHDWalletAPI"

	"rationalgo/internal/config"
	"rationalgo/internal/util"
	"rationalgo/pkg/provenance"
)

// Client wraps testnet Algod access and transaction signing.
type Client struct {
	algod    *algod.Client
	account  crypto.Account
	hdSigner *XHDWalletAPI.AlgoPath
}

// NewClient connects to Algorand Testnet and validates wallet credentials.
func NewClient(cfg config.Config) (*Client, error) {
	if err := cfg.ValidateForSpike(); err != nil {
		return nil, err
	}

	algodClient, err := algod.MakeClient(cfg.AlgodURL, cfg.AlgodToken)
	if err != nil {
		return nil, fmt.Errorf("algod client: %w", err)
	}

	creds, err := util.ResolveWallet(cfg.Mnemonic, cfg.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("mnemonic: %w", err)
	}

	client := &Client{
		algod:   algodClient,
		account: crypto.Account{Address: creds.Address},
	}
	if creds.HDPath != nil {
		client.hdSigner = creds.HDPath
	} else {
		client.account.PrivateKey = creds.LegacyKey
	}

	return client, nil
}

// Address returns the wallet address derived from configuration.
func (c *Client) Address() string {
	return c.account.Address.String()
}

// AccountInfo fetches on-chain account metadata.
func (c *Client) AccountInfo() (models.Account, error) {
	return c.algod.AccountInformation(c.account.Address.String()).Do(context.Background())
}

// SuggestedParams returns current algod suggested transaction parameters.
func (c *Client) SuggestedParams(ctx context.Context) (types.SuggestedParams, error) {
	params, err := c.algod.SuggestedParams().Do(ctx)
	if err != nil {
		return types.SuggestedParams{}, fmt.Errorf("suggested params: %w", err)
	}
	return params, nil
}

// NetworkIdentifier returns the x402 CAIP-2 network id for the connected algod node.
func (c *Client) NetworkIdentifier(ctx context.Context) (string, error) {
	params, err := c.SuggestedParams(ctx)
	if err != nil {
		return "", err
	}
	return "algorand:" + base64.StdEncoding.EncodeToString(params.GenesisHash[:]), nil
}

// SignTxn signs a transaction and returns msgpack-encoded signed txn bytes.
func (c *Client) SignTxn(txn types.Transaction) ([]byte, error) {
	if c.hdSigner != nil {
		_, stxn, err := c.hdSigner.SignTransaction(txn)
		if err != nil {
			return nil, fmt.Errorf("sign txn: %w", err)
		}
		return stxn, nil
	}

	_, stxn, err := crypto.SignTransaction(c.account.PrivateKey, txn)
	if err != nil {
		return nil, fmt.Errorf("sign txn: %w", err)
	}
	return stxn, nil
}

// EnsureAssetOptIn opts the wallet into an ASA if it has not already.
func (c *Client) EnsureAssetOptIn(ctx context.Context, assetID uint64) error {
	addr := c.account.Address.String()
	info, err := c.algod.AccountInformation(addr).Do(ctx)
	if err != nil {
		return fmt.Errorf("account info: %w", err)
	}
	for _, holding := range info.Assets {
		if holding.AssetId == assetID {
			return nil
		}
	}

	params, err := c.SuggestedParams(ctx)
	if err != nil {
		return err
	}
	txn, err := transaction.MakeAssetTransferTxn(
		addr, addr, 0, []byte("x402-asa-opt-in"), params, "", assetID,
	)
	if err != nil {
		return fmt.Errorf("make opt-in txn: %w", err)
	}
	stxn, err := c.SignTxn(txn)
	if err != nil {
		return err
	}
	txID, err := c.algod.SendRawTransaction(stxn).Do(ctx)
	if err != nil {
		return fmt.Errorf("send opt-in txn: %w", err)
	}
	if err := waitConfirmed(ctx, c.algod, txID); err != nil {
		return fmt.Errorf("confirm opt-in txn: %w", err)
	}
	return nil
}

// waitConfirmed waits for a txn without rapid polling (free RPC tiers rate-limit aggressively).
func waitConfirmed(ctx context.Context, client *algod.Client, txID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(18 * time.Second):
	}
	_, _, err := client.PendingTransactionInformation(txID).Do(ctx)
	return err
}

// CommitHash submits a 0-ALGO self-payment with a note carrying the reasoning hash.
func (c *Client) CommitHash(reasoningHash string) (string, error) {
	note := []byte("RationAlgo:commit:" + reasoningHash)
	return c.commitNote(note)
}

// CommitProvenance submits a 0-ALGO self-payment with an RAv1 envelope note.
func (c *Client) CommitProvenance(env *provenance.Envelope) (string, error) {
	noteStr, err := provenance.Encode(env)
	if err != nil {
		return "", fmt.Errorf("algorand: commit provenance: %w", err)
	}
	return c.commitNote([]byte(noteStr))
}

// CommitOutcome submits a 0-ALGO self-payment with an RAv1out envelope note.
func (c *Client) CommitOutcome(env *provenance.OutcomeEnvelope) (string, error) {
	noteStr, err := provenance.EncodeOutcome(env)
	if err != nil {
		return "", fmt.Errorf("algorand: commit outcome: %w", err)
	}
	return c.commitNote([]byte(noteStr))
}

func (c *Client) commitNote(note []byte) (string, error) {
	if len(note) > 1000 {
		return "", fmt.Errorf("note too long (%d bytes); max 1000", len(note))
	}

	params, err := c.algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return "", fmt.Errorf("suggested params: %w", err)
	}

	addr := c.account.Address.String()
	txn, err := transaction.MakePaymentTxn(addr, addr, 0, note, "", params)
	if err != nil {
		return "", fmt.Errorf("make payment txn: %w", err)
	}

	stxn, err := c.SignTxn(txn)
	if err != nil {
		return "", err
	}

	txID, err := c.algod.SendRawTransaction(stxn).Do(context.Background())
	if err != nil {
		return "", fmt.Errorf("send txn: %w", err)
	}

	if _, err := transaction.WaitForConfirmation(c.algod, txID, 4, context.Background()); err != nil {
		return "", fmt.Errorf("confirm txn: %w", err)
	}

	return txID, nil
}
