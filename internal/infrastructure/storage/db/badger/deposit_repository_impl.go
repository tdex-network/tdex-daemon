package dbbadger

import (
	"context"

	"github.com/dgraph-io/badger/v3"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v4"
)

type depositRepositoryImpl struct {
	store *badgerhold.Store
}

// NewDepositRepositoryImpl initialize a badger implementation of the domain.StatsRepository
func NewDepositRepositoryImpl(
	store *badgerhold.Store,
) domain.DepositRepository {
	return depositRepositoryImpl{store}
}

func (d depositRepositoryImpl) AddDeposits(
	ctx context.Context,
	deposits []domain.Deposit,
) (int, error) {
	return d.insertDeposits(ctx, deposits)
}

func (d depositRepositoryImpl) GetDepositsForAccount(
	ctx context.Context, accountName string, page domain.Page,
) ([]domain.Deposit, error) {
	query := badgerhold.Where("AccountName").Eq(accountName)
	if page != nil {
		offset := int(page.GetNumber()*page.GetSize() - page.GetSize())
		query.Skip(offset).Limit(int(page.GetSize()))
	}
	return d.findDeposits(ctx, query)
}

func (d depositRepositoryImpl) GetAllDeposits(
	ctx context.Context, page domain.Page,
) ([]domain.Deposit, error) {
	query := &badgerhold.Query{}
	if page != nil {
		offset := int(page.GetNumber()*page.GetSize() - page.GetSize())
		query.Skip(offset).Limit(int(page.GetSize()))
	}
	return d.findDeposits(ctx, query)
}

func (d depositRepositoryImpl) insertDeposits(
	ctx context.Context, deposits []domain.Deposit,
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
	ctx context.Context, deposit domain.Deposit,
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
	ctx context.Context, query *badgerhold.Query,
) ([]domain.Deposit, error) {
	var deposits []domain.Deposit
	var err error

	query.SortBy("Timestamp").Reverse()
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
