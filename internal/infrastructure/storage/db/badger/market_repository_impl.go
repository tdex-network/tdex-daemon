package dbbadger

import (
	"context"
	"github.com/dgraph-io/badger"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold"
)

type marketRepositoryImpl struct {
	db *DbManager
}

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
	var markets []Market
	err = m.db.Store.TxFind(
		tx,
		&markets,
		badgerhold.Where("QuoteAsset").Eq(quoteAsset),
	)
	if err != nil {
		return
	}

	if len(markets) > 0 {
		market = MapInfraMarketToDomainMarket(markets[0])
		accountIndex = market.AccountIndex()
	}

	return
}

func (m marketRepositoryImpl) GetLatestMarket(
	ctx context.Context,
) (market *domain.Market, accountIndex int, err error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var markets []Market
	err = m.db.Store.TxFind(
		tx,
		&markets,
		badgerhold.Where("AccountIndex").Ge(domain.MarketAccountStart).SortBy("AccountIndex").Reverse(),
	)
	if err != nil {
		return
	}

	if len(markets) > 0 {
		market = MapInfraMarketToDomainMarket(markets[0])
		accountIndex = market.AccountIndex()
	}

	return
}

func (m marketRepositoryImpl) GetOrCreateMarket(
	ctx context.Context,
	accountIndex int,
) (*domain.Market, error) {
	market, err := m.getOrCreateMarket(ctx, accountIndex)
	if err != nil {
		return nil, err
	}
	return market, nil
}

func (m marketRepositoryImpl) getOrCreateMarket(
	ctx context.Context,
	accountIndex int,
) (*domain.Market, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var markets []Market
	err := m.db.Store.TxFind(
		tx,
		&markets,
		badgerhold.Where("AccountIndex").Eq(accountIndex),
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return nil, err
		}
	}

	var newMarket *domain.Market
	if len(markets) == 0 {
		newMarket, err = domain.NewMarket(accountIndex)
		if err != nil {
			return nil, err
		}
		err = m.db.Store.TxInsert(
			tx,
			accountIndex,
			MapDomainMarketToInfraMarket(*newMarket),
		)
		if err != nil {
			return nil, err
		}
	} else {
		newMarket = MapInfraMarketToDomainMarket(markets[0])
	}

	return newMarket, nil
}

func (m marketRepositoryImpl) GetTradableMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var markets []Market
	err := m.db.Store.TxFind(
		tx,
		&markets,
		badgerhold.Where("AccountIndex").Ge(domain.MarketAccountStart).And("Tradable").Eq(true),
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return nil, err
		}
	}

	domainMarkets := make([]domain.Market, 0)
	for _, v := range markets {
		domainMarkets = append(domainMarkets, *MapInfraMarketToDomainMarket(v))
	}

	return domainMarkets, nil
}

func (m marketRepositoryImpl) GetAllMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	var markets []Market
	err := m.db.Store.TxFind(
		tx,
		&markets,
		badgerhold.Where("AccountIndex").Ge(domain.MarketAccountStart),
	)
	if err != nil {
		if err != badgerhold.ErrNotFound {
			return nil, err
		}
	}

	domainMarkets := make([]domain.Market, 0)
	for _, v := range markets {
		domainMarkets = append(domainMarkets, *MapInfraMarketToDomainMarket(v))
	}

	return domainMarkets, nil
}

func (m marketRepositoryImpl) UpdateMarket(
	ctx context.Context,
	accountIndex int,
	updateFn func(m *domain.Market) (*domain.Market, error),
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	currentMarket, err := m.getOrCreateMarket(ctx, accountIndex)
	if err != nil {
		return err
	}

	updatedMarket, err := updateFn(currentMarket)
	if err != nil {
		return err
	}

	return m.db.Store.TxUpdate(
		tx,
		updatedMarket.AccountIndex(),
		MapDomainMarketToInfraMarket(*updatedMarket),
	)
}

func (m marketRepositoryImpl) OpenMarket(
	ctx context.Context,
	quoteAsset string,
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	var markets []Market
	err := m.db.Store.TxFind(
		tx,
		&markets,
		badgerhold.Where("QuoteAsset").Eq(quoteAsset),
	)
	if err != nil {
		return err
	}

	if len(markets) > 0 {
		market := MapInfraMarketToDomainMarket(markets[0])
		if market.IsTradable() {
			return nil
		}

		err = market.MakeTradable()
		if err != nil {
			return err
		}

		err := m.db.Store.TxUpdate(
			tx,
			markets[0].AccountIndex,
			MapDomainMarketToInfraMarket(*market),
		)
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

	var markets []Market
	err := m.db.Store.TxFind(
		tx,
		&markets,
		badgerhold.Where("QuoteAsset").Eq(quoteAsset),
	)
	if err != nil {
		return err
	}

	if len(markets) > 0 {
		market := MapInfraMarketToDomainMarket(markets[0])
		if !market.IsTradable() {
			return nil
		}

		err = market.MakeNotTradable()
		if err != nil {
			return err
		}

		err := m.db.Store.TxUpdate(
			tx,
			markets[0].AccountIndex,
			MapDomainMarketToInfraMarket(*market),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
