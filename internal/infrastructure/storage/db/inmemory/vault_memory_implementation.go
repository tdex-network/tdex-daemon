package inmemory

import (
	"context"
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

// VaultRepositoryImpl represents an in memory storage
type VaultRepositoryImpl struct {
	db *DbManager
}

// NewVaultRepositoryImpl returns a new empty VaultRepositoryImpl
func NewVaultRepositoryImpl(db *DbManager) domain.VaultRepository {
	return &VaultRepositoryImpl{
		db: db,
	}
}

func (r VaultRepositoryImpl) GetAllDerivedExternalAddressesForAccount(ctx context.Context, accountIndex int) ([]string, error) {
	return nil, nil
}

// GetOrCreateVault returns the current Vault.
// If not yet initialized, it creates a new Vault, initialized with the
// mnemonic encrypted with the passphrase
func (r VaultRepositoryImpl) GetOrCreateVault(ctx context.Context,
	mnemonic []string, passphrase string) (*domain.Vault, error) {
	r.db.vaultStore.locker.Lock()
	defer r.db.vaultStore.locker.Unlock()

	return r.getOrCreateVault(mnemonic, passphrase)
}

// UpdateVault updates data to the Vault passing an update function
func (r VaultRepositoryImpl) UpdateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	updateFn func(*domain.Vault) (*domain.Vault, error),
) error {
	r.db.vaultStore.locker.Lock()
	defer r.db.vaultStore.locker.Unlock()

	v, err := r.getOrCreateVault(mnemonic, passphrase)
	if err != nil {
		return err
	}

	updatedVault, err := updateFn(v)
	if err != nil {
		return err
	}

	r.db.vaultStore.vault = updatedVault

	return nil
}

// GetAccountByIndex returns the account with the given index if it exists
func (r VaultRepositoryImpl) GetAccountByIndex(ctx context.Context,
	accountIndex int) (*domain.Account, error) {
	r.db.vaultStore.locker.Lock()
	defer r.db.vaultStore.locker.Unlock()

	return r.db.vaultStore.vault.AccountByIndex(accountIndex)
}

// GetAccountByAddress returns the account with the given index if it exists
func (r VaultRepositoryImpl) GetAccountByAddress(ctx context.Context,
	addr string) (*domain.Account, int, error) {
	r.db.vaultStore.locker.Lock()
	defer r.db.vaultStore.locker.Unlock()

	return r.db.vaultStore.vault.AccountByAddress(addr)
}

// GetAllDerivedAddressesAndBlindingKeysForAccount returns the list of all
// external and internal (change) addresses derived for the provided account
// along with the respective private blinding keys
func (r VaultRepositoryImpl) GetAllDerivedAddressesAndBlindingKeysForAccount(ctx context.Context, accountIndex int) ([]string, [][]byte, error) {
	r.db.vaultStore.locker.Lock()
	defer r.db.vaultStore.locker.Unlock()

	return r.db.vaultStore.vault.AllDerivedAddressesAndBlindingKeysForAccount(accountIndex)
}

// GetDerivationPathByScript returns the derivation paths for the given account
// index and the given list of scripts. If some script of the list does not map
// to any known derivation path, an error is thrown
func (r VaultRepositoryImpl) GetDerivationPathByScript(ctx context.Context, accountIndex int, scripts []string) (map[string]string, error) {
	r.db.vaultStore.locker.Lock()
	defer r.db.vaultStore.locker.Unlock()

	return r.getDerivationPathByScript(accountIndex, scripts)
}

func (r VaultRepositoryImpl) getOrCreateVault(mnemonic []string,
	passphrase string) (*domain.Vault, error) {
	if r.db.vaultStore.vault.IsZero() {
		v, err := domain.NewVault(mnemonic, passphrase)
		if err != nil {
			return nil, err
		}
		r.db.vaultStore.vault = v
	}
	return r.db.vaultStore.vault, nil
}

func (r VaultRepositoryImpl) getDerivationPathByScript(accountIndex int,
	scripts []string) (map[string]string, error) {
	account, err := r.db.vaultStore.vault.AccountByIndex(accountIndex)
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
