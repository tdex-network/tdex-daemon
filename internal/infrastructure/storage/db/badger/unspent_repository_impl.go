package dbbadger

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/uow"
	"github.com/timshannon/badgerhold"
)

const (
	unspentTableName = "MARKET_"
)

var unspentTablePrefixKey = []byte(unspentTableName)

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
			fmt.Sprintf("%v%v", unspentTableName, v.Key()),
			&v,
		); err != nil {
			return err
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
	it := tx.NewIterator(iter)
	defer it.Close()

	for it.Seek(unspentTablePrefixKey); it.ValidForPrefix(unspentTablePrefixKey); it.Next() {
		item := it.Item()
		data, _ := item.ValueCopy(nil)
		var unspent domain.Unspent
		err := badgerhold.DefaultDecode(data, &unspent)
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
) uint64 {
	var markets []Market
	err := m.db.Store.TxFind(
		tx,
		&markets,
		badgerhold.Where("AccountIndex").Ge(domain.MarketAccountStart).And("Tradable").Eq(true),
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return nil, err
		}
	}
}

func (u unspentRepositoryImpl) GetAvailableUnspents(
	ctx context.Context,
) []domain.Unspent {
	panic("implement me")
}

func (u unspentRepositoryImpl) GetAvailableUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) []domain.Unspent {
	panic("implement me")
}

func (u unspentRepositoryImpl) GetUnlockedBalance(
	ctx context.Context,
	address string,
	assetHash string,
) uint64 {
	panic("implement me")
}

func (u unspentRepositoryImpl) LockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) error {
	panic("implement me")
}

func (u unspentRepositoryImpl) UnlockUnspents(
	ctx context.Context,
	unspentKey []domain.UnspentKey) error {
	panic("implement me")
}

func (u unspentRepositoryImpl) Begin() (uow.Tx, error) {
	panic("implement me")
}

func (u unspentRepositoryImpl) ContextKey() interface{} {
	panic("implement me")
}

func (u unspentRepositoryImpl) GetBalanceInfoForAsset(unspents []domain.Unspent) map[string]domain.BalanceInfo {
	panic("implement me")
}
