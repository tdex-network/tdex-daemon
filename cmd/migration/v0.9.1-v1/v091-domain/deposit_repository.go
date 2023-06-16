package v091domain

import "github.com/sekulicd/badgerhold/v2"

type DepositRepository interface {
	GetAllDeposits() ([]*Deposit, error)
}

type depositRepositoryImpl struct {
	mainDb *badgerhold.Store
}

func NewDepositRepositoryImpl(mainDb *badgerhold.Store) DepositRepository {
	return &depositRepositoryImpl{
		mainDb: mainDb,
	}
}

func (d *depositRepositoryImpl) GetAllDeposits() ([]*Deposit, error) {
	var deposits []Deposit
	if err := d.mainDb.Find(&deposits, nil); err != nil {
		return nil, err
	}

	res := make([]*Deposit, 0, len(deposits))
	for i := range deposits {
		d := deposits[i]
		res = append(res, &d)
	}

	return res, nil
}
