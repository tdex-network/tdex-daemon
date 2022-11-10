package inmemory

import (
	"context"
	"sort"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

type depositInMemoryStore struct {
	deposits map[string]domain.Deposit
	locker   *sync.RWMutex
}

type depositRepositoryImpl struct {
	store *depositInMemoryStore
}

// NewDepositRepositoryImpl returns a new empty DepositRepositoryImpl
func NewDepositRepositoryImpl() domain.DepositRepository {
	return &depositRepositoryImpl{&depositInMemoryStore{
		deposits: map[string]domain.Deposit{},
		locker:   &sync.RWMutex{},
	}}
}

func (d *depositRepositoryImpl) AddDeposits(
	ctx context.Context, deposits []domain.Deposit,
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

func (d *depositRepositoryImpl) GetDepositsForAccount(
	ctx context.Context, accountName string, page domain.Page,
) ([]domain.Deposit, error) {
	d.store.locker.RLock()
	defer d.store.locker.RUnlock()

	deposits := make([]domain.Deposit, 0)
	for _, deposit := range d.store.deposits {
		if deposit.AccountName == accountName {
			deposits = append(deposits, deposit)
		}
	}
	sort.SliceStable(deposits, func(i, j int) bool {
		return deposits[i].Timestamp > deposits[j].Timestamp
	})

	if page == nil {
		return deposits, nil
	}

	pagDeposits := make([]domain.Deposit, 0, page.GetSize())
	startIndex := int(page.GetNumber()*page.GetSize() - page.GetSize())
	endIndex := int(page.GetNumber() * page.GetSize())
	for i, deposit := range deposits {
		if i >= startIndex && i < endIndex {
			pagDeposits = append(pagDeposits, deposit)
		}
	}
	return pagDeposits, nil
}

func (d *depositRepositoryImpl) GetAllDeposits(
	ctx context.Context, page domain.Page,
) ([]domain.Deposit, error) {
	d.store.locker.RLock()
	defer d.store.locker.RUnlock()

	deposits := make([]domain.Deposit, 0)
	for _, deposit := range d.store.deposits {
		deposits = append(deposits, deposit)
	}
	sort.SliceStable(deposits, func(i, j int) bool {
		return deposits[i].Timestamp > deposits[j].Timestamp
	})

	if page == nil {
		return deposits, nil
	}

	pagDeposits := make([]domain.Deposit, 0, page.GetSize())
	startIndex := int(page.GetNumber()*page.GetSize() - page.GetSize())
	endIndex := int(page.GetNumber() * page.GetSize())
	for i, deposit := range deposits {
		if i >= startIndex && i < endIndex {
			pagDeposits = append(pagDeposits, deposit)
		}
	}
	return pagDeposits, nil
}
