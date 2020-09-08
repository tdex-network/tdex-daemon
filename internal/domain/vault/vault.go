package vault

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcutil"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/constant"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

const (
	// FeeAccountPath ...
	FeeAccountPath = iota
	// WalletAccountPath ...
	WalletAccountPath
	// FirstUnusedAccountPath ...
	FirstUnusedAccountPath
	// SecondUnusedAccountPath ...
	SecondUnusedAccountPath
	// FirstMarketAccountPath ...
	FirstMarketAccountPath
)

var (
	// ErrMustBeLocked is thrown when trying to change the passphrase with an unlocked wallet
	ErrMustBeLocked = errors.New("wallet must be locked to perform this operation")
	// ErrMustBeUnlocked is thrown when trying to make an operation that requires the wallet to be unlocked
	ErrMustBeUnlocked = errors.New("wallet must be unlocked to perform this operation")
	// ErrInvalidPassphrase ...
	ErrInvalidPassphrase = errors.New("passphrase is not valid")
	// ErrVaultAlreadyInitialized ...
	ErrVaultAlreadyInitialized = errors.New("vault is already initialized")
)

type Vault struct {
	mnemonic          string
	encryptedMnemonic string
	passphraseHash    []byte
	accounts          map[int]*Account
	accountsByAddress map[string]int
}

// NewVault returns a new empty vault
func NewVault() *Vault {
	return &Vault{
		accounts:          map[int]*Account{},
		accountsByAddress: map[string]int{},
	}
}

// GenSeed generates a new mnemonic for the vault
func (v *Vault) GenSeed() (string, error) {
	if !v.IsZero() {
		return "", ErrVaultAlreadyInitialized
	}
	w, err := wallet.NewWallet(wallet.NewWalletOpts{EntropySize: 256})
	if err != nil {
		return "", err
	}
	mnemonic, _ := w.SigningMnemonic()
	v.mnemonic = mnemonic
	return mnemonic, nil
}

// RestoreFromMnemonic validates the provided mnemonic and sets it as the Vault's mnemonic
func (v *Vault) RestoreFromMnemonic(mnemonic string) error {
	if !v.IsZero() {
		return ErrVaultAlreadyInitialized
	}
	_, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})
	if err != nil {
		return err
	}
	v.mnemonic = mnemonic
	return nil
}

// Lock encrypts the mnemonic with the provided passphrase
func (v *Vault) Lock(passphrase string) error {
	// check if the passphrase is correct in case this is not the first time
	// Vault is being locked
	if v.isPassphraseSet() && !v.isValidPassphrase(passphrase) {
		return ErrInvalidPassphrase
	}
	if v.IsLocked() {
		return nil
	}

	mnemonic, err := wallet.Encrypt(wallet.EncryptOpts{
		PlainText:  v.mnemonic,
		Passphrase: passphrase,
	})
	if err != nil {
		return err
	}

	// save the hash of the passphrase if it's the first time Vault is locked
	if !v.isPassphraseSet() {
		v.passphraseHash = btcutil.Hash160([]byte(passphrase))
		v.encryptedMnemonic = mnemonic
	}
	// flush mnemonic in plain text
	v.mnemonic = ""
	return nil
}

// Unlock attempts to decrypt the mnemonic with the provided passphrase
func (v *Vault) Unlock(passphrase string) error {
	if !v.IsLocked() {
		return nil
	}

	mnemonic, err := wallet.Decrypt(wallet.DecryptOpts{
		CypherText: v.encryptedMnemonic,
		Passphrase: passphrase,
	})
	if err != nil {
		return err
	}

	v.mnemonic = mnemonic
	return nil
}

// ChangePassphrase attempts to unlock the
func (v *Vault) ChangePassphrase(currentPassphrase, newPassphrase string) error {
	if !v.isPassphraseSet() {
		return v.Lock(newPassphrase)
	}

	if !v.isValidPassphrase(currentPassphrase) {
		return ErrInvalidPassphrase
	}
	if !v.IsLocked() {
		return ErrMustBeLocked
	}

	mnemonic, err := wallet.Decrypt(wallet.DecryptOpts{
		CypherText: v.encryptedMnemonic,
		Passphrase: currentPassphrase,
	})
	if err != nil {
		return err
	}

	encryptedMnemonic, err := wallet.Encrypt(wallet.EncryptOpts{
		PlainText:  mnemonic,
		Passphrase: newPassphrase,
	})
	if err != nil {
		return err
	}

	v.encryptedMnemonic = encryptedMnemonic
	v.passphraseHash = btcutil.Hash160([]byte(newPassphrase))
	return nil
}

// IsLocked returns whether the Vault is locked
func (v *Vault) IsLocked() bool {
	return len(v.mnemonic) == 0 && len(v.encryptedMnemonic) > 0
}

// IsZero returns whether the Vault is initialized without holding any data
func (v *Vault) IsZero() bool {
	return len(v.mnemonic) <= 0 &&
		len(v.encryptedMnemonic) <= 0 &&
		len(v.passphraseHash) <= 0 &&
		len(v.accounts) <= 0 &&
		len(v.accountsByAddress) <= 0
}

// DeriveNextExternalAddressForAccount returns the next unused address for the
// provided account identified by its index
func (v *Vault) DeriveNextExternalAddressForAccount(accountIndex int) (string, string, error) {
	if v.IsLocked() {
		return "", "", ErrMustBeUnlocked
	}

	return v.deriveNextAddressForAccount(accountIndex, constant.ExternalChain)
}

// DeriveNextInternalAddressForAccount returns the next unused change address for the
// provided account identified by its index
func (v *Vault) DeriveNextInternalAddressForAccount(accountIndex int) (string, string, error) {
	if v.IsLocked() {
		return "", "", ErrMustBeUnlocked
	}

	return v.deriveNextAddressForAccount(accountIndex, constant.InternalChain)
}

// AccountByIndex returns the account with the given index
func (v *Vault) AccountByIndex(accountIndex int) (*Account, error) {
	account, ok := v.accounts[accountIndex]
	if !ok {
		return nil, fmt.Errorf("account not found with index %d", accountIndex)
	}
	return account, nil
}

// AccountByAddress returns the account to which the provided address belongs
func (v *Vault) AccountByAddress(addr string) (*Account, int, error) {
	accountIndex, ok := v.accountsByAddress[addr]
	if !ok {
		return nil, -1, fmt.Errorf("account not found for address '%s", addr)
	}
	account, err := v.AccountByIndex(accountIndex)
	if err != nil {
		return nil, -1, err
	}
	return account, int(accountIndex), nil
}

func (v *Vault) isValidPassphrase(passphrase string) bool {
	return bytes.Equal(v.passphraseHash, btcutil.Hash160([]byte(passphrase)))
}

func (v *Vault) isPassphraseSet() bool {
	return len(v.passphraseHash) > 0
}

func (v *Vault) deriveNextAddressForAccount(accountIndex, chainIndex int) (string, string, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: v.mnemonic,
	})
	if err != nil {
		return "", "", err
	}

	account, ok := v.accounts[accountIndex]
	if !ok {
		account, err = NewAccount(accountIndex)
		if err != nil {
			return "", "", err
		}
		v.accounts[accountIndex] = account
	}

	derivationPath := fmt.Sprintf(
		"%d'/%d/%d",
		account.Index(), chainIndex, account.LastExternalIndex(),
	)
	addr, script, err := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
		DerivationPath: derivationPath,
		Network:        config.GetNetwork(),
	})
	if err != nil {
		return "", "", err
	}
	account.addDerivationPath(hex.EncodeToString(script), derivationPath)
	account.nextExternalIndex()
	v.accountsByAddress[addr] = account.Index()

	return addr, hex.EncodeToString(script), err
}
