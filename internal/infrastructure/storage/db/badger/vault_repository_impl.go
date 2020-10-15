package dbbadger

import (
	"context"
	"github.com/dgraph-io/badger"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/uow"
	"github.com/timshannon/badgerhold"
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
	tx := ctx.Value("tx").(*badger.Txn)

	return v.getOrCreateVault(tx, mnemonic, passphrase)
}

func (v vaultRepositoryImpl) UpdateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	updateFn func(v *domain.Vault) (*domain.Vault, error),
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	var err error
	var vault domain.Vault
	err = v.db.Store.TxFind(
		tx,
		&vault,
		badgerhold.Where(badgerhold.Key).Eq(mnemonic),
	)
	if err != nil {
		return err
	}

	var newVault *domain.Vault
	if vault.IsZero() {
		newVault, err = domain.NewVault(mnemonic, passphrase)
		if err != nil {
			return err
		}

		err = v.db.Store.TxInsert(
			tx,
			newVault.Mnemonic,
			newVault,
		)
		if err != nil {
			return err
		}
	}

	updatedVault, err := updateFn(v)
	if err != nil {
		return err
	}

	*storage = *updatedVault
	return nil
}

func (v vaultRepositoryImpl) GetAccountByIndex(
	ctx context.Context,
	accountIndex int,
) (*domain.Account, error) {
	panic("implement me")
}

func (v vaultRepositoryImpl) GetAccountByAddress(
	ctx context.Context,
	addr string,
) (*domain.Account, int, error) {
	panic("implement me")
}

func (v vaultRepositoryImpl) GetAllDerivedAddressesAndBlindingKeysForAccount(
	ctx context.Context,
	accountIndex int,
) ([]string, [][]byte, error) {
	panic("implement me")
}

func (v vaultRepositoryImpl) GetDerivationPathByScript(
	ctx context.Context,
	accountIndex int,
	scripts []string,
) (map[string]string, error) {
	panic("implement me")
}

func (v vaultRepositoryImpl) Begin() (uow.Tx, error) {
	panic("implement me")
}

func (v vaultRepositoryImpl) ContextKey() interface{} {
	panic("implement me")
}

func (v vaultRepositoryImpl) getOrCreateVault(
	tx *badger.Txn,
	mnemonic []string,
	passphrase string,
) (*domain.Vault, error) {
	vault, err := v.getVault(tx, mnemonic)
	if err != nil {
		return nil, err
	}

	if vault == nil {
		vlt, err := domain.NewVault(mnemonic, passphrase)
		if err != nil {
			return nil, err
		}

		err = v.insertVault(tx, *vlt)
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
	if err := v.db.Store.TxInsert(
		tx,
		vault.Mnemonic,
		&vault,
	); err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}

	return nil
}

func (v vaultRepositoryImpl) getVault(
	tx *badger.Txn,
	mnemonic []string,
) (*domain.Vault, error) {
	var vault domain.Vault
	if err := v.db.Store.TxGet(
		tx,
		mnemonic,
		&vault,
	); err != nil {
		if err != badgerhold.ErrKeyExists {
			return nil, err
		}
	}

	return &vault, nil
}
