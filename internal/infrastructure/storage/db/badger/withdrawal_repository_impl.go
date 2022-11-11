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

// NewWithdrawalRepositoryImpl is the factory for a badger implementation of
// domain.WithdrawalRepository
func NewWithdrawalRepositoryImpl(
	store *badgerhold.Store,
) domain.WithdrawalRepository {
	return withdrawalRepositoryImpl{store}
}

func (w withdrawalRepositoryImpl) AddWithdrawals(
	ctx context.Context,
	withdrawals []domain.Withdrawal,
) (int, error) {
	return w.insertWithdrawals(ctx, withdrawals)
}

func (w withdrawalRepositoryImpl) GetWithdrawalsForAccount(
	ctx context.Context, accountName string, page domain.Page,
) ([]domain.Withdrawal, error) {
	query := badgerhold.Where("AccountName").Eq(accountName)
	if page != nil {
		offset := int(page.GetNumber()*page.GetSize() - page.GetSize())
		query.Skip(offset).Limit(int(page.GetSize()))
	}
	return w.findWithdrawals(ctx, query)
}

func (w withdrawalRepositoryImpl) GetAllWithdrawals(
	ctx context.Context, page domain.Page,
) ([]domain.Withdrawal, error) {
	query := &badgerhold.Query{}
	if page != nil {
		offset := int(page.GetNumber()*page.GetSize() - page.GetSize())
		query.Skip(offset).Limit(int(page.GetSize()))
	}
	return w.findWithdrawals(ctx, query)
}

func (w withdrawalRepositoryImpl) insertWithdrawals(
	ctx context.Context, withdrawals []domain.Withdrawal,
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
	ctx context.Context, withdrawal domain.Withdrawal,
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
	ctx context.Context, query *badgerhold.Query,
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
