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

func (w WithdrawalRepositoryImpl) AddWithdrawals(
	_ context.Context,
	withdrawals []domain.Withdrawal,
) (int, error) {
	w.store.locker.Lock()
	defer w.store.locker.Unlock()

	count := 0
	for _, withdrawal := range withdrawals {
		if _, ok := w.store.withdrawals[withdrawal.TxID]; !ok {
			w.store.withdrawals[withdrawal.TxID] = withdrawal
			count++
		}
	}
	return count, nil
}

func (w WithdrawalRepositoryImpl) ListWithdrawalsForAccount(
	_ context.Context, accountIndex int,
) ([]domain.Withdrawal, error) {
	w.store.locker.RLock()
	defer w.store.locker.RUnlock()

	result := make([]domain.Withdrawal, 0)
	for _, v := range w.store.withdrawals {
		if v.AccountIndex == accountIndex {
			result = append(result, v)
		}
	}

	return result, nil
}
func (w WithdrawalRepositoryImpl) ListWithdrawalsForAccountAndPage(
	_ context.Context, accountIndex int, page domain.Page,
) ([]domain.Withdrawal, error) {
	w.store.locker.RLock()
	defer w.store.locker.RUnlock()

	result := make([]domain.Withdrawal, 0)
	startIndex := page.Number*page.Size - page.Size + 1
	endIndex := page.Number * page.Size
	index := 1
	for _, v := range w.store.withdrawals {
		if v.AccountIndex == accountIndex {
			if index >= startIndex && index <= endIndex {
				result = append(result, v)
			}
			index++
		}
	}

	return result, nil
}

func (w WithdrawalRepositoryImpl) ListAllWithdrawals(
	_ context.Context,
) ([]domain.Withdrawal, error) {
	withdrawals := make([]domain.Withdrawal, 0, len(w.store.withdrawals))
	for _, v := range w.store.withdrawals {
		withdrawals = append(withdrawals, v)
	}
	return withdrawals, nil
}

func (w WithdrawalRepositoryImpl) ListAllWithdrawalsForPage(
	_ context.Context, page domain.Page,
) ([]domain.Withdrawal, error) {
	withdrawals := make([]domain.Withdrawal, 0)
	startIndex := page.Number*page.Size - page.Size + 1
	endIndex := page.Number * page.Size
	index := 1
	for _, v := range w.store.withdrawals {
		if index >= startIndex && index <= endIndex {
			withdrawals = append(withdrawals, v)
		}
		index++
	}
	return withdrawals, nil
}
