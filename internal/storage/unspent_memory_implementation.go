package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
)

// InMemoryUnspentRepository represents an in memory storage
type InMemoryUnspentRepository struct {
	unspents map[unspent.UnspentKey]unspent.Unspent
	lock     *sync.RWMutex
}

//NewInMemoryUnspentRepository returns a new empty InMemoryMarketRepository
func NewInMemoryUnspentRepository() *InMemoryUnspentRepository {
	return &InMemoryUnspentRepository{
		unspents: map[unspent.UnspentKey]unspent.Unspent{},
		lock:     &sync.RWMutex{},
	}
}

//AddUnspents method is used by crawler to add unspent's to the memory,
//it assumes that all unspent's belongs to the same address,
//it assumes that each time it is invoked by crawler,
//it assumes that it will receive all unspent's for specific address
//it adds non exiting unspent's to the memory
//in case that unspent's, passed to the function, are not already in memory
//it will mark unspent in memory, as spent
func (r InMemoryUnspentRepository) AddUnspents(_ context.Context, unspents []unspent.Unspent) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	addr := unspents[0].Address()
	for _, u := range unspents {
		if u.Address() != addr {
			return fmt.Errorf("all unspent's must belong to the same address")
		}
	}

	//add new unspent
	for _, newUnspent := range unspents {
		if _, ok := r.unspents[newUnspent.GetKey()]; !ok {
			r.unspents[unspent.UnspentKey{
				TxID: newUnspent.TxID(),
				VOut: newUnspent.VOut(),
			}] = newUnspent
		}
	}

	//update spent
	for key, oldUnspent := range r.unspents {
		if oldUnspent.Address() == addr {
			exist := false
			for _, newUnspent := range unspents {
				if newUnspent.IsKeyEqual(oldUnspent.GetKey()) {
					exist = true
				}
			}
			if !exist {
				oldUnspent.Spend()
				r.unspents[key] = oldUnspent
			}
		}
	}

	return nil
}

// GetAllUnspents returns all the unspents stored
func (r InMemoryUnspentRepository) GetAllUnspents(_ context.Context) []unspent.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	unspents := make([]unspent.Unspent, 0)
	for _, u := range r.unspents {
		if u.IsSpent() == false {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

// GetAllSpents returns all the unspents that have been spent
func (r InMemoryUnspentRepository) GetAllSpents(_ context.Context) []unspent.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	unspents := make([]unspent.Unspent, 0)
	for _, u := range r.unspents {
		if u.IsSpent() == true {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

// GetBalance returns the balance of the given asset for the given address
func (r InMemoryUnspentRepository) GetBalance(
	_ context.Context,
	address string,
	assetHash string,
) uint64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var balance uint64

	for _, u := range r.unspents {
		if u.Address() == address && u.AssetHash() == assetHash && !u.IsSpent() {
			balance += u.Value()
		}
	}

	return balance
}

// GetUnlockedBalance returns the total amount of unlocked unspents for the
// given asset and address
func (r InMemoryUnspentRepository) GetUnlockedBalance(
	_ context.Context,
	address string,
	assetHash string,
) uint64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var balance uint64

	for _, u := range r.unspents {
		if u.Address() == address && u.AssetHash() == assetHash &&
			!u.IsSpent() && !u.IsLocked() {
			balance += u.Value()
		}
	}

	return balance
}

// GetAvailableUnspents returns the list of unlocked unspents
func (r InMemoryUnspentRepository) GetAvailableUnspents(_ context.Context) []unspent.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	unspents := make([]unspent.Unspent, 0)
	for _, u := range r.unspents {
		if u.IsSpent() == false && u.IsLocked() == false {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

// LockUnspents locks the given unspents associating them with the trade where
// they'are currently used as inputs
func (r InMemoryUnspentRepository) LockUnspents(
	_ context.Context,
	unspentKeys []unspent.UnspentKey,
	tradeID uuid.UUID,
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, key := range unspentKeys {
		if _, ok := r.unspents[key]; !ok {
			return fmt.Errorf("unspent not found for key %v", key)
		}
	}

	for _, key := range unspentKeys {
		unspent := r.unspents[key]
		unspent.Lock(&tradeID)
		r.unspents[key] = unspent
	}
	return nil
}

// UnlockUnspents unlocks the given locked unspents
func (r InMemoryUnspentRepository) UnlockUnspents(
	_ context.Context,
	unspentKeys []unspent.UnspentKey,
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, key := range unspentKeys {
		if _, ok := r.unspents[key]; !ok {
			return fmt.Errorf("unspent not found for key %v", key)
		}
	}

	for _, key := range unspentKeys {
		unspent := r.unspents[key]
		unspent.UnLock()
		r.unspents[key] = unspent
	}
	return nil
}
