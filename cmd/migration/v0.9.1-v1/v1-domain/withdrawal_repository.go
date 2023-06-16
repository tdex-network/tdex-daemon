package v1domain

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v4"
)

type WithdrawalsRepository interface {
	InsertWithdrawals(withdrawals []*domain.Withdrawal) error
}

type withdrawalRepositoryImpl struct {
	store *badgerhold.Store
}

func NewWithdrawalsRepositoryImpl(store *badgerhold.Store) WithdrawalsRepository {
	return &withdrawalRepositoryImpl{store}
}

func (w *withdrawalRepositoryImpl) InsertWithdrawals(
	withdrawals []*domain.Withdrawal,
) error {
	for _, v := range withdrawals {
		if err := w.store.Insert(v.TxID, *v); err != nil {
			return err
		}
	}

	return nil
}
