package inmemory

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

type WithdrawalRepositoryImpl struct {
	store *withdrawalInmemoryStore
}

// NewWithdrawalRepositoryImpl returns a new empty DepositRepositoryImpl
func NewWithdrawalRepositoryImpl(store *withdrawalInmemoryStore) domain.WithdrawalRepository {
	return &WithdrawalRepositoryImpl{store}
}

func (w WithdrawalRepositoryImpl) AddWithdrawal(
	ctx context.Context,
	withdrawal domain.Withdrawal,
) error {
	w.store.locker.Lock()
	defer w.store.locker.Unlock()

	w.store.withdrawals[withdrawal.TxID] = withdrawal

	return nil
}

func (w WithdrawalRepositoryImpl) ListWithdrawalsForAccountIdAndPage(
	ctx context.Context,
	accountIndex int,
	page domain.Page,
) ([]domain.Withdrawal, error) {
	w.store.locker.RLock()
	defer w.store.locker.RUnlock()

	result := make([]domain.Withdrawal, 0)

	startIndex := page.Number*page.Size - page.Size + 1
	endIndex := page.Number * page.Size
	index := 1
	for _, v := range w.store.withdrawals {
		if index >= startIndex && index <= endIndex {
			if v.AccountIndex == accountIndex {
				result = append(result, v)
			}
		}
		index++
	}

	return result, nil
}
