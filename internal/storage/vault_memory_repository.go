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

// NewInMemoryRepository returns a new empty NewInMemoryRepository
func NewInMemoryRepository() (*InMemoryVaultRepository, error) {
	return &InMemoryVaultRepository{
		vault: vault.NewVault(),
		lock:  &sync.RWMutex{},
	}, nil
}

// CreateOrRestoreVault actually creates a new seed for the Vault or uses the one provided
func (r *InMemoryVaultRepository) CreateOrRestoreVault(mnemonic string) (*vault.Vault, string, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.createOrRestoreVault(mnemonic)
}

// UpdateVault updates data to the Vault passing an update function
func (r *InMemoryVaultRepository) UpdateVault(
	_ context.Context,
	updateFn func(*vault.Vault) (*vault.Vault, error),
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	updatedVault, err := updateFn(r.vault)
	if err != nil {
		return err
	}

	r.vault = updatedVault
	return nil
}

func (r *InMemoryVaultRepository) createOrRestoreVault(mnemonic string) (*vault.Vault, string, error) {
	if len(mnemonic) > 0 {
		err := r.vault.RestoreFromMnemonic(mnemonic)
		if err != nil {
			return nil, "", err
		}
		return r.vault, mnemonic, nil
	}

	mnemonic, err := r.vault.GenSeed()
	if err != nil {
		return nil, "", err
	}
	return r.vault, mnemonic, nil
}
