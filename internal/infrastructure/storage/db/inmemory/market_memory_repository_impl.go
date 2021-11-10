package inmemory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

// MarketRepositoryImpl represents an in memory storage
type MarketRepositoryImpl struct {
	store *marketInmemoryStore
}

// NewMarketRepositoryImpl returns a new empty MarketRepositoryImpl
func NewMarketRepositoryImpl(store *marketInmemoryStore) domain.MarketRepository {
	return &MarketRepositoryImpl{store}
}

// GetOrCreateMarket gets a market with a given account index. If not found, a new entry is inserted
func (r MarketRepositoryImpl) GetOrCreateMarket(_ context.Context, market *domain.Market) (*domain.Market, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getOrCreateMarket(market)
}

// GetMarketByAccount return the market for the account index given as parameter
func (r MarketRepositoryImpl) GetMarketByAccount(_ context.Context, accountIndex int) (*domain.Market, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getMarketByAccount(accountIndex)
}

// GetMarketByAsset returns a funded market using the quote asset hash
func (r MarketRepositoryImpl) GetMarketByAssets(_ context.Context, baseAsset, quoteAsset string) (market *domain.Market, accountIndex int, err error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getMarketByAssets(baseAsset, quoteAsset)
}

// GetLatestMarket returns the latest stored market (either funded or not)
func (r MarketRepositoryImpl) GetLatestMarket(_ context.Context) (market *domain.Market, accountIndex int, err error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getLatestMarket()
}

// GetTradableMarkets returns all the markets available for trading
func (r MarketRepositoryImpl) GetTradableMarkets(_ context.Context) (tradableMarkets []domain.Market, err error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	for _, mkt := range r.store.markets {
		if mkt.IsTradable() {
			tradableMarkets = append(tradableMarkets, mkt)
		}
	}

	return tradableMarkets, nil
}

// GetAllMarkets returns all the markets either tradable or not.
func (r MarketRepositoryImpl) GetAllMarkets(_ context.Context) ([]domain.Market, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	markets := make([]domain.Market, 0)

	for _, mkt := range r.store.markets {
		markets = append(markets, mkt)
	}

	return markets, nil
}

// UpdateMarket updates data to a market identified by the account index passing an update function
func (r MarketRepositoryImpl) UpdateMarket(
	_ context.Context,
	accountIndex int,
	updateFn func(m *domain.Market) (*domain.Market, error),
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	currentMarket, err := r.getMarketByAccount(accountIndex)
	if err != nil {
		return err
	}
	if currentMarket == nil {
		return ErrMarketNotFound
	}

	updatedMarket, err := updateFn(currentMarket)
	if err != nil {
		return err
	}

	key := keyFromAssets(updatedMarket.BaseAsset, updatedMarket.QuoteAsset)
	r.store.markets[accountIndex] = *updatedMarket
	r.store.accountsByAssetsKey[key] = accountIndex

	return nil
}

// OpenMarket makes a market found with the given quote asset hash as available for trading
func (r MarketRepositoryImpl) OpenMarket(_ context.Context, accountIndex int) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	currentMarket, err := r.getMarketByAccount(accountIndex)
	if err != nil {
		return err
	}
	if currentMarket == nil {
		return nil
	}

	// We update the market status only if the market is closed.
	if currentMarket.IsTradable() {
		return nil
	}

	err = currentMarket.MakeTradable()
	if err != nil {
		return err
	}

	r.store.markets[accountIndex] = *currentMarket

	return nil
}

// CloseMarket makes a market found with the given quote asset hash as NOT available for trading
func (r MarketRepositoryImpl) CloseMarket(_ context.Context, accountIndex int) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	currentMarket, err := r.getMarketByAccount(accountIndex)
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

	r.store.markets[accountIndex] = *currentMarket

	return nil
}

func (r *MarketRepositoryImpl) UpdatePrices(_ context.Context, accountIndex int, prices domain.Prices) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	market, err := r.getMarketByAccount(accountIndex)
	if err != nil {
		return err
	}

	err = market.ChangeBasePrice(prices.BasePrice)
	if err != nil {
		return err
	}
	err = market.ChangeQuotePrice(prices.QuotePrice)
	if err != nil {
		return err
	}

	r.store.markets[accountIndex] = *market
	return nil
}

func (r *MarketRepositoryImpl) DeleteMarket(
	_ context.Context,
	accountIndex int,
) error {
	delete(r.store.markets, accountIndex)

	return nil
}

func (r MarketRepositoryImpl) getOrCreateMarket(market *domain.Market) (*domain.Market, error) {
	if market == nil {
		return nil, ErrMarketInvalidRequest
	}

	// we can safely skip checking the error here because this functions never
	// returns one actually. The err variable is only needed to be able to
	// override mkt into the if statement.
	mkt, err := r.getMarketByAccount(market.AccountIndex)
	if err != nil {
		return nil, err
	}
	if mkt == nil {
		key := keyFromAssets(market.BaseAsset, market.QuoteAsset)
		r.store.markets[market.AccountIndex] = *market
		r.store.accountsByAssetsKey[key] = market.AccountIndex
		mkt = market
	}
	return mkt, nil
}

func (r MarketRepositoryImpl) getMarketByAccount(accountIndex int) (*domain.Market, error) {
	market, ok := r.store.markets[accountIndex]
	if !ok {
		return nil, nil
	}

	return &market, nil
}

func (r MarketRepositoryImpl) getMarketByAssets(baseAsset, quoteAsset string) (*domain.Market, int, error) {
	key := keyFromAssets(baseAsset, quoteAsset)
	selectedAccountIndex, assetExist := r.store.accountsByAssetsKey[key]
	if !assetExist {
		return nil, -1, nil
	}
	currentMarket, ok := r.store.markets[selectedAccountIndex]
	if !ok {
		return nil, -1, nil
	}
	return &currentMarket, selectedAccountIndex, nil
}

func (r MarketRepositoryImpl) getLatestMarket() (*domain.Market, int, error) {
	// In case we never created any markets yet, first account index usable should be the 5th,
	// becuase accounts 0-4 are reserved for other internal daemon purposes.
	// We returns 4th account index as the latest, so other code will increment and does not need to know of this reserved thing.
	//
	// TODO move in separated constant type mapping
	numberOfMarkets := len(r.store.markets)

	if numberOfMarkets == 0 {
		return nil, 4, nil
	}

	accountIndexes := make([]int, 0, numberOfMarkets)
	for k := range r.store.markets {
		accountIndexes = append(accountIndexes, k)
	}

	sort.Ints(accountIndexes)

	latestAccountIndex := accountIndexes[len(accountIndexes)-1]

	currentMarket, ok := r.store.markets[latestAccountIndex]
	if !ok {
		return nil, -1, ErrMarketNotExist
	}

	return &currentMarket, latestAccountIndex, nil
}

func keyFromAssets(baseAsset, quoteAsset string) string {
	key := baseAsset + quoteAsset
	keyBytes := sha256.Sum256([]byte(key))
	return hex.EncodeToString(keyBytes[:])
}
