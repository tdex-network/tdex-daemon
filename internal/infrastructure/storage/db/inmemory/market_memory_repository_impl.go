package inmemory

import (
	"context"
	"errors"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"sort"
	"sync"
)

var (
	// ErrMarketNotExist is thrown when a market is not found
	ErrMarketNotExist = errors.New("market does not exists")
)

// MarketRepositoryImpl represents an in memory storage
type MarketRepositoryImpl struct {
	markets         map[int]domain.Market
	accountsByAsset map[string]int

	lock *sync.RWMutex
}

//NewMemoryMarketRepository returns a new empty MarketRepositoryImpl
func NewMarketRepositoryImpl() *MarketRepositoryImpl {
	return &MarketRepositoryImpl{
		markets:         map[int]domain.Market{},
		accountsByAsset: map[string]int{},
		lock:            &sync.RWMutex{},
	}
}

//GetOrCreateMarket gets a market with a given account index. If not found, a new entry is inserted
func (r MarketRepositoryImpl) GetOrCreateMarket(_ context.Context, accountIndex int) (market *domain.Market, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getOrCreateMarket(accountIndex)
}

//GetMarketByAsset returns a funded market using the quote asset hash
func (r MarketRepositoryImpl) GetMarketByAsset(_ context.Context, quoteAsset string) (market *domain.Market, accountIndex int, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getMarketByAsset(quoteAsset)
}

//GetLatestMarket returns the latest stored market (either funded or not)
func (r MarketRepositoryImpl) GetLatestMarket(_ context.Context) (market *domain.Market, accountIndex int, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getLatestMarket()
}

//GetTradableMarkets returns all the markets available for trading
func (r MarketRepositoryImpl) GetTradableMarkets(_ context.Context) (tradableMarkets []domain.Market, err error) {
	for _, mkt := range r.markets {
		if mkt.IsTradable() {
			tradableMarkets = append(tradableMarkets, mkt)
		}
	}

	return tradableMarkets, nil
}

// GetAllMarkets returns all the markets either tradable or not.
func (r MarketRepositoryImpl) GetAllMarkets(_ context.Context) (
	markets []domain.Market, err error,
) {
	for _, mkt := range r.markets {
		markets = append(markets, mkt)
	}

	return
}

//UpdateMarket updates data to a market identified by the account index passing an update function
func (r MarketRepositoryImpl) UpdateMarket(
	_ context.Context,
	accountIndex int,
	updateFn func(m *domain.Market) (*domain.Market, error),
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
		r.accountsByAsset[updatedMarket.QuoteAsset] = accountIndex
	}

	return nil
}

// OpenMarket makes a market found with the given quote asset hash as available for trading
func (r MarketRepositoryImpl) OpenMarket(_ context.Context, quoteAsset string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	currentMarket, accountIndex, err := r.getMarketByAsset(quoteAsset)
	if err != nil {
		return err
	}

	// We update the market status only if the market is closed.
	if currentMarket.IsTradable() {
		return nil
	}

	err = currentMarket.MakeTradable()
	if err != nil {
		return err
	}

	r.markets[accountIndex] = *currentMarket

	return nil
}

// CloseMarket makes a market found with the given quote asset hash as NOT available for trading
func (r MarketRepositoryImpl) CloseMarket(_ context.Context, quoteAsset string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	currentMarket, accountIndex, err := r.getMarketByAsset(quoteAsset)
	if err != nil {
		return err
	}

	// We update the market status only if the market is open.
	if !currentMarket.IsTradable() {
		return nil
	}

	err = currentMarket.MakeNotTradable()
	if err != nil {
		return err
	}

	r.markets[accountIndex] = *currentMarket

	return nil
}

func (r MarketRepositoryImpl) getOrCreateMarket(accountIndex int) (*domain.Market, error) {
	currentMarket, ok := r.markets[accountIndex]
	if !ok {
		return domain.NewMarket(accountIndex)
	}

	return &currentMarket, nil
}

func (r MarketRepositoryImpl) getMarketByAsset(quoteAsset string) (*domain.Market, int, error) {
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

func (r MarketRepositoryImpl) getLatestMarket() (*domain.Market, int, error) {
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
