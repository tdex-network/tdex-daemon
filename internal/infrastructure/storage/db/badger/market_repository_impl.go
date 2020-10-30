package dbbadger

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
	"github.com/timshannon/badgerhold/v2"
)

type marketRepositoryImpl struct {
	db *DbManager
}

//NewMarketRepositoryImpl initialize a badger implementation of the domain.MarketRepository
func NewMarketRepositoryImpl(db *DbManager) domain.MarketRepository {
	return marketRepositoryImpl{
		db: db,
	}
}

func (m marketRepositoryImpl) GetMarketByAsset(
	ctx context.Context,
	quoteAsset string,
) (market *domain.Market, accountIndex int, err error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("QuoteAsset").Eq(quoteAsset)
	markets, err := m.findMarkets(tx, query)
	if err != nil {
		return
	}

	if len(markets) > 0 {
		market = &markets[0]
		accountIndex = market.AccountIndex
	}

	return
}

func (m marketRepositoryImpl) GetLatestMarket(
	ctx context.Context,
) (market *domain.Market, accountIndex int, err error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("AccountIndex").
		Ge(domain.MarketAccountStart).
		SortBy("AccountIndex").
		Reverse()
	markets, err := m.findMarkets(tx, query)
	if err != nil {
		return
	}

	accountIndex = domain.MarketAccountStart - 1
	if len(markets) > 0 {
		market = &markets[0]
		accountIndex = market.AccountIndex
	}

	return
}

func (m marketRepositoryImpl) GetOrCreateMarket(
	ctx context.Context,
	accountIndex int,
) (*domain.Market, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	return m.getOrCreateMarket(tx, accountIndex)
}

func (m marketRepositoryImpl) GetTradableMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("AccountIndex").
		Ge(domain.MarketAccountStart).
		And("Tradable").
		Eq(true)

	return m.findMarkets(tx, query)
}

func (m marketRepositoryImpl) GetAllMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("AccountIndex").Ge(domain.MarketAccountStart)

	return m.findMarkets(tx, query)
}

func (m marketRepositoryImpl) UpdateMarket(
	ctx context.Context,
	accountIndex int,
	updateFn func(m *domain.Market) (*domain.Market, error),
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	currentMarket, err := m.getOrCreateMarket(tx, accountIndex)
	if err != nil {
		return err
	}

	updatedMarket, err := updateFn(currentMarket)
	if err != nil {
		return err
	}

	return m.updateMarket(tx, accountIndex, *updatedMarket)
}

func (m marketRepositoryImpl) OpenMarket(
	ctx context.Context,
	quoteAsset string,
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("QuoteAsset").Eq(quoteAsset)
	markets, err := m.findMarkets(tx, query)
	if err != nil {
		return err
	}

	if len(markets) > 0 {
		market := markets[0]
		if market.IsTradable() {
			return nil
		}

		err = market.MakeTradable()
		if err != nil {
			return err
		}

		err := m.updateMarket(tx, market.AccountIndex, market)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m marketRepositoryImpl) CloseMarket(
	ctx context.Context,
	quoteAsset string,
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("QuoteAsset").Eq(quoteAsset)
	markets, err := m.findMarkets(tx, query)
	if err != nil {
		return err
	}

	if len(markets) > 0 {
		market := markets[0]
		if !market.IsTradable() {
			return nil
		}

		err = market.MakeNotTradable()
		if err != nil {
			return err
		}

		err := m.updateMarket(tx, market.AccountIndex, market)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m marketRepositoryImpl) getOrCreateMarket(
	tx *badger.Txn,
	accountIndex int,
) (*domain.Market, error) {
	market, err := m.getMarket(tx, accountIndex)
	if err != nil {
		return nil, err
	}

	if market == nil {
		market, err = domain.NewMarket(accountIndex)
		if err != nil {
			return nil, err
		}

		err = m.insertMarket(tx, accountIndex, *market)
		if err != nil {
			return nil, err
		}
	}

	return market, nil
}

func (m marketRepositoryImpl) insertMarket(
	tx *badger.Txn,
	accountIndex int,
	market domain.Market,
) error {
	var err error
	if tx != nil {
		err = m.db.Store.TxInsert(
			tx,
			accountIndex,
			&market,
		)
	} else {
		err = m.db.Store.Insert(
			vaultKey,
			&market,
		)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}

	//Now we update the price store as well only if market insertion went ok
	err = m.updatePriceByAccountIndex(accountIndex, market.Price)

	return nil
}

func (m marketRepositoryImpl) getMarket(
	tx *badger.Txn,
	accountIndex int,
) (*domain.Market, error) {
	var err error
	var market domain.Market
	if tx != nil {
		err = m.db.Store.TxGet(
			tx,
			accountIndex,
			&market,
		)
	} else {
		err = m.db.Store.Get(
			accountIndex,
			&market,
		)
	}

	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	// Let's get the price from PriceStore
	price, err := m.getPriceByAccountIndex(accountIndex)
	if err != nil {
		return nil, err
	}
	market.Price = *price
	//Restore strategy
	restoreStrategy(&market)

	return &market, nil
}

func (m marketRepositoryImpl) updateMarket(
	tx *badger.Txn,
	accountIndex int,
	market domain.Market,
) error {
	var err error
	if tx != nil {
		err = m.db.Store.TxUpdate(
			tx,
			accountIndex,
			market,
		)
	} else {
		err = m.db.Store.Update(
			accountIndex,
			market,
		)
	}

	if err != nil {
		return fmt.Errorf("trying to update market with account index %v %w", accountIndex, err)
	}

	//Now we update the price store as well only if market went ok
	err = m.updatePriceByAccountIndex(accountIndex, market.Price)

	return err
}

func (m marketRepositoryImpl) findMarkets(
	tx *badger.Txn,
	query *badgerhold.Query,
) ([]domain.Market, error) {
	var markets []domain.Market
	var err error
	if tx != nil {
		err = m.db.Store.TxFind(
			tx,
			&markets,
			query,
		)
	} else {
		err = m.db.Store.Find(
			&markets,
			query,
		)
	}
	for i, mkt := range markets {
		// Let's get the price from PriceStore
		price, err := m.getPriceByAccountIndex(mkt.AccountIndex)
		if err != nil {
			return nil, err
		}
		mkt.Price = *price

		restoreStrategy(&mkt)

		// Assign the restore market with price and startegy
		markets[i] = mkt
	}

	return markets, err
}

func (m marketRepositoryImpl) getPriceByAccountIndex(accountIndex int) (*domain.Prices, error) {
	var prices *domain.Prices

	err := m.db.PriceStore.Get(accountIndex, &prices)
	if err != nil {
		return nil, fmt.Errorf("trying to get price with account index %v %w", accountIndex, err)
	}

	return prices, nil
}

func (m marketRepositoryImpl) updatePriceByAccountIndex(accountIndex int, prices domain.Prices) error {
	err := m.db.PriceStore.Upsert(
		accountIndex,
		prices,
	)
	if err != nil {
		return fmt.Errorf("trying to updating price with account index %v %w", accountIndex, err)
	}

	return nil
}

func restoreStrategy(market *domain.Market) {
	if !market.IsStrategyPluggable() {
		switch market.Strategy.Type {
		case formula.BalancedReservesType:
			market.Strategy = mm.NewStrategyFromFormula(formula.BalancedReserves{})
		}
	}
}
