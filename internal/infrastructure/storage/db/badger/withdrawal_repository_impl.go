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

func (w withdrawalRepositoryImpl) AddWithdrawal(
	ctx context.Context,
	withdrawal domain.Withdrawal,
) error {
	return w.insertWithdrawal(ctx, withdrawal)
}

func (w withdrawalRepositoryImpl) ListWithdrawalsForAccountIdAndPage(
	ctx context.Context,
	accountIndex int,
	page domain.Page,
) ([]domain.Withdrawal, error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)
	var withdrawals []domain.Withdrawal

	from := page.Number*page.Size - page.Size + 1

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		if err := w.store.TxFind(
			tx, &withdrawals,
			query.Skip(from).Limit(page.Size),
		); err != nil {
			return nil, err
		}
	} else {
		if err := w.store.Find(
			&withdrawals,
			query.Skip(from).Limit(page.Size),
		); err != nil {
			return nil, err
		}
	}

	return withdrawals, nil
}

func (w withdrawalRepositoryImpl) insertWithdrawal(
	ctx context.Context,
	withdrawal domain.Withdrawal,
) error {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = w.store.TxInsert(tx, withdrawal.TxID, &withdrawal)
	} else {
		err = w.store.Insert(withdrawal.TxID, &withdrawal)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}
