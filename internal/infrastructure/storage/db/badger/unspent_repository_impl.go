package dbbadger

import (
	"context"
	"errors"

	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
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

	return u.addUnspents(tx, unspents)
}

func (u unspentRepositoryImpl) GetAllUnspents(
	ctx context.Context,
) []domain.Unspent {
	tx := ctx.Value("tx").(*badger.Txn)

	return u.getAllUnspents(tx)
}

func (u unspentRepositoryImpl) GetBalance(
	ctx context.Context,
	address string,
	assetHash string,
) (uint64, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("Address").Eq(address).
		And("AssetHash").Eq(assetHash).
		And("Spent").Eq(false).
		And("Confirmed").Eq(true)

	unspents, err := u.findUnspents(tx, query)
	if err != nil {
		return 0, err
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

	query := badgerhold.Where("Spent").Eq(false).
		And("Locked").Eq(false).
		And("Confirmed").Eq(true)

	unspents, err := u.findUnspents(tx, query)
	if err != nil {
		return nil, err
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	iface := make([]interface{}, 0, len(addresses))
	for _, v := range addresses {
		iface = append(iface, v)
	}

	query := badgerhold.Where("Address").In(iface...)

	unspents, err := u.findUnspents(tx, query)
	if err != nil {
		return nil, err
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetAvailableUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	iface := make([]interface{}, 0, len(addresses))
	for _, v := range addresses {
		iface = append(iface, v)
	}

	query := badgerhold.Where("Spent").Eq(false).
		And("Locked").Eq(false).
		And("Confirmed").Eq(true).
		And("Address").In(iface...)

	unspents, err := u.findUnspents(tx, query)
	if err != nil {
		return nil, err
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetUnlockedBalance(
	ctx context.Context,
	address string,
	assetHash string,
) (uint64, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("Address").Eq(address).
		And("AssetHash").Eq(assetHash).
		And("Spent").Eq(false).
		And("Confirmed").Eq(true).
		And("Locked").Eq(false)

	unspents, err := u.findUnspents(tx, query)
	if err != nil {
		return 0, err
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

	return u.lockUnspents(tx, unspentKeys, tradeID)
}

func (u unspentRepositoryImpl) UnlockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey) error {
	tx := ctx.Value("tx").(*badger.Txn)
	if tx == nil {
		return errors.New("context must contain db transaction value")
	}

	return u.unlockUnspents(tx, unspentKeys)
}

func (u unspentRepositoryImpl) GetUnspentForKey(
	ctx context.Context,
	unspentKey domain.UnspentKey,
) (*domain.Unspent, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	return u.getUnspent(tx, unspentKey)
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

	return u.updateUnspent(tx, unspentKey, *updatedUnspent)
}

func (u unspentRepositoryImpl) addUnspents(
	tx *badger.Txn,
	unspents []domain.Unspent,
) error {
	for _, v := range unspents {
		if err := u.insertUnspent(tx, v); err != nil {
			return err
		}
	}

	return nil
}

func (u unspentRepositoryImpl) getAllUnspents(tx *badger.Txn) []domain.Unspent {
	unspents := make([]domain.Unspent, 0)

	iter := badger.DefaultIteratorOptions
	iter.PrefetchValues = false
	it := tx.NewIterator(iter)
	defer it.Close()

	for it.Seek(unspentTablePrefixKey); it.ValidForPrefix(unspentTablePrefixKey); it.Next() {
		item := it.Item()
		data, _ := item.ValueCopy(nil)
		var unspent domain.Unspent
		err := JSONDecode(data, &unspent)
		if err == nil {
			unspents = append(unspents, unspent)
		}
	}

	return unspents
}

func (u unspentRepositoryImpl) lockUnspents(
	tx *badger.Txn,
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) error {
	for _, key := range unspentKeys {
		if err := u.lockUnspent(tx, key, tradeID); err != nil {
			return err
		}
	}
	return nil
}

func (u unspentRepositoryImpl) lockUnspent(
	tx *badger.Txn,
	key domain.UnspentKey,
	tradeID uuid.UUID,
) error {
	unspent, err := u.getUnspent(tx, key)
	if err != nil {
		return err
	}

	if unspent == nil {
		return nil
	}

	unspent.Lock(&tradeID)

	return u.updateUnspent(tx, key, *unspent)
}

func (u unspentRepositoryImpl) unlockUnspents(
	tx *badger.Txn,
	unspentKeys []domain.UnspentKey,
) error {
	for _, key := range unspentKeys {
		if err := u.unlockUnspent(tx, key); err != nil {
			return err
		}
	}
	return nil
}

func (u unspentRepositoryImpl) unlockUnspent(
	tx *badger.Txn,
	key domain.UnspentKey,
) error {
	unspent, err := u.getUnspent(tx, key)
	if err != nil {
		return err
	}

	if unspent == nil {
		return nil
	}

	unspent.UnLock()

	return u.updateUnspent(tx, key, *unspent)
}

func (u unspentRepositoryImpl) findUnspents(
	tx *badger.Txn,
	query *badgerhold.Query,
) ([]domain.Unspent, error) {
	var unspents []domain.Unspent
	err := u.db.Store.TxFind(
		tx,
		&unspents,
		query,
	)

	return unspents, err
}

func (u unspentRepositoryImpl) getUnspent(
	tx *badger.Txn,
	key domain.UnspentKey,
) (*domain.Unspent, error) {
	var unspent domain.Unspent
	err := u.db.Store.TxGet(
		tx,
		key,
		&unspent,
	)
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &unspent, nil
}

func (u unspentRepositoryImpl) updateUnspent(
	tx *badger.Txn,
	key domain.UnspentKey,
	unspent domain.Unspent,
) error {
	return u.db.Store.TxUpdate(
		tx,
		key,
		unspent,
	)
}

func (u unspentRepositoryImpl) insertUnspent(
	tx *badger.Txn,
	unspent domain.Unspent,
) error {
	if err := u.db.Store.TxInsert(
		tx,
		unspent.Key(),
		&unspent,
	); err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}
