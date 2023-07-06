package v1domain

import (
	"github.com/timshannon/badgerhold/v4"
)

type TransactionRepository interface {
	AddTransactions(txs []*Transaction) error
}

type txRepositoryImpl struct {
	store *badgerhold.Store
}

func NewTransactionRepositoryImpl(store *badgerhold.Store) TransactionRepository {
	return &txRepositoryImpl{store}
}

func (u *txRepositoryImpl) AddTransactions(txs []*Transaction) error {
	for _, tx := range txs {
		if err := u.store.Insert(tx.TxID, tx); err != nil {
			return err
		}
	}

	return nil
}
