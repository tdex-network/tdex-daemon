package domain

import (
	"strings"

	"github.com/btcsuite/btcutil"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/network"
)

type AccountAndKey struct {
	AccountIndex int
	BlindingKey  []byte
}

type Vault struct {
	EncryptedMnemonic      string
	PassphraseHash         []byte
	Accounts               map[int]*Account
	AccountAndKeyByAddress map[string]AccountAndKey
	Network                *network.Network
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

// IMnemonicStore defines the required methods to override the default
// storage of the plaintext mnemonic once Unlocking a Vault.
// At the moment this is achieved by storing the mnemonic in the in-memory
// pkg/config store. Soon this will be changed since this package shouldn't
// depend on config.
type IMnemonicStore interface {
	Set(mnemonic string)
	Unset()
	IsSet() bool
	Get() []string
}

// TOOD: move this outside the domain to decouple from config and use an
// equivalent alternative.
type configStore struct{}

func (c configStore) Set(mnemonic string) {
	config.Set(config.MnemonicKey, mnemonic)
}

func (c configStore) Unset() {
	config.Set(config.MnemonicKey, nil)
}

func (c configStore) IsSet() bool {
	return len(config.GetMnemonic()) > 0
}

func (c configStore) Get() []string {
	return config.GetMnemonic()
}

// IEncrypter defines the required methods to override the default encryption
// performed through pkg/wallet package.
type IEncrypter interface {
	Encrypt(mnemonic, passphrase string) (string, error)
	Decrypt(encryptedMnemonic, passphrase string) (string, error)
}

type walletEncrypter struct{}

func (w walletEncrypter) Encrypt(mnemonic, passphrase string) (string, error) {
	return wallet.Encrypt(wallet.EncryptOpts{
		PlainText:  mnemonic,
		Passphrase: passphrase,
	})
}

func (w walletEncrypter) Decrypt(encryptedMnemonic, passphrase string) (string, error) {
	return wallet.Decrypt(wallet.DecryptOpts{
		CypherText: encryptedMnemonic,
		Passphrase: passphrase,
	})
}

// MnemonicStore can be set externally by the user of the domain, assigning it
// to an instance of a IMnemonicStore implementation
var (
	MnemonicStore IMnemonicStore
	Encrypter     IEncrypter
)

func init() {
	Encrypter = walletEncrypter{}
	MnemonicStore = configStore{}
}

// NewVault encrypts the provided mnemonic with the passhrase and returns a new
// Vault initialized with the encrypted mnemonic and the hash of the passphrase.
// The Vault is locked by default since it is initialized without the mnemonic
// in plain text
func NewVault(mnemonic []string, passphrase string, net *network.Network) (*Vault, error) {
	if len(mnemonic) <= 0 || len(passphrase) <= 0 {
		return nil, ErrVaultNullMnemonicOrPassphrase
	}
	if net == nil {
		return nil, ErrVaultNullNetwork
	}

	if _, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	}); err != nil {
		return nil, err
	}

	strMnemonic := strings.Join(mnemonic, " ")
	encryptedMnemonic, err := Encrypter.Encrypt(strMnemonic, passphrase)
	if err != nil {
		return nil, err
	}

	MnemonicStore.Set(strMnemonic)
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
