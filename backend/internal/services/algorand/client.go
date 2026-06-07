package algorand

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
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

const suggestedParamsTTL = 5 * time.Minute

// Client wraps testnet Algod access and transaction signing.
type Client struct {
	algod    *algod.Client
	account  crypto.Account
	hdSigner *XHDWalletAPI.AlgoPath

	paramsMu      sync.Mutex
	cachedParams  types.SuggestedParams
	paramsCached  bool
	paramsExpires time.Time
	cachedNetwork string

	optInMu    sync.Mutex
	optInAssets map[uint64]bool

	rpcMu       sync.Mutex
	lastRPC     time.Time
	minInterval time.Duration
}

// NewClient connects to Algorand Testnet using the buyer/agent wallet from config.
func NewClient(cfg config.Config) (*Client, error) {
	if err := cfg.ValidateForSpike(); err != nil {
		return nil, err
	}
	return newClient(cfg, cfg.WalletAddress, cfg.Mnemonic)
}

// NewSellerClient connects using the x402 seller wallet (RATIONALGO_SELLER_* env vars).
func NewSellerClient(cfg config.Config) (*Client, error) {
	if err := cfg.ValidateForSeller(); err != nil {
		return nil, err
	}
	return newClient(cfg, cfg.SellerWalletAddressEffective(), cfg.SellerMnemonicEffective())
}

func newClient(cfg config.Config, walletAddress, mnemonic string) (*Client, error) {
	algodClient, err := algod.MakeClient(cfg.AlgodURL, cfg.AlgodToken)
	if err != nil {
		return nil, fmt.Errorf("algod client: %w", err)
	}

	creds, err := util.ResolveWallet(mnemonic, walletAddress)
	if err != nil {
		return nil, fmt.Errorf("mnemonic: %w", err)
	}

	client := &Client{
		algod:       algodClient,
		account:     crypto.Account{Address: creds.Address},
		minInterval: cfg.AlgodMinInterval(),
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
func (c *Client) AccountInfo(ctx context.Context) (models.Account, error) {
	return retryRPC(ctx, c, func() (models.Account, error) {
		return c.algod.AccountInformation(c.account.Address.String()).Do(ctx)
	})
}

// SuggestedParams returns current algod suggested transaction parameters.
// Results are cached briefly to avoid RPC rate limits during multi-step hero demos.
func (c *Client) SuggestedParams(ctx context.Context) (types.SuggestedParams, error) {
	c.paramsMu.Lock()
	if c.paramsCached && time.Now().Before(c.paramsExpires) {
		params := c.cachedParams
		c.paramsMu.Unlock()
		return params, nil
	}
	c.paramsMu.Unlock()

	params, err := retryRPC(ctx, c, func() (types.SuggestedParams, error) {
		return c.algod.SuggestedParams().Do(ctx)
	})
	if err != nil {
		return types.SuggestedParams{}, fmt.Errorf("suggested params: %w", err)
	}

	c.paramsMu.Lock()
	c.cachedParams = params
	c.paramsCached = true
	c.paramsExpires = time.Now().Add(suggestedParamsTTL)
	c.cachedNetwork = "algorand:" + base64.StdEncoding.EncodeToString(params.GenesisHash[:])
	c.paramsMu.Unlock()

	return params, nil
}

// NetworkIdentifier returns the x402 CAIP-2 network id for the connected algod node.
func (c *Client) NetworkIdentifier(ctx context.Context) (string, error) {
	c.paramsMu.Lock()
	if c.paramsCached && c.cachedNetwork != "" && time.Now().Before(c.paramsExpires) {
		network := c.cachedNetwork
		c.paramsMu.Unlock()
		return network, nil
	}
	c.paramsMu.Unlock()

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
	c.optInMu.Lock()
	if c.optInAssets != nil && c.optInAssets[assetID] {
		c.optInMu.Unlock()
		return nil
	}
	c.optInMu.Unlock()

	addr := c.account.Address.String()
	info, err := retryRPC(ctx, c, func() (models.Account, error) {
		return c.algod.AccountInformation(addr).Do(ctx)
	})
	if err != nil {
		return fmt.Errorf("account info: %w", err)
	}
	for _, holding := range info.Assets {
		if holding.AssetId == assetID {
			c.markOptIn(assetID)
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
	txID, err := retryRPC(ctx, c, func() (string, error) {
		return c.algod.SendRawTransaction(stxn).Do(ctx)
	})
	if err != nil {
		return fmt.Errorf("send opt-in txn: %w", err)
	}
	// Opt-in is rare (once at startup); one confirmation poll is acceptable.
	if err := waitConfirmed(ctx, c.algod, txID); err != nil {
		return fmt.Errorf("confirm opt-in txn: %w", err)
	}
	c.markOptIn(assetID)
	return nil
}

func (c *Client) markOptIn(assetID uint64) {
	c.optInMu.Lock()
	defer c.optInMu.Unlock()
	if c.optInAssets == nil {
		c.optInAssets = make(map[uint64]bool)
	}
	c.optInAssets[assetID] = true
}

// SubmitSignedTxn broadcasts a pre-signed, msgpack-encoded transaction.
// Confirmation is skipped to keep algod RPC usage low on free-tier providers.
func (c *Client) SubmitSignedTxn(ctx context.Context, raw []byte) (string, error) {
	txID, err := retryRPC(ctx, c, func() (string, error) {
		return c.algod.SendRawTransaction(raw).Do(ctx)
	})
	if err != nil {
		return "", fmt.Errorf("send raw txn: %w", err)
	}
	return txID, nil
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

	ctx := context.Background()
	params, err := c.SuggestedParams(ctx)
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

	txID, err := retryRPC(ctx, c, func() (string, error) {
		return c.algod.SendRawTransaction(stxn).Do(ctx)
	})
	if err != nil {
		return "", fmt.Errorf("send txn: %w", err)
	}

	return txID, nil
}
