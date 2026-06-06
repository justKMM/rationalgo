package util

import (
	"fmt"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algonode/go-algorand-hd-wallet/XHDWalletAPI"
	"github.com/tyler-smith/go-bip39"
)

const maxUniversalWalletAccounts = 50

// WalletCredentials holds signing material for legacy or Pera Universal (BIP39) wallets.
type WalletCredentials struct {
	Address   types.Address
	LegacyKey []byte // ed25519 private key; nil for Universal Wallet
	HDPath    *XHDWalletAPI.AlgoPath
}

// ValidateWalletMnemonic checks that raw signs for expectedAddr.
func ValidateWalletMnemonic(raw, expectedAddr string) error {
	_, err := ResolveWallet(raw, expectedAddr)
	return err
}

// ResolveWallet accepts a legacy Algorand passphrase (24/25 words) or a Pera Universal
// Wallet BIP39 seed phrase (24 words, m/44'/283'/account'/0/0).
func ResolveWallet(raw, expectedAddr string) (*WalletCredentials, error) {
	phrase := cleanMnemonicPhrase(raw)
	if phrase == "" {
		return nil, fmt.Errorf("mnemonic is empty")
	}

	if creds, err := resolveLegacyMnemonic(phrase, expectedAddr); err == nil {
		return creds, nil
	}

	if !bip39.IsMnemonicValid(phrase) {
		words := strings.Fields(phrase)
		if len(words) == 24 {
			return nil, invalid24WordMnemonicError(phrase)
		}
		return nil, fmt.Errorf("mnemonic must be a 24- or 25-word Algorand passphrase or a 24-word Pera Universal Wallet seed phrase")
	}

	return resolveUniversalWallet(phrase, expectedAddr)
}

func resolveLegacyMnemonic(phrase, expectedAddr string) (*WalletCredentials, error) {
	normalized, err := NormalizeMnemonic(phrase)
	if err != nil {
		return nil, err
	}

	sk, err := mnemonic.ToPrivateKey(normalized)
	if err != nil {
		return nil, err
	}
	account, err := crypto.AccountFromPrivateKey(sk)
	if err != nil {
		return nil, err
	}
	if account.Address.String() != expectedAddr {
		return nil, fmt.Errorf(
			"mnemonic address %s does not match RATIONALGO_WALLET_ADDRESS %s",
			account.Address.String(), expectedAddr,
		)
	}

	return &WalletCredentials{
		Address:   account.Address,
		LegacyKey: account.PrivateKey,
	}, nil
}

func resolveUniversalWallet(phrase, expectedAddr string) (*WalletCredentials, error) {
	w, err := XHDWalletAPI.NewWalletFromMnemonic(phrase, "")
	if err != nil {
		return nil, fmt.Errorf("universal wallet: %w", err)
	}

	for account := uint32(0); account < maxUniversalWalletAccounts; account++ {
		ap := w.Path(account, 0, 0)
		addr, err := ap.AlgorandAddress()
		if err != nil {
			continue
		}
		if addr.String() != expectedAddr {
			continue
		}
		path := ap
		return &WalletCredentials{
			Address: addr,
			HDPath:  &path,
		}, nil
	}

	return nil, fmt.Errorf(
		"BIP39 seed phrase does not derive RATIONALGO_WALLET_ADDRESS %s — "+
			"confirm the address and passphrase are from the same Pera Universal Wallet account",
		expectedAddr,
	)
}

func cleanMnemonicPhrase(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"'`)
	words := strings.Fields(raw)
	for i := range words {
		words[i] = strings.ToLower(strings.Trim(words[i], `"'`))
	}
	return strings.Join(words, " ")
}
