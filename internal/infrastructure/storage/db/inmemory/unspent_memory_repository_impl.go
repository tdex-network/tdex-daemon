package inmemory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

// UnspentRepositoryImpl represents an in memory storage
type UnspentRepositoryImpl struct {
	db *DbManager
}

//NewUnspentRepositoryImpl returns a new empty MarketRepositoryImpl
func NewUnspentRepositoryImpl(db *DbManager) domain.UnspentRepository {
	return &UnspentRepositoryImpl{
		db: db,
	}
}

//AddUnspents method is used by crawler to add unspent's to the memory,
//it assumes that all unspent's belongs to the same address,
//it assumes that each time it is invoked by crawler,
//it assumes that it will receive all unspent's for specific address
//it adds non exiting unspent's to the memory
//in case that unspent's, passed to the function, are not already in memory
//it will mark unspent in memory, as spent
func (r UnspentRepositoryImpl) AddUnspents(_ context.Context, unspents []domain.Unspent) error {
	r.db.unspentStore.locker.Lock()
	defer r.db.unspentStore.locker.Unlock()

	return r.addUnspents(unspents)
}

// GetAllUnspents returns all the unspents stored
func (r UnspentRepositoryImpl) GetAllUnspents(_ context.Context) []domain.Unspent {
	r.db.unspentStore.locker.RLock()
	defer r.db.unspentStore.locker.RUnlock()

	includeSpent := true
	return r.getAllUnspents(includeSpent)
}

// GetAvailableUnspents ...
func (r UnspentRepositoryImpl) GetAvailableUnspents(_ context.Context) ([]domain.Unspent, error) {
	r.db.unspentStore.locker.RLock()
	defer r.db.unspentStore.locker.RUnlock()

	return r.getAvailableUnspents(nil), nil
}

// GetAllUnspentsForAddresses ...
func (r UnspentRepositoryImpl) GetAllUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	r.db.unspentStore.locker.Lock()
	defer r.db.unspentStore.locker.Unlock()

	results := make([]domain.Unspent, 0)

	for _, unspent := range r.db.unspentStore.unspents {
		unspentAddress := unspent.Address
		for _, addr := range addresses {
			if unspentAddress == addr {
				results = append(results, unspent)
			}
		}
	}

	return results, nil
}

// GetUnspentsForAddresses ...
func (r UnspentRepositoryImpl) GetUnspentsForAddresses(
	_ context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	r.db.unspentStore.locker.RLock()
	defer r.db.unspentStore.locker.RUnlock()

	return r.getUnspentsForAddresses(addresses), nil
}

// GetAvailableUnspentsForAddresses ...
func (r UnspentRepositoryImpl) GetAvailableUnspentsForAddresses(
	_ context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	r.db.unspentStore.locker.RLock()
	defer r.db.unspentStore.locker.RUnlock()

	return r.getAvailableUnspents(addresses), nil
}

// GetUnspentWithKey ...
func (r UnspentRepositoryImpl) GetUnspentWithKey(
	_ context.Context,
	unspentKey domain.UnspentKey,
) (*domain.Unspent, error) {
	r.db.unspentStore.locker.RLock()
	defer r.db.unspentStore.locker.RUnlock()

	unspent, ok := r.db.unspentStore.unspents[unspentKey]
	if !ok {
		return nil, nil
	}
	return &unspent, nil
}

// GetBalance ...
func (r UnspentRepositoryImpl) GetBalance(
	_ context.Context,
	addresses []string,
	assetHash string,
) (uint64, error) {
	balance := uint64(0)

	r.db.unspentStore.locker.RLock()
	defer r.db.unspentStore.locker.RUnlock()

	for _, address := range addresses {
		balance += r.getBalance(address, assetHash)
	}

	return balance, nil
}

// GetUnlockedBalance ...
func (r UnspentRepositoryImpl) GetUnlockedBalance(
	_ context.Context,
	addresses []string,
	assetHash string,
) (uint64, error) {
	balance := uint64(0)

	r.db.unspentStore.locker.RLock()
	defer r.db.unspentStore.locker.RUnlock()

	for _, address := range addresses {
		balance += r.getUnlockedBalance(address, assetHash)
	}

	return balance, nil
}

// SpendUnspents ...
func (r UnspentRepositoryImpl) SpendUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) error {
	r.db.unspentStore.locker.Lock()
	defer r.db.unspentStore.locker.Unlock()

	for _, key := range unspentKeys {
		if unspent, ok := r.db.unspentStore.unspents[key]; ok {
			if !unspent.IsSpent() {
				unspent.Spend()
				unspent.Unlock()
				r.db.unspentStore.unspents[key] = unspent
			}
		}
	}

	return nil
}

// ConfirmUnspents ...
func (r UnspentRepositoryImpl) ConfirmUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) error {
	r.db.unspentStore.locker.Lock()
	defer r.db.unspentStore.locker.Unlock()

	for _, key := range unspentKeys {
		if unspent, ok := r.db.unspentStore.unspents[key]; ok {
			if !unspent.IsConfirmed() {
				unspent.Confirm()
				unspent.Unlock()
				r.db.unspentStore.unspents[key] = unspent
			}
		}
	}

	return nil
}

// LockUnspents ...
func (r UnspentRepositoryImpl) LockUnspents(
	_ context.Context,
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) error {
	r.db.unspentStore.locker.Lock()
	defer r.db.unspentStore.locker.Unlock()

	go func() {
		time.Sleep(3 * time.Second)
		r.db.unspentStore.locker.Lock()
		defer r.db.unspentStore.locker.Unlock()
		for _, key := range unspentKeys {
			unspent := r.db.unspentStore.unspents[key]
			unspent.Unlock()
			r.db.unspentStore.unspents[key] = unspent
		}
	}()

	return r.lockUnspents(unspentKeys, tradeID)
}

// UnlockUnspents ...
func (r UnspentRepositoryImpl) UnlockUnspents(
	_ context.Context,
	unspentKeys []domain.UnspentKey,
) error {
	r.db.unspentStore.locker.Lock()
	defer r.db.unspentStore.locker.Unlock()

	return r.unlockUnspents(unspentKeys)
}

func (r UnspentRepositoryImpl) addUnspents(unspents []domain.Unspent) error {
	//add new unspent
	for _, newUnspent := range unspents {
		if _, ok := r.db.unspentStore.unspents[newUnspent.Key()]; !ok {
			r.db.unspentStore.unspents[domain.UnspentKey{
				TxID: newUnspent.TxID,
				VOut: newUnspent.VOut,
			}] = newUnspent
		}
	}

	//update spent
	for key, oldUnspent := range r.db.unspentStore.unspents {
		exist := false
		for _, newUnspent := range unspents {
			if newUnspent.IsKeyEqual(oldUnspent.Key()) {
				exist = true
			}
		}
		if !exist {
			r.db.unspentStore.unspents[key] = oldUnspent
		}
	}

	return nil
}

func (r UnspentRepositoryImpl) getAllUnspents(includeSpent bool) []domain.Unspent {
	unspents := make([]domain.Unspent, 0)
	for _, u := range r.db.unspentStore.unspents {
		if includeSpent {
			unspents = append(unspents, u)
			continue
		}

		if !u.IsSpent() {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

func (r UnspentRepositoryImpl) getBalance(address string, assetHash string) uint64 {
	var balance uint64
	for _, u := range r.db.unspentStore.unspents {
		if u.Address == address && u.AssetHash == assetHash && !u.IsSpent() && u.IsConfirmed() {
			balance += u.Value
		}
	}
	return balance
}

func (r UnspentRepositoryImpl) getUnlockedBalance(address string, assetHash string) uint64 {
	var balance uint64

	for _, u := range r.db.unspentStore.unspents {
		if u.Address == address && u.AssetHash == assetHash &&
			!u.IsSpent() && !u.IsLocked() && u.IsConfirmed() {
			balance += u.Value
		}
	}

	return balance
}

func (r UnspentRepositoryImpl) getUnspentsForAddresses(addresses []string) []domain.Unspent {
	unspentsUnlocked := r.getUnspents(addresses, false)
	unspentsLocked := r.getUnspents(addresses, true)
	return append(unspentsUnlocked, unspentsLocked...)
}

func (r UnspentRepositoryImpl) getAvailableUnspents(addresses []string) []domain.Unspent {
	return r.getUnspents(addresses, false)
}

func (r UnspentRepositoryImpl) getUnspents(addresses []string, isLocked bool) []domain.Unspent {
	unspents := make([]domain.Unspent, 0)
	for _, u := range r.db.unspentStore.unspents {
		if !u.IsSpent() && u.IsLocked() == isLocked && u.IsConfirmed() {
			if len(addresses) == 0 {
				unspents = append(unspents, u)
			} else {
				for _, addr := range addresses {
					if addr == u.Address {
						unspents = append(unspents, u)
						break
					}
				}
			}
		}
	}
	return unspents
}

func (r UnspentRepositoryImpl) lockUnspents(unspentKeys []domain.UnspentKey, tradeID uuid.UUID) error {
	for _, key := range unspentKeys {
		if _, ok := r.db.unspentStore.unspents[key]; !ok {
			return fmt.Errorf("unspent not found for key %v", key)
		}
	}

	for _, key := range unspentKeys {
		unspent := r.db.unspentStore.unspents[key]
		unspent.Lock(&tradeID)
		r.db.unspentStore.unspents[key] = unspent
	}
	return nil
}

func (r UnspentRepositoryImpl) unlockUnspents(unspentKeys []domain.UnspentKey) error {
	for _, key := range unspentKeys {
		if _, ok := r.db.unspentStore.unspents[key]; !ok {
			return fmt.Errorf("unspent not found for key %v", key)
		}
	}

	for _, key := range unspentKeys {
		unspent := r.db.unspentStore.unspents[key]
		unspent.Unlock()
		r.db.unspentStore.unspents[key] = unspent
	}
	return nil
}
