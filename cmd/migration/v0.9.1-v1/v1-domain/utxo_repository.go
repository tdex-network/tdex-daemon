package v1domain

import (
	"github.com/timshannon/badgerhold/v4"
)

type UtxoRepository interface {
	InsertUtxos(utxos []*Utxo) error
}

type utxoRepositoryImpl struct {
	store *badgerhold.Store
}

func NewUtxoRepositoryImpl(store *badgerhold.Store) UtxoRepository {
	return &utxoRepositoryImpl{store}
}

func (u *utxoRepositoryImpl) InsertUtxos(utxos []*Utxo) error {
	for _, v := range utxos {
		if err := u.store.Insert(v.Key().Hash(), *v); err != nil {
			return err
		}
	}

	return nil
}
