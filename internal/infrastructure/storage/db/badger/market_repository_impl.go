package dbbadger

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
)

type marketRepositoryImpl struct {
	store      *badgerhold.Store
	priceStore *badgerhold.Store
}

// NewMarketRepositoryImpl initialize a badger implementation of the domain.MarketRepository
func NewMarketRepositoryImpl(
	store, priceStore *badgerhold.Store,
) domain.MarketRepository {
	return marketRepositoryImpl{store, priceStore}
}

func (m marketRepositoryImpl) AddMarket(
	ctx context.Context, market *domain.Market,
) error {
	return m.insertMarket(ctx, market)
}

func (m marketRepositoryImpl) GetMarketByName(
	ctx context.Context, marketName string,
) (market *domain.Market, err error) {
	return m.getMarket(ctx, marketName)
}

func (m marketRepositoryImpl) GetMarketByAssets(
	ctx context.Context, baseAsset, quoteAsset string,
) (market *domain.Market, err error) {
	query := badgerhold.Where("BaseAsset").Eq(baseAsset).
		And("QuoteAsset").Eq(quoteAsset)
	markets, err := m.findMarkets(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(markets) == 0 {
		return nil, fmt.Errorf(
			"market with assets %s %s not found", baseAsset, quoteAsset,
		)
	}

	market = &markets[0]
	return
}

func (m marketRepositoryImpl) GetTradableMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	query := badgerhold.Where("Tradable").Eq(true)
	return m.findMarkets(ctx, query)
}

func (m marketRepositoryImpl) GetAllMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	query := &badgerhold.Query{}
	return m.findMarkets(ctx, query)
}

func (m marketRepositoryImpl) UpdateMarket(
	ctx context.Context,
	marketName string, updateFn func(m *domain.Market) (*domain.Market, error),
) error {
	currentMarket, err := m.getMarket(ctx, marketName)
	if err != nil {
		return err
	}
	if currentMarket == nil {
		return fmt.Errorf("market with name %s not found", marketName)
	}

	updatedMarket, err := updateFn(currentMarket)
	if err != nil {
		return err
	}

	return m.updateMarket(ctx, *updatedMarket)
}

func (m marketRepositoryImpl) OpenMarket(
	ctx context.Context, marketName string,
) error {
	market, err := m.getMarket(ctx, marketName)
	if err != nil {
		return err
	}

	if market.IsTradable() {
		return nil
	}

	err = market.MakeTradable()
	if err != nil {
		return err
	}

	return m.updateMarket(ctx, *market)
}

func (m marketRepositoryImpl) CloseMarket(
	ctx context.Context, marketName string,
) error {
	market, err := m.getMarket(ctx, marketName)
	if err != nil {
		return err
	}

	if !market.IsTradable() {
		return nil
	}

	market.MakeNotTradable()

	return m.updateMarket(ctx, *market)
}

func (m marketRepositoryImpl) DeleteMarket(
	ctx context.Context, marketName string,
) error {
	var err error

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = m.store.TxDelete(tx, marketName, domain.Market{})
	} else {
		err = m.store.Delete(marketName, domain.Market{})
	}
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return fmt.Errorf("market with name %s not found", marketName)
		}
		return err
	}
	return nil
}

func (m marketRepositoryImpl) UpdateMarketPrice(
	ctx context.Context, marketName string, price domain.MarketPrice,
) error {
	return m.updateMarketPrice(ctx, marketName, price)
}

func (m marketRepositoryImpl) insertMarket(
	ctx context.Context, market *domain.Market,
) error {
	var err error

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = m.store.TxInsert(tx, market.Name, market)
	} else {
		err = m.store.Insert(market.Name, market)
	}
	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return fmt.Errorf(
				"market with assets %s %s already exists",
				market.BaseAsset, market.QuoteAsset,
			)
		}
	}

	return m.updateMarketPrice(ctx, market.Name, market.Price)
}

func (m marketRepositoryImpl) getMarket(
	ctx context.Context, marketName string,
) (*domain.Market, error) {
	var err error
	var market domain.Market

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = m.store.TxGet(tx, marketName, &market)
	} else {
		err = m.store.Get(marketName, &market)
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, fmt.Errorf("market with name %s not found", marketName)
		}
		return nil, err
	}

	// Retrieve MarketPrice from dedicated storage.
	price, err := m.getMarketPrice(ctx, market.Name)
	if err != nil {
		return nil, err
	}
	market.Price = *price

	return &market, nil
}

func (m marketRepositoryImpl) updateMarket(
	ctx context.Context, market domain.Market,
) error {
	var err error
	// MarketPrice is stored in a dedicated storage, therefore it's always
	// zero-ed when a market is updated to make sure its value never changes in
	// the market storage.
	market.Price = domain.MarketPrice{}

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = m.store.TxUpdate(tx, market.Name, market)
	} else {
		err = m.store.Update(market.Name, market)
	}
	if err != nil {
		return fmt.Errorf("failed to update market %s: %s", market.Name, err)
	}

	return nil
}

func (m marketRepositoryImpl) findMarkets(
	ctx context.Context, query *badgerhold.Query,
) ([]domain.Market, error) {
	var markets []domain.Market
	var err error

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = m.store.TxFind(tx, &markets, query)
	} else {
		err = m.store.Find(&markets, query)
	}
	for i := range markets {
		// Retrieve price from Price storage.
		price, err := m.getMarketPrice(ctx, markets[i].Name)
		if err != nil {
			return nil, err
		}
		markets[i].Price = *price
	}

	return markets, err
}

func (m marketRepositoryImpl) getMarketPrice(
	ctx context.Context, marketName string,
) (price *domain.MarketPrice, err error) {
	if ctx.Value("ptx") != nil {
		tx := ctx.Value("ptx").(*badger.Txn)
		err = m.store.TxGet(tx, marketName, &price)
	} else {
		err = m.priceStore.Get(marketName, &price)
	}
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get price for market %s: %s", marketName, err,
		)
	}

	return price, nil
}

func (m marketRepositoryImpl) updateMarketPrice(
	ctx context.Context, marketName string, price domain.MarketPrice,
) (err error) {
	if ctx.Value("ptx") != nil {
		tx := ctx.Value("ptx").(*badger.Txn)
		err = m.store.TxUpsert(tx, marketName, price)
	} else {
		err = m.priceStore.Upsert(marketName, price)
	}
	if err != nil {
		return fmt.Errorf(
			"failed to update price for market %s: %s", marketName, err,
		)
	}

	return nil
}
