package dbbadger

import (
	"context"
	"errors"
	"github.com/dgraph-io/badger"
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold"
)

const (
	UnspentBadgerholdKeyPrefix = "bh_Unspent"
)

//badgerhold internal implementation adds prefix to the key
var unspentTablePrefixKey = []byte(UnspentBadgerholdKeyPrefix)

type unspentRepositoryImpl struct {
	db *DbManager
}

func NewUnspentRepositoryImpl(db *DbManager) domain.UnspentRepository {
	return unspentRepositoryImpl{
		db: db,
	}
}

func (u unspentRepositoryImpl) AddUnspents(
	ctx context.Context,
	unspents []domain.Unspent,
) error {
	tx := ctx.Value("tx").(*badger.Txn)
	for _, v := range unspents {
		if err := u.db.Store.TxInsert(
			tx,
			v.Key(),
			&v,
		); err != nil {
			if err != badgerhold.ErrKeyExists {
				return err
			}
		}
	}
	return nil
}

func (u unspentRepositoryImpl) GetAllUnspents(
	ctx context.Context,
) []domain.Unspent {
	tx := ctx.Value("tx").(*badger.Txn)
	unspents := make([]domain.Unspent, 0)

	iter := badger.DefaultIteratorOptions
	iter.PrefetchValues = false
	it := tx.NewIterator(iter)
	defer it.Close()

	for it.Seek(unspentTablePrefixKey); it.ValidForPrefix(unspentTablePrefixKey); it.Next() {
		item := it.Item()
		data, _ := item.ValueCopy(nil)
		var unspent domain.Unspent
		err := JsonDecode(data, &unspent)
		if err == nil {
			unspents = append(unspents, unspent)
		}

	}

	return unspents
}

func (u unspentRepositoryImpl) GetBalance(
	ctx context.Context,
	address string,
	assetHash string,
) (uint64, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var unspents []domain.Unspent

	query := badgerhold.Where("Address").Eq(address).
		And("AssetHash").Eq(assetHash).
		And("Spent").Eq(false).
		And("Confirmed").Eq(true)

	err := u.db.Store.TxFind(
		tx,
		&unspents,
		query,
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return 0, err
		}
	}

	var balance uint64
	for _, v := range unspents {
		balance += v.Value
	}

	return balance, nil
}

func (u unspentRepositoryImpl) GetAvailableUnspents(
	ctx context.Context,
) ([]domain.Unspent, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var unspents []domain.Unspent

	query := badgerhold.Where("Spent").Eq(false).
		And("Locked").Eq(false).
		And("Confirmed").Eq(true)

	err := u.db.Store.TxFind(
		tx,
		&unspents,
		query,
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return nil, err
		}
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var unspents []domain.Unspent

	iface := make([]interface{}, len(addresses))
	for i := range addresses {
		iface[i] = addresses[i]
	}

	query := badgerhold.Where("Address").In(iface...)

	err := u.db.Store.TxFind(
		tx,
		&unspents,
		query,
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return nil, err
		}
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetAvailableUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var unspents []domain.Unspent

	iface := make([]interface{}, len(addresses))
	for i := range addresses {
		iface[i] = addresses[i]
	}

	query := badgerhold.Where("Spent").Eq(false).
		And("Locked").Eq(false).
		And("Confirmed").Eq(true).
		And("Address").In(iface...)

	err := u.db.Store.TxFind(
		tx,
		&unspents,
		query,
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return nil, err
		}
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetUnlockedBalance(
	ctx context.Context,
	address string,
	assetHash string,
) (uint64, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var unspents []domain.Unspent

	query := badgerhold.Where("Address").Eq(address).
		And("AssetHash").Eq(assetHash).
		And("Spent").Eq(false).
		And("Confirmed").Eq(true).
		And("Locked").Eq(false)

	err := u.db.Store.TxFind(
		tx,
		&unspents,
		query,
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return 0, err
		}
	}

	var balance uint64
	for _, v := range unspents {
		balance += v.Value
	}

	return balance, nil
}

func (u unspentRepositoryImpl) LockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) error {
	tx := ctx.Value("tx").(*badger.Txn)
	if tx == nil {
		return errors.New("context must contain db transaction value")
	}

	for _, v := range unspentKeys {
		var unspents []domain.Unspent
		err := u.db.Store.TxFind(
			tx,
			&unspents,
			badgerhold.Where(badgerhold.Key).Eq(v),
		)
		if err != nil {
			return err
		}

		unspent := unspents[0]
		unspent.Lock(&tradeID)

		err = u.db.Store.TxUpdate(
			tx,
			v,
			unspent,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u unspentRepositoryImpl) UnlockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey) error {
	tx := ctx.Value("tx").(*badger.Txn)
	if tx == nil {
		return errors.New("context must contain db transaction value")
	}

	for _, v := range unspentKeys {
		var unspents []domain.Unspent
		err := u.db.Store.TxFind(
			tx,
			&unspents,
			badgerhold.Where(badgerhold.Key).Eq(v),
		)
		if err != nil {
			return err
		}

		unspent := unspents[0]
		unspent.UnLock()

		err = u.db.Store.TxUpdate(
			tx,
			v,
			unspent,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u unspentRepositoryImpl) GetUnspentForKey(
	ctx context.Context,
	unspentKey domain.UnspentKey,
) (*domain.Unspent, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var unspents []domain.Unspent
	err := u.db.Store.TxFind(
		tx,
		&unspents,
		badgerhold.Where(badgerhold.Key).Eq(unspentKey),
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return nil, err
		}
	}

	return &unspents[0], nil
}

func (u unspentRepositoryImpl) UpdateUnspent(
	ctx context.Context,
	unspentKey domain.UnspentKey,
	updateFn func(m *domain.Unspent) (*domain.Unspent, error),
) error {
	tx := ctx.Value("tx").(*badger.Txn)
	if tx == nil {
		return errors.New("context must contain db transaction value")
	}

	currentUnspent, err := u.GetUnspentForKey(ctx, unspentKey)
	if err != nil {
		return err
	}

	updatedUnspent, err := updateFn(currentUnspent)
	if err != nil {
		return err
	}

	return u.db.Store.TxUpdate(
		tx,
		unspentKey,
		updatedUnspent,
	)
}

//
//func (u unspentRepositoryImpl) Begin() (uow.Tx, error) {
//	return nil, nil
//}
//func (u unspentRepositoryImpl) ContextKey() interface{} {
//	return nil
//}
