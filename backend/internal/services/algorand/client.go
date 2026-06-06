package algorand

import (
	"context"
	"fmt"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/transaction"

	"rationalgo/internal/config"
	"rationalgo/internal/util"
	"rationalgo/pkg/provenance"
)

// Client wraps testnet Algod access and transaction signing.
type Client struct {
	algod   *algod.Client
	account crypto.Account
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

	normalized, err := util.NormalizeMnemonic(cfg.Mnemonic)
	if err != nil {
		return nil, fmt.Errorf("mnemonic: %w", err)
	}

	sk, err := mnemonic.ToPrivateKey(normalized)
	if err != nil {
		return nil, fmt.Errorf("mnemonic: %w", err)
	}
	account, err := crypto.AccountFromPrivateKey(sk)
	if err != nil {
		return nil, fmt.Errorf("account from mnemonic: %w", err)
	}

	addr := account.Address.String()
	if addr != cfg.WalletAddress {
		return nil, fmt.Errorf(
			"mnemonic address %s does not match RATIONALGO_WALLET_ADDRESS %s",
			addr, cfg.WalletAddress,
		)
	}

	return &Client{algod: algodClient, account: account}, nil
}

// Address returns the wallet address derived from configuration.
func (c *Client) Address() string {
	return c.account.Address.String()
}

// AccountInfo fetches on-chain account metadata.
func (c *Client) AccountInfo() (models.Account, error) {
	return c.algod.AccountInformation(c.account.Address.String()).Do(context.Background())
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

	_, stxn, err := crypto.SignTransaction(c.account.PrivateKey, txn)
	if err != nil {
		return "", fmt.Errorf("sign txn: %w", err)
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
