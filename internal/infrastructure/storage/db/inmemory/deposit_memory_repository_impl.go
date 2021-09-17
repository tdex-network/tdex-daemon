package inmemory

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

type DepositRepositoryImpl struct {
	store *depositInmemoryStore
}

// NewDepositRepositoryImpl returns a new empty DepositRepositoryImpl
func NewDepositRepositoryImpl(store *depositInmemoryStore) domain.DepositRepository {
	return &DepositRepositoryImpl{store}
}

func (d DepositRepositoryImpl) AddDeposits(
	ctx context.Context,
	deposits []domain.Deposit,
) (int, error) {
	d.store.locker.Lock()
	defer d.store.locker.Unlock()

	count := 0
	for _, deposit := range deposits {
		if _, ok := d.store.deposits[deposit.Key()]; !ok {
			d.store.deposits[deposit.Key()] = deposit
			count++
		}
	}

	return count, nil
}

func (d DepositRepositoryImpl) ListDepositsForAccount(
	ctx context.Context,
	accountIndex int,
) ([]domain.Deposit, error) {
	d.store.locker.RLock()
	defer d.store.locker.RUnlock()

	result := make([]domain.Deposit, 0)
	for _, v := range d.store.deposits {
		if v.AccountIndex == accountIndex {
			result = append(result, v)
		}
	}
	return result, nil
}

func (d DepositRepositoryImpl) ListDepositsForAccountAndPage(
	ctx context.Context,
	accountIndex int,
	page domain.Page,
) ([]domain.Deposit, error) {
	d.store.locker.RLock()
	defer d.store.locker.RUnlock()

	result := make([]domain.Deposit, 0)
	startIndex := page.Number*page.Size - page.Size + 1
	endIndex := page.Number * page.Size
	index := 1
	for _, v := range d.store.deposits {
		if v.AccountIndex == accountIndex {
			if index >= startIndex && index <= endIndex {
				result = append(result, v)
			}
			index++
		}
	}

	return result, nil
}

func (d DepositRepositoryImpl) ListAllDeposits(
	ctx context.Context,
) ([]domain.Deposit, error) {
	d.store.locker.RLock()
	defer d.store.locker.RUnlock()

	deposits := make([]domain.Deposit, 0, len(d.store.deposits))
	for _, v := range d.store.deposits {
		deposits = append(deposits, v)
	}
	return deposits, nil
}

func (d DepositRepositoryImpl) ListAllDepositsForPage(
	ctx context.Context, page domain.Page,
) ([]domain.Deposit, error) {
	d.store.locker.RLock()
	defer d.store.locker.RUnlock()

	deposits := make([]domain.Deposit, 0, len(d.store.deposits))
	startIndex := page.Number*page.Size - page.Size + 1
	endIndex := page.Number * page.Size
	index := 1
	for _, v := range d.store.deposits {
		if index >= startIndex && index <= endIndex {
			deposits = append(deposits, v)
		}
		index++
	}

	return deposits, nil
}
