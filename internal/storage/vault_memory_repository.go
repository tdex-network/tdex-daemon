package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
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
	vault *vault.Vault

	lock *sync.RWMutex
}

// NewInMemoryVaultRepository returns a new empty InMemoryVaultRepository
func NewInMemoryVaultRepository() *InMemoryVaultRepository {
	return &InMemoryVaultRepository{
		vault: &vault.Vault{},
		lock:  &sync.RWMutex{},
	}
}

// GetOrCreateVault returns the current Vault.
// If not yet initialized, it creates a new Vault, initialized with the
// mnemonic encrypted with the passphrase
func (r *InMemoryVaultRepository) GetOrCreateVault(mnemonic []string, passphrase string) (*vault.Vault, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getOrCreateVault(mnemonic, passphrase)
}

// UpdateVault updates data to the Vault passing an update function
func (r *InMemoryVaultRepository) UpdateVault(
	_ context.Context,
	mnemonic []string,
	passphrase string,
	updateFn func(*vault.Vault) (*vault.Vault, error),
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	v, err := r.getOrCreateVault(mnemonic, passphrase)
	if err != nil {
		return err
	}

	updatedVault, err := updateFn(v)
	if err != nil {
		return err
	}

	r.vault = updatedVault
	return nil
}

// GetAccountByIndex returns the account with the given index if it exists
func (r *InMemoryVaultRepository) GetAccountByIndex(_ context.Context, accountIndex int) (*vault.Account, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.vault.AccountByIndex(accountIndex)
}

// GetAccountByAddress returns the account with the given index if it exists
func (r *InMemoryVaultRepository) GetAccountByAddress(_ context.Context, addr string) (*vault.Account, int, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.vault.AccountByAddress(addr)
}

// GetAllDerivedAddressesAndBlindingKeysForAccount returns the list of all
// external and internal (change) addresses derived for the provided account
// along with the respective private blinding keys
func (r *InMemoryVaultRepository) GetAllDerivedAddressesAndBlindingKeysForAccount(_ context.Context, accountIndex int) ([]string, [][]byte, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.vault.AllDerivedAddressesAndBlindingKeysForAccount(accountIndex)
}

func (r *InMemoryVaultRepository) getOrCreateVault(mnemonic []string, passphrase string) (*vault.Vault, error) {
	if r.vault.IsZero() {
		v, err := vault.NewVault(mnemonic, passphrase)
		if err != nil {
			return nil, errors.New(
				"vault must be initialized with mnemonic and passphrase",
			)
		}
		return v, nil
	}
	return r.vault, nil
}
