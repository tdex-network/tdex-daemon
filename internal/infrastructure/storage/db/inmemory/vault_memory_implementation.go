package inmemory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/uow"
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

// VaultRepositoryImpl represents an in memory storage
type VaultRepositoryImpl struct {
	vault *domain.Vault

	lock *sync.RWMutex
}

// NewVaultRepositoryImpl returns a new empty VaultRepositoryImpl
func NewVaultRepositoryImpl() domain.VaultRepository {
	return &VaultRepositoryImpl{
		vault: &domain.Vault{},
		lock:  &sync.RWMutex{},
	}
}

// GetOrCreateVault returns the current Vault.
// If not yet initialized, it creates a new Vault, initialized with the
// mnemonic encrypted with the passphrase
func (r VaultRepositoryImpl) GetOrCreateVault(ctx context.Context,
	mnemonic []string, passphrase string) (*domain.Vault, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getOrCreateVault(r.storageByContext(ctx), mnemonic, passphrase)
}

// UpdateVault updates data to the Vault passing an update function
func (r VaultRepositoryImpl) UpdateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	updateFn func(*domain.Vault) (*domain.Vault, error),
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
func (r VaultRepositoryImpl) GetAccountByIndex(ctx context.Context,
	accountIndex int) (*domain.Account, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	storage := r.storageByContext(ctx)
	return storage.AccountByIndex(accountIndex)
}

// GetAccountByAddress returns the account with the given index if it exists
func (r VaultRepositoryImpl) GetAccountByAddress(ctx context.Context,
	addr string) (*domain.Account, int, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	storage := r.storageByContext(ctx)
	return storage.AccountByAddress(addr)
}

// GetAllDerivedAddressesAndBlindingKeysForAccount returns the list of all
// external and internal (change) addresses derived for the provided account
// along with the respective private blinding keys
func (r VaultRepositoryImpl) GetAllDerivedAddressesAndBlindingKeysForAccount(ctx context.Context, accountIndex int) ([]string, [][]byte, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	storage := r.storageByContext(ctx)
	return storage.AllDerivedAddressesAndBlindingKeysForAccount(accountIndex)
}

// GetDerivationPathByScript returns the derivation paths for the given account
// index and the given list of scripts. If some script of the list does not map
// to any known derivation path, an error is thrown
func (r VaultRepositoryImpl) GetDerivationPathByScript(ctx context.Context, accountIndex int, scripts []string) (map[string]string, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	storage := r.storageByContext(ctx)
	return getDerivationPathByScript(storage, accountIndex, scripts)
}

// Begin returns a new VaultRepositoryTx
func (r VaultRepositoryImpl) Begin() (uow.Tx, error) {
	tx := &VaultRepositoryTx{
		root:  r,
		vault: &domain.Vault{},
	}

	// copy the current state of the repo into the transaction
	*tx.vault = *r.vault
	return tx, nil
}

// ContextKey returns the context key shared between in-memory repositories
func (r VaultRepositoryImpl) ContextKey() interface{} {
	return uow.InMemoryContextKey
}

func (r VaultRepositoryImpl) storageByContext(ctx context.Context) (
	vault *domain.Vault) {
	vault = r.vault
	if tx, ok := ctx.Value(r).(*VaultRepositoryTx); ok {
		vault = tx.vault
	}
	return
}

func getOrCreateVault(storage *domain.Vault, mnemonic []string,
	passphrase string) (*domain.Vault, error) {
	if storage.IsZero() {
		v, err := domain.NewVault(mnemonic, passphrase)
		if err != nil {
			return nil, err
		}
		*storage = *v
	}
	return storage, nil
}

func getDerivationPathByScript(storage *domain.Vault, accountIndex int,
	scripts []string) (map[string]string, error) {
	account, err := storage.AccountByIndex(accountIndex)
	if err != nil {
		return nil, err
	}

	m := map[string]string{}
	for _, script := range scripts {
		derivationPath, ok := account.DerivationPathByScript[script]
		if !ok {
			return nil, fmt.Errorf("derivation path not found for script '%s'", script)
		}
		m[script] = derivationPath
	}

	return m, nil
}

// VaultRepositoryTx allows to make transactional read/write operation
// on the in-memory repository
type VaultRepositoryTx struct {
	root  VaultRepositoryImpl
	vault *domain.Vault
}

// Commit applies the updates made to the state of the transaction to its root
func (tx *VaultRepositoryTx) Commit() error {
	*tx.root.vault = *tx.vault
	return nil
}

// Rollback resets the state of the transaction to the state of its root
func (tx *VaultRepositoryTx) Rollback() error {
	*tx.vault = *tx.root.vault
	return nil
}
