package dbbadger

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold"
)

const (
	//since there can be only 1 vault in database,
	//key is hardcoded for easier retrival
	vaultKey = "vault"
)

type vaultRepositoryImpl struct {
	db *DbManager
}

func NewVaultRepositoryImpl(db *DbManager) domain.VaultRepository {
	return vaultRepositoryImpl{
		db: db,
	}
}

func (v vaultRepositoryImpl) GetOrCreateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
) (*domain.Vault, error) {
	var tx *badger.Txn
	if ctx.Value("tx") != nil {
		tx = ctx.Value("tx").(*badger.Txn)
	}

	return v.getOrCreateVault(tx, mnemonic, passphrase)
}

func (v vaultRepositoryImpl) UpdateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	updateFn func(v *domain.Vault) (*domain.Vault, error),
) error {
	var tx *badger.Txn
	if ctx.Value("tx") != nil {
		tx = ctx.Value("tx").(*badger.Txn)
	}

	vault, err := v.getOrCreateVault(tx, mnemonic, passphrase)
	if err != nil {
		return err
	}

	updatedVault, err := updateFn(vault)
	if err != nil {
		return err
	}

	return v.updateVault(tx, *updatedVault)
}

func (v vaultRepositoryImpl) GetAccountByIndex(
	ctx context.Context,
	accountIndex int,
) (*domain.Account, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	vault, err := v.getVault(tx)
	if err != nil {
		return nil, err
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
	tx := ctx.Value("tx").(*badger.Txn)

	vault, err := v.getVault(tx)
	if err != nil {
		return nil, 0, err
	}

	account, accountIndex, err := vault.AccountByAddress(addr)
	if err != nil {
		return nil, 0, err
	}

	return account, accountIndex, nil
}

func (v vaultRepositoryImpl) GetAllDerivedAddressesAndBlindingKeysForAccount(
	ctx context.Context,
	accountIndex int,
) ([]string, [][]byte, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	vault, err := v.getVault(tx)
	if err != nil {
		return nil, nil, err
	}

	return vault.AllDerivedAddressesAndBlindingKeysForAccount(accountIndex)
}

func (v vaultRepositoryImpl) GetDerivationPathByScript(
	ctx context.Context,
	accountIndex int,
	scripts []string,
) (map[string]string, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	vault, err := v.getVault(tx)
	if err != nil {
		return nil, err
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
	tx *badger.Txn,
	mnemonic []string,
	passphrase string,
) (*domain.Vault, error) {
	vault, err := v.getVault(tx)
	if err != nil {
		return nil, err
	}

	if vault == nil {
		vault, err = domain.NewVault(mnemonic, passphrase)
		if err != nil {
			return nil, err
		}

		err = v.insertVault(tx, *vault)
		if err != nil {
			return nil, err
		}
	}

	return vault, nil
}

func (v vaultRepositoryImpl) insertVault(
	tx *badger.Txn,
	vault domain.Vault,
) error {
	var err error
	if tx != nil {
		err = v.db.Store.TxInsert(
			tx,
			vaultKey,
			&vault,
		)
	} else {
		err = v.db.Store.Insert(
			vaultKey,
			&vault,
		)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}

	return nil
}

func (v vaultRepositoryImpl) getVault(
	tx *badger.Txn,
) (*domain.Vault, error) {
	var err error
	var vault domain.Vault
	if tx != nil {
		err = v.db.Store.TxGet(
			tx,
			vaultKey,
			&vault,
		)
	} else {
		err = v.db.Store.Get(
			vaultKey,
			&vault,
		)
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
	tx *badger.Txn,
	vault domain.Vault,
) error {
	var err error
	if tx != nil {
		err = v.db.Store.TxUpdate(
			tx,
			vaultKey,
			vault,
		)
	} else {
		err = v.db.Store.Update(
			vaultKey,
			vault,
		)
	}
	return err
}
