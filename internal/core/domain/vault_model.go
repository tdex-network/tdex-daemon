package domain

import (
	"strings"

	"github.com/btcsuite/btcutil"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

type AccountAndKey struct {
	AccountIndex int
	BlindingKey  []byte
}

type Vault struct {
	//Mnemonic               []string
	EncryptedMnemonic      string
	PassphraseHash         []byte
	Accounts               map[int]*Account
	AccountAndKeyByAddress map[string]AccountAndKey
}

// Account defines the entity data struture for a derived account of the
// daemon's HD wallet
type Account struct {
	AccountIndex           int
	LastExternalIndex      int
	LastInternalIndex      int
	DerivationPathByScript map[string]string
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

	strMnemonic := strings.Join(mnemonic, " ")
	encryptedMnemonic, err := wallet.Encrypt(wallet.EncryptOpts{
		PlainText:  strMnemonic,
		Passphrase: passphrase,
	})
	if err != nil {
		return nil, err
	}

	config.Set(config.MnemonicKey, strMnemonic)
	return &Vault{
		EncryptedMnemonic:      encryptedMnemonic,
		PassphraseHash:         btcutil.Hash160([]byte(passphrase)),
		Accounts:               map[int]*Account{},
		AccountAndKeyByAddress: map[string]AccountAndKey{},
	}, nil
}

// NewAccount returns an empty Account instance
func NewAccount(positiveAccountIndex int) (*Account, error) {
	if err := validateAccountIndex(positiveAccountIndex); err != nil {
		return nil, err
	}

	return &Account{
		AccountIndex:           positiveAccountIndex,
		DerivationPathByScript: map[string]string{},
		LastExternalIndex:      0,
		LastInternalIndex:      0,
	}, nil
}
