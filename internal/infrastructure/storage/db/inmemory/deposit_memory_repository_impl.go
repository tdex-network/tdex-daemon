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

	d.store.deposits[deposit.Key()] = deposit

	return nil
}

func (d DepositRepositoryImpl) ListDepositsForAccountIdAndPage(
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
		if index >= startIndex && index <= endIndex {
			if v.AccountIndex == accountIndex {
				result = append(result, v)
			}
		}
		index++
	}

	return result, nil
}
