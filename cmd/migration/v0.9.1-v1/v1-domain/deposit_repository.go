package v1domain

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v4"
)

type DepositRepository interface {
	InsertDeposits(deposits []*domain.Deposit) error
}

type depositRepositoryImpl struct {
	store *badgerhold.Store
}

func NewDepositRepositoryImpl(store *badgerhold.Store) DepositRepository {
	return &depositRepositoryImpl{store}
}

func (d *depositRepositoryImpl) InsertDeposits(deposits []*domain.Deposit) error {
	for _, v := range deposits {
		if err := d.store.Insert(v.Key(), *v); err != nil {
			return err
		}
	}

	return nil
}
