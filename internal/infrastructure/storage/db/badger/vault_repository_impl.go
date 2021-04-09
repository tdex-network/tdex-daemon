package dbbadger

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
	"github.com/vulpemventures/go-elements/network"
)

const (
	//since there can be only 1 vault in database,
	//key is hardcoded for easier retrival
	vaultKey = "vault"
)

type vaultRepositoryImpl struct {
	store *badgerhold.Store
}

func NewVaultRepositoryImpl(store *badgerhold.Store) domain.VaultRepository {
	return vaultRepositoryImpl{store}
}

func (v vaultRepositoryImpl) GetOrCreateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	net *network.Network,
) (*domain.Vault, error) {
	return v.getOrCreateVault(ctx, mnemonic, passphrase, net)
}

func (v vaultRepositoryImpl) UpdateVault(
	ctx context.Context,
	updateFn func(v *domain.Vault) (*domain.Vault, error),
) error {
	vault, err := v.getVault(ctx)
	if err != nil {
		return err
	}
	if vault == nil {
		return ErrVaultNotFound
	}

	updatedVault, err := updateFn(vault)
	if err != nil {
		return err
	}

	return v.updateVault(ctx, *updatedVault)
}

func (v vaultRepositoryImpl) GetAccountByIndex(
	ctx context.Context,
	accountIndex int,
) (*domain.Account, error) {
	vault, err := v.getVault(ctx)
	if err != nil {
		return nil, err
	}
	if vault == nil {
		return nil, ErrVaultNotFound
	}

	account, err := vault.AccountByIndex(accountIndex)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (v vaultRepositoryImpl) GetAccountByAddress(
	ctx context.Context,
	addr string,
) (*domain.Account, int, error) {
	vault, err := v.getVault(ctx)
	if err != nil {
		return nil, -1, err
	}
	if vault == nil {
		return nil, -1, ErrVaultNotFound
	}

	account, accountIndex, err := vault.AccountByAddress(addr)
	if err != nil {
		return nil, -1, err
	}

	return account, accountIndex, nil
}

func (v vaultRepositoryImpl) GetAllDerivedAddressesInfoForAccount(
	ctx context.Context,
	accountIndex int,
) (domain.AddressesInfo, error) {
	vault, err := v.getVault(ctx)
	if err != nil {
		return nil, err
	}
	if vault == nil {
		return nil, ErrVaultNotFound
	}

	return vault.AllDerivedAddressesInfoForAccount(accountIndex)
}

func (v vaultRepositoryImpl) GetAllDerivedExternalAddressesInfoForAccount(
	ctx context.Context,
	accountIndex int,
) (domain.AddressesInfo, error) {
	vault, err := v.getVault(ctx)
	if err != nil {
		return nil, err
	}
	if vault == nil {
		return nil, ErrVaultNotFound
	}

	return vault.AllDerivedExternalAddressesInfoForAccount(accountIndex)
}

func (v vaultRepositoryImpl) GetDerivationPathByScript(
	ctx context.Context,
	accountIndex int,
	scripts []string,
) (map[string]string, error) {
	vault, err := v.getVault(ctx)
	if err != nil {
		return nil, err
	}
	if vault == nil {
		return nil, ErrVaultNotFound
	}

	return v.getDerivationPathByScript(vault, accountIndex, scripts)
}

func (v vaultRepositoryImpl) getDerivationPathByScript(
	vault *domain.Vault,
	accountIndex int,
	scripts []string,
) (map[string]string, error) {
	account, err := vault.AccountByIndex(accountIndex)
	if err != nil {
		return nil, err
	}

	m := map[string]string{}
	for _, script := range scripts {
		derivationPath, ok := account.DerivationPathByScript[script]
		if !ok {
			return nil, fmt.Errorf(
				"derivation path not found for script '%s'",
				script,
			)
		}
		m[script] = derivationPath
	}

	return m, nil
}

func (v vaultRepositoryImpl) getOrCreateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	net *network.Network,
) (*domain.Vault, error) {
	vault, err := v.getVault(ctx)
	if err != nil {
		return nil, err
	}

	if vault == nil {
		vault, err = domain.NewVault(mnemonic, passphrase, net)
		if err != nil {
			return nil, err
		}

		err = v.insertVault(ctx, *vault)
		if err != nil {
			return nil, err
		}
	}

	return vault, nil
}

func (v vaultRepositoryImpl) insertVault(
	ctx context.Context,
	vault domain.Vault,
) error {
	var err error

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = v.store.TxInsert(tx, vaultKey, &vault)
	} else {
		err = v.store.Insert(vaultKey, &vault)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}

	return nil
}

func (v vaultRepositoryImpl) getVault(ctx context.Context) (*domain.Vault, error) {
	var err error
	var vault domain.Vault

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = v.store.TxGet(tx, vaultKey, &vault)
	} else {
		err = v.store.Get(vaultKey, &vault)
	}

	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &vault, nil
}

func (v vaultRepositoryImpl) updateVault(
	ctx context.Context,
	vault domain.Vault,
) error {
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		return v.store.TxUpdate(tx, vaultKey, vault)
	}
	return v.store.Update(vaultKey, vault)
}
