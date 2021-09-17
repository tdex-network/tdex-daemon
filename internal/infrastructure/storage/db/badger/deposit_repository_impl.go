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

func (d depositRepositoryImpl) AddDeposits(
	ctx context.Context,
	deposits []domain.Deposit,
) (int, error) {
	return d.insertDeposits(ctx, deposits)
}

func (d depositRepositoryImpl) ListDepositsForAccount(
	ctx context.Context, accountIndex int,
) ([]domain.Deposit, error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)
	return d.findDeposits(ctx, query)
}

func (d depositRepositoryImpl) ListDepositsForAccountAndPage(
	ctx context.Context, accountIndex int, page domain.Page,
) ([]domain.Deposit, error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)
	from := page.Number*page.Size - page.Size
	query.Skip(from).Limit(page.Size)
	return d.findDeposits(ctx, query)
}

func (d depositRepositoryImpl) ListAllDeposits(
	ctx context.Context,
) ([]domain.Deposit, error) {
	return d.findDeposits(ctx, nil)
}

func (d depositRepositoryImpl) ListAllDepositsForPage(
	ctx context.Context, page domain.Page,
) ([]domain.Deposit, error) {
	from := page.Number*page.Size - page.Size
	query := &badgerhold.Query{}
	query.Skip(from).Limit(page.Size)
	return d.findDeposits(ctx, query)
}

func (d depositRepositoryImpl) insertDeposits(
	ctx context.Context,
	deposits []domain.Deposit,
) (int, error) {
	count := 0
	for _, dd := range deposits {
		done, err := d.insertDeposit(ctx, dd)
		if err != nil {
			return -1, err
		}
		if done {
			count++
		}
	}

	return count, nil
}

func (d depositRepositoryImpl) insertDeposit(
	ctx context.Context,
	deposit domain.Deposit,
) (bool, error) {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = d.store.TxInsert(tx, deposit.Key(), &deposit)
	} else {
		err = d.store.Insert(deposit.Key(), &deposit)
	}
	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return false, nil
		}
		return false, err
	}
	return true, nil
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
