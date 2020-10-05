package domain

import (
	"strings"

	"github.com/btcsuite/btcutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

type accountAndKey struct {
	accountIndex int
	blindingKey  []byte
}

type Vault struct {
	mnemonic               []string
	encryptedMnemonic      string
	passphraseHash         []byte
	accounts               map[int]*Account
	accountAndKeyByAddress map[string]accountAndKey
}

// Account defines the entity data struture for a derived account of the
// daemon's HD wallet
type Account struct {
	accountIndex           int
	lastExternalIndex      int
	lastInternalIndex      int
	derivationPathByScript map[string]string
}

type AddressInfo struct {
	AccountIndex   int
	Address        string
	BlindingKey    []byte
	DerivationPath string
}

// NewVault encrypts the provided mnemonic with the passhrase and returns a new
// Vault initialized with the encrypted mnemonic and the hash of the passphrase.
// The Vault is locked by default since it is initialized without the mnemonic
// in plain text
func NewVault(mnemonic []string, passphrase string) (*Vault, error) {
	if len(mnemonic) <= 0 || len(passphrase) <= 0 {
		return nil, ErrNullMnemonicOrPassphrase
	}
	if _, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	}); err != nil {
		return nil, err
	}

	encryptedMnemonic, err := wallet.Encrypt(wallet.EncryptOpts{
		PlainText:  strings.Join(mnemonic, " "),
		Passphrase: passphrase,
	})
	if err != nil {
		return nil, err
	}

	return &Vault{
		mnemonic:               mnemonic,
		encryptedMnemonic:      encryptedMnemonic,
		passphraseHash:         btcutil.Hash160([]byte(passphrase)),
		accounts:               map[int]*Account{},
		accountAndKeyByAddress: map[string]accountAndKey{},
	}, nil
}

// NewAccount returns an empty Account instance
func NewAccount(positiveAccountIndex int) (*Account, error) {
	if err := validateAccountIndex(positiveAccountIndex); err != nil {
		return nil, err
	}

	return &Account{
		accountIndex:           positiveAccountIndex,
		derivationPathByScript: map[string]string{},
		lastExternalIndex:      0,
		lastInternalIndex:      0,
	}, nil
}
