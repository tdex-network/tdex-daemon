package inmemory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

type marketInmemoryStore struct {
	markets         map[string]domain.Market
	nameByAssetsKey map[string]string
	locker          *sync.RWMutex
}

type marketRepositoryImpl struct {
	store *marketInmemoryStore
}

// NewMarketRepositoryImpl returns a new inmemory MarketRepository implementation.
func NewMarketRepositoryImpl() domain.MarketRepository {
	return &marketRepositoryImpl{&marketInmemoryStore{
		markets:         map[string]domain.Market{},
		nameByAssetsKey: map[string]string{},
		locker:          &sync.RWMutex{},
	}}
}

func (r *marketRepositoryImpl) AddMarket(
	_ context.Context, market *domain.Market,
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.createMarket(market)
}

func (r *marketRepositoryImpl) GetMarketByName(
	_ context.Context, marketName string,
) (*domain.Market, error) {
	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

	return r.getMarketByName(marketName)
}

func (r *marketRepositoryImpl) GetMarketByAssets(
	_ context.Context, baseAsset, quoteAsset string,
) (*domain.Market, error) {
	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

	return r.getMarketByAssets(baseAsset, quoteAsset)
}

func (r *marketRepositoryImpl) GetTradableMarkets(
	_ context.Context,
) (tradableMarkets []domain.Market, err error) {
	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

	for _, mkt := range r.store.markets {
		if mkt.IsTradable() {
			tradableMarkets = append(tradableMarkets, mkt)
		}
	}

	return tradableMarkets, nil
}

func (r *marketRepositoryImpl) GetAllMarkets(
	_ context.Context,
) ([]domain.Market, error) {
	r.store.locker.RLock()
	defer r.store.locker.RUnlock()

	markets := make([]domain.Market, 0)

	for _, mkt := range r.store.markets {
		markets = append(markets, mkt)
	}

	return markets, nil
}

func (r *marketRepositoryImpl) UpdateMarket(
	_ context.Context,
	marketName string,
	updateFn func(m *domain.Market) (*domain.Market, error),
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	currentMarket, err := r.getMarketByName(marketName)
	if err != nil {
		return err
	}

	updatedMarket, err := updateFn(currentMarket)
	if err != nil {
		return err
	}

	key := keyFromAssets(updatedMarket.BaseAsset, updatedMarket.QuoteAsset)
	r.store.markets[marketName] = *updatedMarket
	r.store.nameByAssetsKey[key] = marketName

	return nil
}

func (r *marketRepositoryImpl) OpenMarket(
	_ context.Context, marketName string,
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	currentMarket, err := r.getMarketByName(marketName)
	if err != nil {
		return err
	}

	if currentMarket.IsTradable() {
		return nil
	}

	err = currentMarket.MakeTradable()
	if err != nil {
		return err
	}

	r.store.markets[marketName] = *currentMarket

	return nil
}

func (r *marketRepositoryImpl) CloseMarket(
	_ context.Context, marketName string,
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	currentMarket, err := r.getMarketByName(marketName)
	if err != nil {
		return err
	}

	if !currentMarket.IsTradable() {
		return nil
	}

	currentMarket.MakeNotTradable()

	r.store.markets[marketName] = *currentMarket

	return nil
}

func (r *marketRepositoryImpl) UpdateMarketPrice(
	_ context.Context, marketName string, price domain.MarketPrice,
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	market, err := r.getMarketByName(marketName)
	if err != nil {
		return err
	}

	if err := market.ChangePrice(
		price.GetBasePrice(), price.GetQuotePrice(),
	); err != nil {
		return err
	}

	r.store.markets[marketName] = *market
	return nil
}

func (r *marketRepositoryImpl) DeleteMarket(
	_ context.Context, marketName string,
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	if _, ok := r.store.markets[marketName]; !ok {
		return fmt.Errorf("market with name %s not found", marketName)
	}
	delete(r.store.markets, marketName)

	return nil
}

func (r *marketRepositoryImpl) createMarket(market *domain.Market) error {
	if mkt, _ := r.getMarketByName(market.Name); mkt != nil {
		return fmt.Errorf(
			"market with assets %s %s already exists",
			market.BaseAsset, market.QuoteAsset,
		)
	}
	key := keyFromAssets(market.BaseAsset, market.QuoteAsset)
	r.store.markets[market.Name] = *market
	r.store.nameByAssetsKey[key] = market.Name
	return nil
}

func (r *marketRepositoryImpl) getMarketByName(
	name string,
) (*domain.Market, error) {
	market, ok := r.store.markets[name]
	if !ok {
		return nil, fmt.Errorf("market with name %s not found", name)
	}

	return &market, nil
}

func (r *marketRepositoryImpl) getMarketByAssets(
	baseAsset, quoteAsset string,
) (*domain.Market, error) {
	key := keyFromAssets(baseAsset, quoteAsset)
	marketName, ok := r.store.nameByAssetsKey[key]
	if !ok {
		return nil, fmt.Errorf(
			"market with assets %s %s not found", baseAsset, quoteAsset,
		)
	}
	market := r.store.markets[marketName]
	return &market, nil
}

func keyFromAssets(baseAsset, quoteAsset string) string {
	key := baseAsset + quoteAsset
	keyBytes := sha256.Sum256([]byte(key))
	return hex.EncodeToString(keyBytes[:])
}
