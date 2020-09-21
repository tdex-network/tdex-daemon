package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"github.com/tdex-network/tdex-daemon/internal/storageutil/uow"
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
func (r InMemoryUnspentRepository) AddUnspents(ctx context.Context, unspents []unspent.Unspent) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return addUnspents(r.storageByContext(ctx), unspents)
}

// GetAllUnspents returns all the unspents stored
func (r InMemoryUnspentRepository) GetAllUnspents(ctx context.Context) []unspent.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getAllUnspents(r.storageByContext(ctx), false)
}

// GetAllSpents returns all the unspents that have been spent
func (r InMemoryUnspentRepository) GetAllSpents(ctx context.Context) []unspent.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getAllUnspents(r.storageByContext(ctx), true)
}

// GetBalance returns the balance of the given asset for the given address
func (r InMemoryUnspentRepository) GetBalance(
	ctx context.Context,
	address, assetHash string,
) uint64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getBalance(r.storageByContext(ctx), address, assetHash)
}

// GetUnlockedBalance returns the total amount of unlocked unspents for the
// given asset and address
func (r InMemoryUnspentRepository) GetUnlockedBalance(
	ctx context.Context,
	address, assetHash string,
) uint64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getUnlockedBalance(r.storageByContext(ctx), address, assetHash)
}

// GetAvailableUnspents returns the list of unlocked unspents
func (r InMemoryUnspentRepository) GetAvailableUnspents(ctx context.Context) []unspent.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getAvailableUnspents(r.storageByContext(ctx))
}

// LockUnspents locks the given unspents associating them with the trade where
// they'are currently used as inputs
func (r InMemoryUnspentRepository) LockUnspents(
	ctx context.Context,
	unspentKeys []unspent.UnspentKey,
	tradeID uuid.UUID,
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return lockUnspents(r.storageByContext(ctx), unspentKeys, tradeID)
}

// UnlockUnspents unlocks the given locked unspents
func (r InMemoryUnspentRepository) UnlockUnspents(
	ctx context.Context,
	unspentKeys []unspent.UnspentKey,
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return unlockUnspents(r.storageByContext(ctx), unspentKeys)
}

// Begin returns a new InMemoryUnspentRepositoryTx
func (r InMemoryUnspentRepository) Begin() (*InMemoryUnspentRepositoryTx, error) {
	tx := &InMemoryUnspentRepositoryTx{
		root:     r,
		unspents: map[unspent.UnspentKey]unspent.Unspent{},
	}

	// copy the current state of the repo into the transaction
	for k, v := range r.unspents {
		tx.unspents[k] = v
	}
	return tx, nil
}

// ContextKey returns the context key shared between in-memory repositories
func (r InMemoryUnspentRepository) ContextKey() interface{} {
	return uow.InMemoryContextKey
}

func (r InMemoryUnspentRepository) storageByContext(ctx context.Context) (
	unspents map[unspent.UnspentKey]unspent.Unspent,
) {
	unspents = r.unspents
	if tx, ok := ctx.Value(r).(*InMemoryUnspentRepositoryTx); ok {
		unspents = tx.unspents
	}
	return
}

func addUnspents(storage map[unspent.UnspentKey]unspent.Unspent, unspents []unspent.Unspent) error {
	addr := unspents[0].Address()
	for _, u := range unspents {
		if u.Address() != addr {
			return fmt.Errorf("all unspent's must belong to the same address")
		}
	}

	//add new unspent
	for _, newUnspent := range unspents {
		if _, ok := storage[newUnspent.GetKey()]; !ok {
			storage[unspent.UnspentKey{
				TxID: newUnspent.TxID(),
				VOut: newUnspent.VOut(),
			}] = newUnspent
		}
	}

	//update spent
	for key, oldUnspent := range storage {
		if oldUnspent.Address() == addr {
			exist := false
			for _, newUnspent := range unspents {
				if newUnspent.IsKeyEqual(oldUnspent.GetKey()) {
					exist = true
				}
			}
			if !exist {
				oldUnspent.Spend()
				storage[key] = oldUnspent
			}
		}
	}

	return nil
}

func getAllUnspents(storage map[unspent.UnspentKey]unspent.Unspent, spent bool) []unspent.Unspent {
	unspents := make([]unspent.Unspent, 0)
	for _, u := range storage {
		if u.IsSpent() == spent {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

func getBalance(storage map[unspent.UnspentKey]unspent.Unspent, address, assetHash string) uint64 {
	var balance uint64
	for _, u := range storage {
		if u.Address() == address && u.AssetHash() == assetHash && !u.IsSpent() {
			balance += u.Value()
		}
	}
	return balance
}

func getUnlockedBalance(storage map[unspent.UnspentKey]unspent.Unspent, address, assetHash string) uint64 {
	var balance uint64
	for _, u := range storage {
		if u.Address() == address && u.AssetHash() == assetHash &&
			!u.IsSpent() && !u.IsLocked() {
			balance += u.Value()
		}
	}
	return balance
}

func getAvailableUnspents(storage map[unspent.UnspentKey]unspent.Unspent) []unspent.Unspent {
	unspents := make([]unspent.Unspent, 0)
	for _, u := range storage {
		if u.IsSpent() == false && u.IsLocked() == false {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

func lockUnspents(
	storage map[unspent.UnspentKey]unspent.Unspent,
	unspentKeys []unspent.UnspentKey,
	tradeID uuid.UUID,
) error {
	for _, key := range unspentKeys {
		if _, ok := storage[key]; !ok {
			return fmt.Errorf("unspent not found for key %v", key)
		}
	}

	for _, key := range unspentKeys {
		unspent := storage[key]
		unspent.Lock(&tradeID)
		storage[key] = unspent
	}
	return nil
}

func unlockUnspents(
	storage map[unspent.UnspentKey]unspent.Unspent,
	unspentKeys []unspent.UnspentKey,
) error {
	for _, key := range unspentKeys {
		if _, ok := storage[key]; !ok {
			return fmt.Errorf("unspent not found for key %v", key)
		}
	}

	for _, key := range unspentKeys {
		unspent := storage[key]
		unspent.UnLock()
		storage[key] = unspent
	}
	return nil
}

// InMemoryUnspentRepositoryTx allows to make transactional read/write operation
// on the in-memory repository
type InMemoryUnspentRepositoryTx struct {
	root     InMemoryUnspentRepository
	unspents map[unspent.UnspentKey]unspent.Unspent
}

// Commit applies the updates made to the state of the transaction to its root
func (tx *InMemoryUnspentRepositoryTx) Commit() error {
	for k, v := range tx.unspents {
		tx.root.unspents[k] = v
	}
	return nil
}

// Rollback resets the state of the transaction to the state of its root
func (tx *InMemoryUnspentRepositoryTx) Rollback() error {
	tx.unspents = map[unspent.UnspentKey]unspent.Unspent{}
	for k, v := range tx.root.unspents {
		tx.unspents[k] = v
	}
	return nil
}
