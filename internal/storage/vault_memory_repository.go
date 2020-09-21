package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/internal/storageutil/uow"
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
func NewInMemoryVaultRepository() InMemoryVaultRepository {
	return InMemoryVaultRepository{
		vault: &vault.Vault{},
		lock:  &sync.RWMutex{},
	}
}

// GetOrCreateVault returns the current Vault.
// If not yet initialized, it creates a new Vault, initialized with the
// mnemonic encrypted with the passphrase
func (r InMemoryVaultRepository) GetOrCreateVault(ctx context.Context, mnemonic []string, passphrase string) (*vault.Vault, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getOrCreateVault(r.storageByContext(ctx), mnemonic, passphrase)
}

// UpdateVault updates data to the Vault passing an update function
func (r InMemoryVaultRepository) UpdateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	updateFn func(*vault.Vault) (*vault.Vault, error),
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	storage := r.storageByContext(ctx)

	v, err := getOrCreateVault(storage, mnemonic, passphrase)
	if err != nil {
		return err
	}

	updatedVault, err := updateFn(v)
	if err != nil {
		return err
	}

	*storage = *updatedVault
	return nil
}

// GetAccountByIndex returns the account with the given index if it exists
func (r InMemoryVaultRepository) GetAccountByIndex(ctx context.Context, accountIndex int) (*vault.Account, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	storage := r.storageByContext(ctx)
	return storage.AccountByIndex(accountIndex)
}

// GetAccountByAddress returns the account with the given index if it exists
func (r InMemoryVaultRepository) GetAccountByAddress(ctx context.Context, addr string) (*vault.Account, int, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	storage := r.storageByContext(ctx)
	return storage.AccountByAddress(addr)
}

// GetAllDerivedAddressesAndBlindingKeysForAccount returns the list of all
// external and internal (change) addresses derived for the provided account
// along with the respective private blinding keys
func (r InMemoryVaultRepository) GetAllDerivedAddressesAndBlindingKeysForAccount(ctx context.Context, accountIndex int) ([]string, [][]byte, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	storage := r.storageByContext(ctx)
	return storage.AllDerivedAddressesAndBlindingKeysForAccount(accountIndex)
}

// Begin returns a new InMemoryVaultRepositoryTx
func (r InMemoryVaultRepository) Begin() (uow.Tx, error) {
	tx := &InMemoryVaultRepositoryTx{
		root:  r,
		vault: &vault.Vault{},
	}

	// copy the current state of the repo into the transaction
	*tx.vault = *r.vault
	return tx, nil
}

// ContextKey returns the context key shared between in-memory repositories
func (r InMemoryVaultRepository) ContextKey() interface{} {
	return uow.InMemoryContextKey
}

func (r InMemoryVaultRepository) storageByContext(ctx context.Context) (vault *vault.Vault) {
	vault = r.vault
	if tx, ok := ctx.Value(r).(*InMemoryVaultRepositoryTx); ok {
		vault = tx.vault
	}
	return
}

func getOrCreateVault(storage *vault.Vault, mnemonic []string, passphrase string) (*vault.Vault, error) {
	if storage.IsZero() {
		v, err := vault.NewVault(mnemonic, passphrase)
		if err != nil {
			return nil, errors.New(
				"vault must be initialized with mnemonic and passphrase",
			)
		}
		*storage = *v
	}
	return storage, nil
}

// InMemoryVaultRepositoryTx allows to make transactional read/write operation
// on the in-memory repository
type InMemoryVaultRepositoryTx struct {
	root  InMemoryVaultRepository
	vault *vault.Vault
}

// Commit applies the updates made to the state of the transaction to its root
func (tx *InMemoryVaultRepositoryTx) Commit() error {
	*tx.root.vault = *tx.vault
	return nil
}

// Rollback resets the state of the transaction to the state of its root
func (tx *InMemoryVaultRepositoryTx) Rollback() error {
	*tx.vault = *tx.root.vault
	return nil
}
