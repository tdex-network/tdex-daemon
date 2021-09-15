package dbbadger

import (
	"context"

	"github.com/dgraph-io/badger/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
)

type depositRepositoryImpl struct {
	store *badgerhold.Store
}

// NewDepositRepositoryImpl initialize a badger implementation of the domain.StatsRepository
func NewDepositRepositoryImpl(store *badgerhold.Store) domain.DepositRepository {
	return depositRepositoryImpl{store}
}

func (d depositRepositoryImpl) AddDeposit(
	ctx context.Context,
	deposit domain.Deposit,
) error {
	return d.insertDeposit(ctx, deposit)
}

func (d depositRepositoryImpl) ListDepositsForAccountId(
	ctx context.Context, accountIndex int, page *domain.Page,
) ([]domain.Deposit, error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)
	if page != nil {
		from := page.Number*page.Size - page.Size
		query.Skip(from).Limit(page.Size)
	}
	return d.findDeposits(ctx, query)
}

func (d depositRepositoryImpl) ListAllDeposits(
	ctx context.Context, page *domain.Page,
) ([]domain.Deposit, error) {
	var query *badgerhold.Query
	if page != nil {
		from := page.Number*page.Size - page.Size
		query = &badgerhold.Query{}
		query.Skip(from).Limit(page.Size)
	}
	return d.findDeposits(ctx, query)
}

func (d depositRepositoryImpl) insertDeposit(
	ctx context.Context,
	deposit domain.Deposit,
) error {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = d.store.TxInsert(tx, deposit.Key(), &deposit)
	} else {
		err = d.store.Insert(deposit.Key(), &deposit)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}

func (d depositRepositoryImpl) findDeposits(
	ctx context.Context,
	query *badgerhold.Query,
) ([]domain.Deposit, error) {
	var deposits []domain.Deposit
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = d.store.TxFind(tx, &deposits, query)
	} else {
		err = d.store.Find(&deposits, query)
	}
	if err != nil {
		return nil, err
	}

	return deposits, nil
}
