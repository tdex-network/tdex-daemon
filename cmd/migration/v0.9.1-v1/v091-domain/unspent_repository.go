package v091domain

import (
	"github.com/google/uuid"
	"github.com/sekulicd/badgerhold/v2"
)

const (
	LockedUnspentBadgerholdKeyPrefix = "bh_LockedUnspent"
)

type UnspentRepository interface {
	GetAllUnspents() ([]*Unspent, error)
}

type unspentRepositoryImpl struct {
	store     *badgerhold.Store
	lockStore *badgerhold.Store
}

func NewUnspentRepositoryImpl(
	store, lockStore *badgerhold.Store,
) UnspentRepository {
	return &unspentRepositoryImpl{
		store:     store,
		lockStore: lockStore,
	}
}

func (u *unspentRepositoryImpl) GetAllUnspents() ([]*Unspent, error) {
	return u.findUnspents(nil, false)
}

func (u *unspentRepositoryImpl) findUnspents(
	query *badgerhold.Query,
	unlockedOnly bool,
) ([]*Unspent, error) {
	var unspents []Unspent
	var unlockedUnspents []Unspent
	if err := u.store.Find(&unspents, query); err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	for i, unspent := range unspents {
		tradeID, err := u.getLock(unspent.Key())
		if err != nil {
			return nil, err
		}
		if tradeID != nil {
			unspent.Lock(tradeID)
			unspents[i] = unspent
		}
		if unlockedOnly && !unspent.IsLocked() {
			unlockedUnspents = append(unlockedUnspents, unspent)
		}
	}

	if unlockedOnly {
		resp := make([]*Unspent, 0, len(unlockedUnspents))
		for i := range unlockedUnspents {
			un := unlockedUnspents[i]
			resp = append(resp, &un)
		}

		return resp, nil
	}

	resp := make([]*Unspent, 0, len(unspents))
	for i := range unspents {
		un := unspents[i]
		resp = append(resp, &un)
	}

	return resp, nil
}

type LockedUnspent struct {
	TradeID uuid.UUID
}

func (u *unspentRepositoryImpl) getLock(key UnspentKey) (*uuid.UUID, error) {
	var lockedUnspent LockedUnspent
	var err error

	encKey, err := EncodeKey(key, LockedUnspentBadgerholdKeyPrefix)
	if err != nil {
		return nil, err
	}

	if err := u.lockStore.Get(encKey, &lockedUnspent); err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &lockedUnspent.TradeID, nil
}

func EncodeKey(key interface{}, typeName string) ([]byte, error) {
	encoded, err := badgerhold.DefaultEncode(key)
	if err != nil {
		return nil, err
	}

	return append([]byte(typeName), encoded...), nil
}
