package inmemory

import (
	"context"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

// UnspentRepositoryImpl represents an in memory storage
type UnspentRepositoryImpl struct {
	store *unspentInmemoryStore
}

//NewUnspentRepositoryImpl returns a new empty MarketRepositoryImpl
func NewUnspentRepositoryImpl(store *unspentInmemoryStore) domain.UnspentRepository {
	return &UnspentRepositoryImpl{store}
}

//AddUnspents method is used by crawler to add unspent's to the memory,
//it assumes that all unspent's belongs to the same address,
//it assumes that each time it is invoked by crawler,
//it assumes that it will receive all unspent's for specific address
//it adds non exiting unspent's to the memory
//in case that unspent's, passed to the function, are not already in memory
//it will mark unspent in memory, as spent
func (r UnspentRepositoryImpl) AddUnspents(
	_ context.Context, unspents []domain.Unspent,
) (int, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.addUnspents(unspents)
}

// GetAllUnspents returns all the unspents stored
func (r UnspentRepositoryImpl) GetAllUnspents(_ context.Context) []domain.Unspent {
	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

	includeSpent := true
	return r.getAllUnspents(includeSpent)
}

// GetAvailableUnspents ...
func (r UnspentRepositoryImpl) GetAvailableUnspents(_ context.Context) ([]domain.Unspent, error) {
	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

	return r.getAvailableUnspents(nil), nil
}

// GetAllUnspentsForAddresses ...
func (r UnspentRepositoryImpl) GetAllUnspentsForAddresses(
	ctx context.Context, addresses []string,
) ([]domain.Unspent, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	results := make([]domain.Unspent, 0)
	for _, unspent := range r.store.unspents {
		unspentAddress := unspent.Address
		for _, addr := range addresses {
			if unspentAddress == addr {
				results = append(results, unspent)
			}
		}
	}

	return results, nil
}

// GetAllUnspentsForAddressesAndPage ...
func (r UnspentRepositoryImpl) GetAllUnspentsForAddressesAndPage(
	ctx context.Context, addresses []string, page domain.Page,
) ([]domain.Unspent, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	results := make([]domain.Unspent, 0)
	startIndex := page.Number*page.Size - page.Size + 1
	endIndex := page.Number * page.Size
	index := 1
	for _, unspent := range r.store.unspents {
		unspentAddress := unspent.Address
		for _, addr := range addresses {
			if unspentAddress == addr {
				if index >= startIndex && index <= endIndex {
					results = append(results, unspent)
				}
				index++
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
	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

	return r.getUnspentsForAddresses(addresses), nil
}

// GetAvailableUnspentsForAddresses ...
func (r UnspentRepositoryImpl) GetAvailableUnspentsForAddresses(
	_ context.Context,
	addresses []string,
) ([]domain.Unspent, error) {
	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

	return r.getAvailableUnspents(addresses), nil
}

// GetBalance ...
func (r UnspentRepositoryImpl) GetBalance(
	_ context.Context,
	addresses []string,
	assetHash string,
) (uint64, error) {
	balance := uint64(0)

	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

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

	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

	for _, address := range addresses {
		balance += r.getUnlockedBalance(address, assetHash)
	}

	return balance, nil
}

// SpendUnspents ...
func (r UnspentRepositoryImpl) SpendUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	count := 0
	for _, key := range unspentKeys {
		if unspent, ok := r.store.unspents[key]; ok {
			if !unspent.IsSpent() {
				unspent.Spend()
				unspent.Unlock()
				r.store.unspents[key] = unspent
				count++
			}
		}
	}

	return count, nil
}

// ConfirmUnspents ...
func (r UnspentRepositoryImpl) ConfirmUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	count := 0
	for _, key := range unspentKeys {
		if unspent, ok := r.store.unspents[key]; ok {
			if !unspent.IsConfirmed() {
				unspent.Confirm()
				unspent.Unlock()
				r.store.unspents[key] = unspent
				count++
			}
		}
	}

	return count, nil
}

// LockUnspents ...
func (r UnspentRepositoryImpl) LockUnspents(
	_ context.Context,
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) (int, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.lockUnspents(unspentKeys, tradeID)
}

// UnlockUnspents ...
func (r UnspentRepositoryImpl) UnlockUnspents(
	_ context.Context,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.unlockUnspents(unspentKeys)
}

func (r UnspentRepositoryImpl) addUnspents(unspents []domain.Unspent) (int, error) {
	count := 0
	for _, newUnspent := range unspents {
		if _, ok := r.store.unspents[newUnspent.Key()]; !ok {
			r.store.unspents[domain.UnspentKey{
				TxID: newUnspent.TxID,
				VOut: newUnspent.VOut,
			}] = newUnspent
			count++
		}
	}
	return count, nil
}

func (r UnspentRepositoryImpl) getAllUnspents(includeSpent bool) []domain.Unspent {
	unspents := make([]domain.Unspent, 0)
	for _, u := range r.store.unspents {
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
	for _, u := range r.store.unspents {
		if u.Address == address && u.AssetHash == assetHash && !u.IsSpent() && u.IsConfirmed() {
			balance += u.Value
		}
	}
	return balance
}

func (r UnspentRepositoryImpl) getUnlockedBalance(address string, assetHash string) uint64 {
	var balance uint64

	for _, u := range r.store.unspents {
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
	for _, u := range r.store.unspents {
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

func (r UnspentRepositoryImpl) lockUnspents(
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) (int, error) {
	count := 0
	for _, key := range unspentKeys {
		if unspent, ok := r.store.unspents[key]; ok {
			unspent.Lock(&tradeID)
			r.store.unspents[key] = unspent
			count++
		}
	}
	return count, nil
}

func (r UnspentRepositoryImpl) unlockUnspents(
	unspentKeys []domain.UnspentKey,
) (int, error) {
	count := 0
	for _, key := range unspentKeys {
		if unspent, ok := r.store.unspents[key]; ok {
			unspent.Unlock()
			r.store.unspents[key] = unspent
			count++
		}
	}
	return count, nil
}
