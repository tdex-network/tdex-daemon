package inmemory

import (
	"context"
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/vulpemventures/go-elements/network"
)

// VaultRepositoryImpl represents an in memory storage
type VaultRepositoryImpl struct {
	store *vaultInmemoryStore
}

// NewVaultRepositoryImpl returns a new empty VaultRepositoryImpl
func NewVaultRepositoryImpl(store *vaultInmemoryStore) domain.VaultRepository {
	return &VaultRepositoryImpl{store}
}

func (r VaultRepositoryImpl) GetOrCreateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	net *network.Network,
) (*domain.Vault, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getOrCreateVault(ctx, mnemonic, passphrase, net)
}

func (r VaultRepositoryImpl) GetAllDerivedExternalAddressesInfoForAccount(
	ctx context.Context,
	accountIndex int,
) (domain.AddressesInfo, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	v := r.getVault(ctx)
	if v == nil {
		return nil, ErrVaultNotFound
	}

	return v.AllDerivedExternalAddressesInfoForAccount(accountIndex)
}

// UpdateVault updates data to the Vault passing an update function
func (r VaultRepositoryImpl) UpdateVault(
	ctx context.Context,
	updateFn func(*domain.Vault) (*domain.Vault, error),
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	v := r.getVault(ctx)
	if v == nil {
		return ErrVaultNotFound
	}

	updatedVault, err := updateFn(v)
	if err != nil {
		return err
	}

	r.store.vault = updatedVault

	return nil
}

func (r VaultRepositoryImpl) GetAccountByIndex(ctx context.Context,
	accountIndex int) (*domain.Account, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.store.vault.AccountByIndex(accountIndex)
}

func (r VaultRepositoryImpl) GetAccountByAddress(
	ctx context.Context,
	addr string,
) (*domain.Account, int, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.store.vault.AccountByAddress(addr)
}

func (r VaultRepositoryImpl) GetAllDerivedAddressesInfoForAccount(
	ctx context.Context,
	accountIndex int,
) (domain.AddressesInfo, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.store.vault.AllDerivedAddressesInfoForAccount(accountIndex)
}

func (r VaultRepositoryImpl) GetDerivationPathByScript(
	ctx context.Context,
	accountIndex int,
	scripts []string,
) (map[string]string, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getDerivationPathByScript(accountIndex, scripts)
}

func (r VaultRepositoryImpl) getOrCreateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	net *network.Network,
) (*domain.Vault, error) {
	if r.getVault(ctx) == nil {
		v, err := domain.NewVault(mnemonic, passphrase, net)
		if err != nil {
			return nil, err
		}
		r.store.vault = v
	}
	return r.store.vault, nil
}

func (r VaultRepositoryImpl) getVault(_ context.Context) *domain.Vault {
	if r.store.vault.IsZero() {
		return nil
	}
	return r.store.vault
}

func (r VaultRepositoryImpl) getDerivationPathByScript(
	accountIndex int,
	scripts []string) (map[string]string, error,
) {
	account, err := r.store.vault.AccountByIndex(accountIndex)
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
