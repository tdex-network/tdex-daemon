package dbbadger

import (
	"context"

	"github.com/dgraph-io/badger/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
)

type withdrawalRepositoryImpl struct {
	store *badgerhold.Store
}

// NewWithdrawalRepositoryImpl initialize a badger implementation of the domain.StatsRepository
func NewWithdrawalRepositoryImpl(store *badgerhold.Store) domain.WithdrawalRepository {
	return withdrawalRepositoryImpl{store}
}

func (w withdrawalRepositoryImpl) AddWithdrawals(
	ctx context.Context,
	withdrawals []domain.Withdrawal,
) (int, error) {
	return w.insertWithdrawals(ctx, withdrawals)
}

func (w withdrawalRepositoryImpl) ListWithdrawalsForAccount(
	ctx context.Context, accountIndex int,
) ([]domain.Withdrawal, error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)

	return w.findWithdrawals(ctx, query)
}

func (w withdrawalRepositoryImpl) ListWithdrawalsForAccountAndPage(
	ctx context.Context, accountIndex int, page domain.Page,
) ([]domain.Withdrawal, error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)
	from := page.Number*page.Size - page.Size
	query.Skip(from).Limit(page.Size)

	return w.findWithdrawals(ctx, query)
}

func (w withdrawalRepositoryImpl) ListAllWithdrawals(
	ctx context.Context,
) ([]domain.Withdrawal, error) {
	query := &badgerhold.Query{}
	return w.findWithdrawals(ctx, query)
}

func (w withdrawalRepositoryImpl) ListAllWithdrawalsForPage(
	ctx context.Context, page domain.Page,
) ([]domain.Withdrawal, error) {
	query := &badgerhold.Query{}
	from := page.Number*page.Size - page.Size
	query.Skip(from).Limit(page.Size)

	return w.findWithdrawals(ctx, query)
}

func (w withdrawalRepositoryImpl) insertWithdrawals(
	ctx context.Context,
	withdrawals []domain.Withdrawal,
) (int, error) {
	count := 0
	for _, ww := range withdrawals {
		done, err := w.insertWithdrawal(ctx, ww)
		if err != nil {
			return -1, err
		}
		if done {
			count++
		}
	}
	return count, nil

}

func (w withdrawalRepositoryImpl) insertWithdrawal(
	ctx context.Context,
	withdrawal domain.Withdrawal,
) (bool, error) {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = w.store.TxInsert(tx, withdrawal.TxID, &withdrawal)
	} else {
		err = w.store.Insert(withdrawal.TxID, &withdrawal)
	}
	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (w withdrawalRepositoryImpl) findWithdrawals(
	ctx context.Context,
	query *badgerhold.Query,
) ([]domain.Withdrawal, error) {
	var withdrawals []domain.Withdrawal
	var err error

	query.SortBy("Timestamp").Reverse()
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = w.store.TxFind(tx, &withdrawals, query)
	} else {
		err = w.store.Find(&withdrawals, query)
	}
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
