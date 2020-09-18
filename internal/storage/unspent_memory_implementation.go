package storage

import (
	"errors"
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"sync"
)

// InMemoryUnspentRepository represents an in memory storage
type InMemoryUnspentRepository struct {
	unspents map[unspent.UnspentKey]unspent.Unspent
	lock     *sync.RWMutex
}

//NewInMemoryUnspentRepository returns a new empty InMemoryMarketRepository
func NewInMemoryUnspentRepository() *InMemoryUnspentRepository {
	return &InMemoryUnspentRepository{
		unspents: make(map[unspent.UnspentKey]unspent.Unspent),
		lock:     &sync.RWMutex{},
	}
}

//AddUnspent method is used by crawler to add unspent's to the memory,
//it assumes that all unspent's belongs to the same address,
//it assumes that each time it is invoked by crawler,
//it assumes that it will receive all unspent's for specific address
//it adds non exiting unspent's to the memory
//in case that unspent's, passed to the function, are not already in memory
//it will mark unspent in memory, as spent
func (i *InMemoryUnspentRepository) AddUnspent(unspents []unspent.Unspent) error {

	addr := unspents[0].Address()
	for _, u := range unspents {
		if u.Address() != addr {
			return errors.New("all unspent's must belong to the same address")
		}
	}

	//add new unspent
	for _, newUnspent := range unspents {
		if _, ok := i.unspents[newUnspent.GetKey()]; !ok {
			i.unspents[unspent.UnspentKey{
				TxID: newUnspent.TxID(),
				VOut: newUnspent.VOut(),
			}] = newUnspent
		}
	}

	//update spent
	for key, oldUnspent := range i.unspents {
		if oldUnspent.Address() == addr {
			exist := false
			for _, newUnspent := range unspents {
				if newUnspent.IsKeyEqual(oldUnspent.GetKey()) {
					exist = true
				}
			}
			if !exist {
				oldUnspent.Spend()
				i.unspents[key] = oldUnspent
			}
		}
	}

	return nil
}

func (i *InMemoryUnspentRepository) GetAllUnspent() []unspent.Unspent {
	unspents := make([]unspent.Unspent, 0)
	for _, u := range i.unspents {
		if u.IsSpent() == false {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

func (i *InMemoryUnspentRepository) GetAllSpent() []unspent.Unspent {
	unspents := make([]unspent.Unspent, 0)
	for _, u := range i.unspents {
		if u.IsSpent() == true {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

func (i *InMemoryUnspentRepository) GetBalance(
	address string,
	assetHash string,
) uint64 {
	var balance uint64

	for _, u := range i.unspents {
		if u.Address() == address && u.AssetHash() == assetHash && !u.IsSpent() {
			balance += u.Value()
		}
	}

	return balance
}

func (i *InMemoryUnspentRepository) GetUnlockedBalance(
	address string,
	assetHash string,
) uint64 {
	var balance uint64

	for _, u := range i.unspents {
		if u.Address() == address && u.AssetHash() == assetHash &&
			!u.IsSpent() && !u.IsLocked() {
			balance += u.Value()
		}
	}

	return balance
}

func (i *InMemoryUnspentRepository) GetAvailableUnspent() []unspent.Unspent {
	unspents := make([]unspent.Unspent, 0)
	for _, u := range i.unspents {
		if u.IsSpent() == false && u.IsLocked() == false {
			unspents = append(unspents, u)
		}
	}
	return unspents
}
