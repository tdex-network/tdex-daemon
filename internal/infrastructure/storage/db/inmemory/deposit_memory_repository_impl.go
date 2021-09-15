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

func (d DepositRepositoryImpl) AddDeposit(
	ctx context.Context,
	deposit domain.Deposit,
) error {
	d.store.locker.Lock()
	defer d.store.locker.Unlock()

	if _, ok := d.store.deposits[deposit.Key()]; !ok {
		d.store.deposits[deposit.Key()] = deposit
	}

	return nil
}

func (d DepositRepositoryImpl) ListDepositsForAccountId(
	ctx context.Context,
	accountIndex int,
	page *domain.Page,
) ([]domain.Deposit, error) {
	d.store.locker.RLock()
	defer d.store.locker.RUnlock()

	result := make([]domain.Deposit, 0)

	if page == nil {
		for _, v := range d.store.deposits {
			if v.AccountIndex == accountIndex {
				result = append(result, v)
			}
		}
		return result, nil
	}

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
	ctx context.Context, page *domain.Page,
) ([]domain.Deposit, error) {
	d.store.locker.RLock()
	defer d.store.locker.RUnlock()

	deposits := make([]domain.Deposit, 0, len(d.store.deposits))

	if page == nil {
		for _, v := range d.store.deposits {
			deposits = append(deposits, v)
		}
		return deposits, nil
	}

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
