package inmemory

import (
	"context"
	"sort"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

type withdrawalInmemoryStore struct {
	withdrawals map[string]domain.Withdrawal
	locker      *sync.RWMutex
}

type withdrawalRepositoryImpl struct {
	store *withdrawalInmemoryStore
}

// NewWithdrawalRepositoryImpl returns a new empty DepositRepositoryImpl
func NewWithdrawalRepositoryImpl() domain.WithdrawalRepository {
	return &withdrawalRepositoryImpl{&withdrawalInmemoryStore{
		withdrawals: map[string]domain.Withdrawal{},
		locker:      &sync.RWMutex{},
	}}
}

func (w *withdrawalRepositoryImpl) AddWithdrawals(
	_ context.Context, withdrawals []domain.Withdrawal,
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

func (w *withdrawalRepositoryImpl) GetWithdrawalsForAccount(
	_ context.Context, accountName string, page domain.Page,
) ([]domain.Withdrawal, error) {
	w.store.locker.RLock()
	defer w.store.locker.RUnlock()

	withdrawals := make([]domain.Withdrawal, 0)
	for _, withdrawal := range w.store.withdrawals {
		if withdrawal.AccountName == accountName {
			withdrawals = append(withdrawals, withdrawal)
		}
	}
	sort.SliceStable(withdrawals, func(i, j int) bool {
		return withdrawals[i].Timestamp > withdrawals[j].Timestamp
	})

	if page == nil {
		return withdrawals, nil
	}

	pagWithdrawals := make([]domain.Withdrawal, 0, page.GetSize())
	startIndex := int(page.GetNumber()*page.GetSize() - page.GetSize())
	endIndex := int(page.GetNumber() * page.GetSize())
	for i, withdrawal := range withdrawals {
		if i >= startIndex && i < endIndex {
			pagWithdrawals = append(pagWithdrawals, withdrawal)
		}
	}

	return pagWithdrawals, nil
}

func (w *withdrawalRepositoryImpl) GetAllWithdrawals(
	_ context.Context, page domain.Page,
) ([]domain.Withdrawal, error) {
	w.store.locker.RLock()
	defer w.store.locker.RUnlock()

	withdrawals := make([]domain.Withdrawal, 0)
	for _, withdrawal := range w.store.withdrawals {
		withdrawals = append(withdrawals, withdrawal)
	}
	sort.SliceStable(withdrawals, func(i, j int) bool {
		return withdrawals[i].Timestamp > withdrawals[j].Timestamp
	})

	if page == nil {
		return withdrawals, nil
	}

	pagWithdrawals := make([]domain.Withdrawal, 0, page.GetSize())
	startIndex := int(page.GetNumber()*page.GetSize() - page.GetSize())
	endIndex := int(page.GetNumber() * page.GetSize())
	for i, withdrawal := range withdrawals {
		if i >= startIndex && i < endIndex {
			pagWithdrawals = append(pagWithdrawals, withdrawal)
		}
	}

	return pagWithdrawals, nil
}
