package v091domain

import "github.com/sekulicd/badgerhold/v2"

type WithdrawalRepository interface {
	GetAllWithdrawals() ([]*Withdrawal, error)
}

type withdrawalRepositoryImpl struct {
	mainDb *badgerhold.Store
}

func NewWithdrawalRepositoryImpl(mainDb *badgerhold.Store) WithdrawalRepository {
	return &withdrawalRepositoryImpl{
		mainDb: mainDb,
	}
}

func (w *withdrawalRepositoryImpl) GetAllWithdrawals() ([]*Withdrawal, error) {
	var withdrawals []Withdrawal
	if err := w.mainDb.Find(&withdrawals, nil); err != nil {
		return nil, err
	}

	res := make([]*Withdrawal, 0, len(withdrawals))
	for i := range withdrawals {
		wdls := withdrawals[i]
		res = append(res, &wdls)
	}

	return res, nil
}
