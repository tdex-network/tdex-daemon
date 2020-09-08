package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

var (
	// ErrAlreadyLocked is thrown when trying to lock an already locked wallet
	ErrAlreadyLocked = errors.New("wallet is already locked")
	// ErrAlreadyUnlocked is thrown when trying to lunock an already unlocked wallet
	ErrAlreadyUnlocked = errors.New("wallet is already unlocked")
	// ErrWalletNotExist is thrown when mnemonic is not found
	ErrWalletNotExist = errors.New("wallet does not exist")
	// ErrWalletAlreadyExist is thrown when trying to create a new mnemonic if another one already exists
	ErrWalletAlreadyExist = errors.New("wallet already initialized with mnemonic")
	// ErrMustBeLocked is thrown when trying to change the passphrase with an unlocked wallet
	ErrMustBeLocked = errors.New("wallet must be locked to perform this operation")
	// ErrMustBeUnlocked is thrown when trying to make an operation that requires the wallet to be unlocked
	ErrMustBeUnlocked = errors.New("wallet must be unlocked to perform this operation")
	// ErrAccountNotExist is thrown when account is not found
	ErrAccountNotExist = errors.New("account does not exist")
)

// InMemoryVaultRepository represents an in memory storage
type InMemoryVaultRepository struct {
	mnemonic          string
	accounts          map[uint32]vault.Account
	accountsByAddress map[string]uint32
	isLocked          bool

	lock *sync.RWMutex
}

// NewInMemoryRepository returns a new empty NewInMemoryRepository
func NewInMemoryRepository() (*InMemoryVaultRepository, error) {
	return &InMemoryVaultRepository{
		accounts:          map[uint32]vault.Account{},
		accountsByAddress: map[string]uint32{},
		lock:              &sync.RWMutex{},
	}, nil
}

// GetOrCreateMnemonic actually creates a new mnemonic for this kind of Vault's
// repository implementation
func (r *InMemoryVaultRepository) GetOrCreateMnemonic(_ context.Context) (string, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if r.isLocked {
		return "", ErrMustBeUnlocked
	}
	if len(r.mnemonic) > 0 {
		return r.mnemonic, nil
	}

	mnemonic, err := r.createMnemonic()
	if err != nil {
		return "", err
	}

	r.mnemonic = mnemonic
	r.isLocked = false
	return mnemonic, nil
}

// RestoreFromMnemonic checks that the provided mnemonic is valid and uses it as the wallet seed
func (r *InMemoryVaultRepository) RestoreFromMnemonic(_ context.Context, mnemonic string) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(r.mnemonic) > 0 {
		return ErrWalletAlreadyExist
	}

	_, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})
	if err != nil {
		return err
	}

	r.mnemonic = mnemonic
	r.isLocked = false
	return nil
}

// Lock encrypts the generated mnemonic with the provided passphrase, flushes the
// unencrypted mnemonic from storage and replace it with the encrypted one
func (r *InMemoryVaultRepository) Lock(_ context.Context, passphrase string) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(r.mnemonic) <= 0 {
		return ErrWalletNotExist
	}
	if r.isLocked {
		return ErrAlreadyLocked
	}

	encryptedMnemonic, err := wallet.Encrypt(wallet.EncryptOpts{
		PlainText:  r.mnemonic,
		Passphrase: passphrase,
	})
	if err != nil {
		return err
	}

	r.mnemonic = encryptedMnemonic
	r.isLocked = true
	return nil
}

// Unlock decrypts the encrypted mnemonic with the provided passphrase, flushes the
// encrypted mnemonic from storage and replace it with the unencrypted one
func (r *InMemoryVaultRepository) Unlock(_ context.Context, passphrase string) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(r.mnemonic) <= 0 {
		return ErrWalletNotExist
	}
	if !r.isLocked {
		return ErrAlreadyUnlocked
	}

	mnemonic, err := wallet.Decrypt(wallet.DecryptOpts{
		CypherText: r.mnemonic,
		Passphrase: passphrase,
	})
	if err != nil {
		return err
	}

	r.mnemonic = mnemonic
	r.isLocked = false
	return nil
}

// ChangePassphrase attempts to decrypt the stored encrypted mnemonic
// with the provided oldPassphrase, then it encrypts and stores the new
// encrypted one obtained using newPassphrase
func (r *InMemoryVaultRepository) ChangePassphrase(_ context.Context, oldPassphrase, newPassphrase string) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(r.mnemonic) <= 0 {
		return ErrWalletNotExist
	}
	if !r.isLocked {
		return ErrMustBeLocked
	}

	mnemonic, err := wallet.Decrypt(wallet.DecryptOpts{
		CypherText: r.mnemonic,
		Passphrase: oldPassphrase,
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

	r.mnemonic = encryptedMnemonic
	return nil
}

// GetOrCreateAccount gets an account with a given account index. If not found, a new entry is inserted
func (r *InMemoryVaultRepository) GetOrCreateAccount(_ context.Context, accountIndex uint32) (*vault.Account, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(r.mnemonic) <= 0 {
		return nil, ErrWalletNotExist
	}
	if r.isLocked {
		return nil, ErrMustBeUnlocked
	}

	return r.getOrCreateAccount(accountIndex)
}

// UpdateAccount updates data to an account identified by its index passing an update function
func (r *InMemoryVaultRepository) UpdateAccount(
	_ context.Context,
	accountIndex uint32,
	updateFn func(a *vault.Account) (*vault.Account, error),
) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(r.mnemonic) <= 0 {
		return ErrWalletNotExist
	}
	if r.isLocked {
		return ErrMustBeUnlocked
	}

	currentAccount, _ := r.getOrCreateAccount(accountIndex)
	updatedAccount, err := updateFn(currentAccount)
	if err != nil {
		return err
	}

	r.accounts[accountIndex] = *updatedAccount
	return nil
}

// GetAccountByAddress returns the Account using for the provided address
func (r *InMemoryVaultRepository) GetAccountByAddress(_ context.Context, addr string) (*vault.Account, int, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(r.mnemonic) <= 0 {
		return nil, -1, ErrWalletNotExist
	}
	if r.isLocked {
		return nil, -1, ErrMustBeUnlocked
	}

	return r.getAccountByAddress(addr)
}

// AddAccountByAddress adds a new entry address:accountIndex to the omonym map
func (r *InMemoryVaultRepository) AddAccountByAddress(_ context.Context, addr string, accountIndex uint32) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(r.mnemonic) <= 0 {
		return ErrWalletNotExist
	}
	if r.isLocked {
		return ErrMustBeUnlocked
	}

	return r.addAccountByAddress(addr, accountIndex)
}

func (r *InMemoryVaultRepository) getMnemonic(passphrase string) (string, error) {
	if len(passphrase) > 0 {
		return wallet.Decrypt(wallet.DecryptOpts{
			CypherText: r.mnemonic,
			Passphrase: passphrase,
		})
	}
	return r.mnemonic, nil
}

func (r *InMemoryVaultRepository) createMnemonic() (string, error) {
	w, err := wallet.NewWallet(wallet.NewWalletOpts{
		EntropySize:   256,
		ExtraMnemonic: false,
	})
	if err != nil {
		return "", nil
	}

	return w.SigningMnemonic()
}

func (r *InMemoryVaultRepository) getOrCreateAccount(accountIndex uint32) (*vault.Account, error) {
	if account, ok := r.accounts[accountIndex]; ok {
		return &account, nil
	}
	account := vault.NewAccount()
	r.accounts[accountIndex] = *account
	return account, nil
}

func (r *InMemoryVaultRepository) getAccountByAddress(addr string) (*vault.Account, int, error) {
	accountIndex, ok := r.accountsByAddress[addr]
	if !ok {
		return nil, -1, ErrAccountNotExist
	}
	account, ok := r.accounts[accountIndex]
	if !ok {
		return nil, -1, ErrAccountNotExist
	}
	return &account, int(accountIndex), nil
}

func (r *InMemoryVaultRepository) addAccountByAddress(addr string, accountIndex uint32) error {
	if _, ok := r.accounts[accountIndex]; !ok {
		return ErrAccountNotExist
	}
	// add the entry only if not already existing
	if _, ok := r.accountsByAddress[addr]; !ok {
		r.accountsByAddress[addr] = accountIndex
	}
	return nil
}
