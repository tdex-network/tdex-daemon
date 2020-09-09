package storage

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/domain/market"
)

var (
	// ErrMarketNotExist is thrown when a market is not found
	ErrMarketNotExist = errors.New("Market does not exists")
)

// InMemoryMarketRepository represents an in memory storage
type InMemoryMarketRepository struct {
	markets         map[int]market.Market
	accountsByAsset map[string]int

	lock *sync.RWMutex
}

//NewInMemoryMarketRepository returns a new empty InMemoryMarketRepository
func NewInMemoryMarketRepository() *InMemoryMarketRepository {
	return &InMemoryMarketRepository{
		markets:         map[int]market.Market{},
		accountsByAsset: map[string]int{},
		lock:            &sync.RWMutex{},
	}
}

//GetOrCreateMarket gets a market with a given account index. If not found, a new entry is inserted
func (r InMemoryMarketRepository) GetOrCreateMarket(_ context.Context, accountIndex int) (market *market.Market, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getOrCreateMarket(accountIndex)
}

//GetMarketByAsset returns a funded market using the quote asset hash
func (r InMemoryMarketRepository) GetMarketByAsset(_ context.Context, quoteAsset string) (market *market.Market, accountIndex int, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getMarketByAsset(quoteAsset)
}

//GetLatestMarket returns the latest stored market (either funded or not)
func (r InMemoryMarketRepository) GetLatestMarket(_ context.Context) (market *market.Market, accountIndex int, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getLatestMarket()
}

//GetTradableMarkets returns all the markets available for trading
func (r InMemoryMarketRepository) GetTradableMarkets(_ context.Context) (tradableMarkets []market.Market, err error) {
	for _, mkt := range r.markets {
		if mkt.IsTradable() {
			tradableMarkets = append(tradableMarkets, mkt)
		}
	}

	return tradableMarkets, nil
}

// GetAllMarkets returns all the markets either tradable or not.
func (r InMemoryMarketRepository) GetAllMarkets(_ context.Context) (
	markets []market.Market, err error,
) {
	for _, mkt := range r.markets {
		markets = append(markets, mkt)
	}

	return
}

//UpdateMarket updates data to a market identified by the account index passing an update function
func (r InMemoryMarketRepository) UpdateMarket(
	_ context.Context,
	accountIndex int,
	updateFn func(m *market.Market) (*market.Market, error),
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	currentMarket, err := r.getOrCreateMarket(accountIndex)
	if err != nil {
		return err
	}

	updatedMarket, err := updateFn(currentMarket)
	if err != nil {
		return err
	}

	r.markets[accountIndex] = *updatedMarket
	if updatedMarket.IsFunded() {
		r.accountsByAsset[updatedMarket.QuoteAssetHash()] = accountIndex
	}

	return nil
}

// OpenMarket makes a market found with the given quote asset hash as available for trading
func (r InMemoryMarketRepository) OpenMarket(_ context.Context, quoteAsset string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	currentMarket, accountIndex, err := r.getMarketByAsset(quoteAsset)
	if err != nil {
		return err
	}

	err = currentMarket.MakeTradable()
	if err != nil {
		return err
	}

	r.markets[accountIndex] = *currentMarket

	return nil
}

// CloseMarket makes a market found with the given quote asset hash as NOT available for trading
func (r InMemoryMarketRepository) CloseMarket(_ context.Context, quoteAsset string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	currentMarket, accountIndex, err := r.getMarketByAsset(quoteAsset)
	if err != nil {
		return err
	}

	err = currentMarket.MakeNotTradable()
	if err != nil {
		return err
	}

	r.markets[accountIndex] = *currentMarket

	return nil
}

func (r InMemoryMarketRepository) getOrCreateMarket(accountIndex int) (*market.Market, error) {
	currentMarket, ok := r.markets[accountIndex]
	if !ok {
		return market.NewMarket(accountIndex)
	}

	return &currentMarket, nil
}

func (r InMemoryMarketRepository) getMarketByAsset(quoteAsset string) (*market.Market, int, error) {
	selectedAccountIndex, assetExist := r.accountsByAsset[quoteAsset]
	if !assetExist {
		return nil, -1, ErrMarketNotExist
	}
	currentMarket, ok := r.markets[selectedAccountIndex]
	if !ok {
		return nil, -1, ErrMarketNotExist
	}
	return &currentMarket, selectedAccountIndex, nil
}

func (r InMemoryMarketRepository) getLatestMarket() (*market.Market, int, error) {
	// In case we never created any markets yet, first account index usable should be the 5th,
	// becuase accounts 0-4 are reserved for other internal daemon purposes.
	// We returns 4th account index as the latest, so other code will increment and does not need to know of this reserved thing.
	//
	// TODO move in separated constant type mapping
	if len(r.markets) == 0 {
		return nil, 4, nil
	}

	accountIndexes := make([]int, 0, len(r.markets))
	for k := range r.markets {
		accountIndexes = append(accountIndexes, k)
	}

	sort.Ints(accountIndexes)

	latestAccountIndex := accountIndexes[len(accountIndexes)-1]

	currentMarket, ok := r.markets[latestAccountIndex]
	if !ok {
		return nil, -1, ErrMarketNotExist
	}

	return &currentMarket, latestAccountIndex, nil
}
