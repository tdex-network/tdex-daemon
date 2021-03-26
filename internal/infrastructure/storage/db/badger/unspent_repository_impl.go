package dbbadger

import (
	"context"

	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
)

const (
	UnspentBadgerholdKeyPrefix       = "bh_Unspent"
	LockedUnspentBadgerholdKeyPrefix = "bh_LockedUnspent"
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
	return u.addUnspents(ctx, unspents)
}

func (u unspentRepositoryImpl) GetAllUnspents(
	ctx context.Context,
) []domain.Unspent {
	return u.getAllUnspents(ctx)
}

func (u unspentRepositoryImpl) GetAvailableUnspents(
	ctx context.Context,
) ([]domain.Unspent, error) {
	query := badgerhold.Where("Spent").Eq(false).And("Confirmed").Eq(true)
	unlockedOnly := true

	unspents, err := u.findUnspents(ctx, query, unlockedOnly)
	if err != nil {
		return nil, err
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetAllUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	iface := make([]interface{}, 0, len(addresses))
	for _, v := range addresses {
		iface = append(iface, v)
	}

	query := badgerhold.Where("Address").In(iface...)
	unlockedOnly := false

	unspents, err := u.findUnspents(ctx, query, unlockedOnly)
	if err != nil {
		return nil, err
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	iface := make([]interface{}, 0, len(addresses))
	for _, v := range addresses {
		iface = append(iface, v)
	}

	query := badgerhold.Where("Spent").Eq(false).
		And("Confirmed").Eq(true).
		And("Address").In(iface...)
	unlockedOnly := false

	unspents, err := u.findUnspents(ctx, query, unlockedOnly)
	if err != nil {
		return nil, err
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetAvailableUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	iface := make([]interface{}, 0, len(addresses))
	for _, v := range addresses {
		iface = append(iface, v)
	}

	query := badgerhold.Where("Spent").Eq(false).
		And("Confirmed").Eq(true).
		And("Address").In(iface...)
	unlockedOnly := true

	unspents, err := u.findUnspents(ctx, query, unlockedOnly)
	if err != nil {
		return nil, err
	}

	return unspents, nil
}

func (u unspentRepositoryImpl) GetUnspentWithKey(
	ctx context.Context,
	unspentKey domain.UnspentKey,
) (*domain.Unspent, error) {
	return u.getUnspent(ctx, unspentKey)
}

func (u unspentRepositoryImpl) GetBalance(
	ctx context.Context,
	addresses []string,
	assetHash string,
) (uint64, error) {
	unlockedOnly := false
	return u.getBalance(ctx, addresses, assetHash, unlockedOnly)
}

func (u unspentRepositoryImpl) GetUnlockedBalance(
	ctx context.Context,
	addresses []string,
	assetHash string,
) (uint64, error) {
	unlockedOnly := true
	return u.getBalance(ctx, addresses, assetHash, unlockedOnly)
}

func (u unspentRepositoryImpl) SpendUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	return u.spendUnspents(ctx, unspentKeys)
}

func (u unspentRepositoryImpl) ConfirmUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	return u.confirmUnspents(ctx, unspentKeys)
}

func (u unspentRepositoryImpl) LockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) (int, error) {
	return u.lockUnspents(ctx, unspentKeys, tradeID)
}

func (u unspentRepositoryImpl) UnlockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	return u.unlockUnspents(ctx, unspentKeys)
}

func (u unspentRepositoryImpl) addUnspents(
	ctx context.Context,
	unspents []domain.Unspent,
) error {
	for _, v := range unspents {
		if err := u.insertUnspent(ctx, v); err != nil {
			return err
		}
	}

	return nil
}

func (u unspentRepositoryImpl) getAllUnspents(ctx context.Context) []domain.Unspent {
	scan := func(tx *badger.Txn) []domain.Unspent {
		unspents := make([]domain.Unspent, 0)

		iter := badger.DefaultIteratorOptions
		it := tx.NewIterator(iter)
		iter.PrefetchValues = false
		defer it.Close()

		for it.Seek(unspentTablePrefixKey); it.ValidForPrefix(unspentTablePrefixKey); it.Next() {
			item := it.Item()
			data, _ := item.ValueCopy(nil)
			var unspent domain.Unspent
			err := badgerhold.DefaultDecode(data, &unspent)
			if err == nil {
				tradeID, err := u.getLock(ctx, unspent.Key())
				if err == nil {
					if tradeID != nil {
						unspent.Lock(tradeID)
					}
					unspents = append(unspents, unspent)
				}
			}
		}

		return unspents
	}

	if ctx.Value("utx") != nil {
		return scan(ctx.Value("utx").(*badger.Txn))
	}

	var unspents []domain.Unspent
	u.db.UnspentStore.Badger().View(func(tx *badger.Txn) error {
		unspents = scan(tx)
		return nil
	})
	return unspents
}

func (u unspentRepositoryImpl) getBalance(
	ctx context.Context,
	addresses []string,
	assetHash string,
	unlockedOnly bool,
) (uint64, error) {
	iface := make([]interface{}, 0, len(addresses))
	for _, v := range addresses {
		iface = append(iface, v)
	}

	query := badgerhold.Where("AssetHash").Eq(assetHash).
		And("Spent").Eq(false).
		And("Confirmed").Eq(true).
		And("Address").In(iface...)

	unspents, err := u.findUnspents(ctx, query, unlockedOnly)
	if err != nil {
		return 0, err
	}

	var balance uint64
	for _, v := range unspents {
		balance += v.Value
	}

	return balance, nil
}

func (u unspentRepositoryImpl) spendUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	count := 0
	for _, key := range unspentKeys {
		done, err := u.spendUnspent(ctx, key)
		if err != nil {
			return -1, err
		}
		if done {
			count++
		}
	}
	return count, nil
}

func (u unspentRepositoryImpl) spendUnspent(
	ctx context.Context,
	key domain.UnspentKey,
) (bool, error) {
	unspent, err := u.getUnspent(ctx, key)
	if err != nil {
		return false, err
	}

	if unspent == nil {
		return false, nil
	}

	unspent.Spend()
	unspent.Unlock() // prevent conflict, locks not stored under unspent prefix

	if err := u.updateUnspent(ctx, key, *unspent); err != nil {
		return false, err
	}

	return true, nil
}

func (u unspentRepositoryImpl) confirmUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	count := 0
	for _, key := range unspentKeys {
		done, err := u.confirmUnspent(ctx, key)
		if err != nil {
			return -1, err
		}
		if done {
			count++
		}
	}
	return count, nil
}

func (u unspentRepositoryImpl) confirmUnspent(
	ctx context.Context,
	key domain.UnspentKey,
) (bool, error) {
	unspent, err := u.getUnspent(ctx, key)
	if err != nil {
		return false, err
	}

	if unspent == nil {
		return false, nil
	}

	unspent.Confirm()
	unspent.Unlock() // prevent conflict, locks not stored under unspent prefix

	if err := u.updateUnspent(ctx, key, *unspent); err != nil {
		return false, err
	}

	return true, nil
}

func (u unspentRepositoryImpl) lockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) (int, error) {
	count := 0
	for _, key := range unspentKeys {
		done, err := u.lockUnspent(ctx, key, tradeID)
		if err != nil {
			return -1, err
		}
		if done {
			count++
		}
	}
	return count, nil
}

func (u unspentRepositoryImpl) lockUnspent(
	ctx context.Context,
	key domain.UnspentKey,
	tradeID uuid.UUID,
) (bool, error) {
	unspent, err := u.getUnspent(ctx, key)
	if err != nil {
		return false, err
	}

	if unspent == nil {
		return false, nil
	}

	if unspent.IsLocked() {
		return false, nil
	}

	if err := u.insertLock(ctx, key, tradeID); err != nil {
		return false, err
	}

	return true, nil
}

func (u unspentRepositoryImpl) unlockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	count := 0
	for _, key := range unspentKeys {
		done, err := u.unlockUnspent(ctx, key)
		if err != nil {
			return -1, err
		}
		if done {
			count++
		}
	}
	return count, nil
}

func (u unspentRepositoryImpl) unlockUnspent(
	ctx context.Context,
	key domain.UnspentKey,
) (bool, error) {
	unspent, err := u.getUnspent(ctx, key)
	if err != nil {
		return false, err
	}

	if unspent == nil {
		return false, nil
	}

	if !unspent.IsLocked() {
		return false, nil
	}

	if err := u.deleteLock(ctx, key); err != nil {
		return false, err
	}

	return true, nil
}

func (u unspentRepositoryImpl) findUnspents(
	ctx context.Context,
	query *badgerhold.Query,
	unlockedOnly bool,
) ([]domain.Unspent, error) {
	var unspents []domain.Unspent
	var unlockedUnspents []domain.Unspent

	if ctx.Value("utx") != nil {
		tx := ctx.Value("utx").(*badger.Txn)
		if err := u.db.UnspentStore.TxFind(tx, &unspents, query); err != nil {
			return nil, err
		}
	} else {
		if err := u.db.UnspentStore.Find(&unspents, query); err != nil {
			return nil, err
		}
	}

	for _, unspent := range unspents {
		tradeID, err := u.getLock(ctx, unspent.Key())
		if err != nil {
			return nil, err
		}
		if tradeID != nil {
			unspent.Lock(tradeID)
		} else {
			unlockedUnspents = append(unlockedUnspents, unspent)
		}
	}

	if unlockedOnly {
		return unlockedUnspents, nil
	}
	return unspents, nil
}

func (u unspentRepositoryImpl) getUnspent(
	ctx context.Context,
	key domain.UnspentKey,
) (*domain.Unspent, error) {
	var unspent domain.Unspent
	var err error

	if ctx.Value("utx") != nil {
		tx := ctx.Value("utx").(*badger.Txn)
		err = u.db.UnspentStore.TxGet(tx, key, &unspent)
	} else {
		err = u.db.UnspentStore.Get(key, &unspent)
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	tradeID, err := u.getLock(ctx, key)
	if err != nil {
		return nil, err
	}
	if tradeID != nil {
		unspent.Lock(tradeID)
	}

	return &unspent, nil
}

func (u unspentRepositoryImpl) updateUnspent(
	ctx context.Context,
	key domain.UnspentKey,
	unspent domain.Unspent,
) error {
	if ctx.Value("utx") != nil {
		tx := ctx.Value("utx").(*badger.Txn)
		return u.db.UnspentStore.TxUpdate(tx, key, unspent)
	}

	return u.db.UnspentStore.Update(key, unspent)
}

func (u unspentRepositoryImpl) insertUnspent(
	ctx context.Context,
	unspent domain.Unspent,
) error {
	var err error
	if ctx.Value("utx") != nil {
		tx := ctx.Value("utx").(*badger.Txn)
		err = u.db.UnspentStore.TxInsert(tx, unspent.Key(), &unspent)
	} else {
		err = u.db.UnspentStore.Insert(unspent.Key(), &unspent)
	}

	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}

type LockedUnspent struct {
	TradeID uuid.UUID
}

func (u unspentRepositoryImpl) getLock(
	ctx context.Context,
	key domain.UnspentKey,
) (*uuid.UUID, error) {
	var lockedUnspent LockedUnspent
	var err error

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = u.db.Store.TxGet(tx, key, &lockedUnspent)
	} else {
		err = u.db.Store.Get(key, &lockedUnspent)
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &lockedUnspent.TradeID, nil
}

func (u unspentRepositoryImpl) insertLock(
	ctx context.Context,
	key domain.UnspentKey,
	tradeID uuid.UUID,
) error {
	var err error

	encKey, err := EncodeKey(key, LockedUnspentBadgerholdKeyPrefix)
	if err != nil {
		return err
	}

	encData, err := badgerhold.DefaultEncode(LockedUnspent{tradeID})
	if err != nil {
		return err
	}

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		return tx.SetEntry(
			badger.NewEntry(encKey, encData),
		)
	} else {
		err = u.db.Store.Badger().Update(func(txn *badger.Txn) error {
			return txn.SetEntry(
				badger.NewEntry(encKey, encData),
			)
		})
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}

func (u unspentRepositoryImpl) deleteLock(
	ctx context.Context,
	key domain.UnspentKey,
) error {
	var err error

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = u.db.Store.TxDelete(tx, key, LockedUnspent{})
	} else {
		err = u.db.Store.Delete(key, LockedUnspent{})
	}
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return err
		}
	}
	return nil
}
