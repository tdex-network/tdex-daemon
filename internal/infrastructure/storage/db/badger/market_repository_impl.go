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

func (m marketRepositoryImpl) GetMarketByAsset(
	ctx context.Context,
	quoteAsset string,
) (market *domain.Market, accountIndex int, err error) {
	mkt, err := m.getMarketByAsset(ctx, quoteAsset)
	if err != nil {
		return nil, -1, err
	}

	market = MapInfraMarketToDomainMarket(mkt)
	accountIndex = market.AccountIndex()
	return
}

func (m marketRepositoryImpl) GetLatestMarket(
	ctx context.Context,
) (market *domain.Market, accountIndex int, err error) {
	markets, err := m.getAllMarkets(ctx, true, false)
	if err != nil {
		return nil, -1, err
	}

	accountIndex = domain.MarketAccountStart - 1
	if len(markets) > 0 {
		market = MapInfraMarketToDomainMarket(markets[0])
		accountIndex = market.AccountIndex()
	}

	return
}

func (m marketRepositoryImpl) GetTradableMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	markets, err := m.getAllMarkets(ctx, false, true)
	if err != nil {
		return nil, err
	}

	domainMarkets := make([]domain.Market, 0, len(markets))
	for _, v := range markets {
		domainMarkets = append(domainMarkets, *MapInfraMarketToDomainMarket(v))
	}

	return domainMarkets, nil
}

func (m marketRepositoryImpl) GetAllMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	markets, err := m.getAllMarkets(ctx, false, false)
	if err != nil {
		return nil, err
	}

	domainMarkets := make([]domain.Market, 0, len(markets))
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
	currentMarket, err := m.getOrCreateMarket(ctx, accountIndex)
	if err != nil {
		return err
	}

	updatedMarket, err := updateFn(currentMarket)
	if err != nil {
		return err
	}

	return m.updateMarket(ctx, MapDomainMarketToInfraMarket(*updatedMarket))
}

func (m marketRepositoryImpl) OpenMarket(
	ctx context.Context,
	quoteAsset string,
) error {
	market, err := m.getMarketByAsset(ctx, quoteAsset)
	if err != nil {
		return err
	}

	mkt := MapInfraMarketToDomainMarket(market)
	if mkt.IsTradable() {
		return nil
	}

	if err := mkt.MakeTradable(); err != nil {
		return err
	}

	if err := m.updateMarket(ctx, MapDomainMarketToInfraMarket(*mkt)); err != nil {
		return err
	}

	return nil
}

func (m marketRepositoryImpl) CloseMarket(
	ctx context.Context,
	quoteAsset string,
) error {
	market, err := m.getMarketByAsset(ctx, quoteAsset)
	if err != nil {
		return err
	}

	mkt := MapInfraMarketToDomainMarket(market)
	if !mkt.IsTradable() {
		return nil
	}

	err = mkt.MakeNotTradable()
	if err != nil {
		return err
	}

	err = m.updateMarket(ctx, MapDomainMarketToInfraMarket(*mkt))
	if err != nil {
		return err
	}

	return nil
}

func (m marketRepositoryImpl) getOrCreateMarket(
	ctx context.Context,
	accountIndex int,
) (*domain.Market, error) {
	market, err := m.getMarketByIndex(ctx, accountIndex)
	if err != nil {
		return nil, err
	}
	if market != nil {
		return MapInfraMarketToDomainMarket(*market), nil
	}

	newMarket, err := domain.NewMarket(accountIndex)
	if err != nil {
		return nil, err
	}

	if err := m.insertMarket(
		ctx, accountIndex, MapDomainMarketToInfraMarket(*newMarket),
	); err != nil {
		return nil, err
	}

	return newMarket, nil
}

// getAllMarkets returns all the markets. If no market is present, the function
// does not raise an error but just return
func (m marketRepositoryImpl) getAllMarkets(ctx context.Context, sorted, filterTradable bool) ([]Market, error) {
	markets := make([]Market, 0)
	query := badgerhold.Where("AccountIndex").Ge(domain.MarketAccountStart)
	if sorted {
		query = query.SortBy("AccountIndex").Reverse()
	}
	if filterTradable {
		query = query.And("Tradable").Eq(true)
	}

	var err error
	if ctx.Value("tx") != nil {
		err = m.db.Store.TxFind(
			ctx.Value("tx").(*badger.Txn),
			&markets,
			query,
		)
	} else {
		err = m.db.Store.Find(
			&markets,
			query,
		)
	}
	if err != nil && err == badgerhold.ErrNotFound {
		err = nil
	}
	return markets, err
}

// getMarketByIndex returns the market identified by the given accountIndex.
// If not found, the function won't raise an error.
func (m marketRepositoryImpl) getMarketByIndex(ctx context.Context, accountIndex int) (market *Market, err error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)
	markets := make([]Market, 0)

	if ctx.Value("tx") != nil {
		err = m.db.Store.TxFind(
			ctx.Value("tx").(*badger.Txn),
			&markets,
			query,
		)
	} else {
		err = m.db.Store.Find(
			&markets,
			query,
		)
	}
	if err != nil && err == badgerhold.ErrNotFound {
		err = nil
	}
	if len(markets) > 0 {
		market = &markets[0]
	}

	return
}

// getMarketByAsset returns the market identified by the given quote asset.
// If not found, the the function will raise an error.
func (m marketRepositoryImpl) getMarketByAsset(ctx context.Context, quoteAsset string) (market Market, err error) {
	query := badgerhold.Where("QuoteAsset").Eq(quoteAsset)
	markets := make([]Market, 0)

	if ctx.Value("tx") != nil {
		err = m.db.Store.TxFind(
			ctx.Value("tx").(*badger.Txn),
			&markets,
			query,
		)
	} else {
		err = m.db.Store.Find(
			&markets,
			query,
		)
	}
	if len(markets) > 0 {
		market = markets[0]
	}

	return
}

func (m marketRepositoryImpl) insertMarket(ctx context.Context, accountIndex int, market *Market) (err error) {
	if ctx.Value("tx") != nil {
		err = m.db.Store.TxInsert(
			ctx.Value("tx").(*badger.Txn),
			accountIndex,
			market,
		)
	} else {
		err = m.db.Store.Insert(accountIndex, market)
	}
	return
}

func (m marketRepositoryImpl) updateMarket(ctx context.Context, market *Market) (err error) {
	if ctx.Value("tx") != nil {
		err = m.db.Store.TxUpdate(
			ctx.Value("tx").(*badger.Txn),
			market.AccountIndex,
			market,
		)
	} else {
		err = m.db.Store.Update(
			market.AccountIndex,
			market,
		)
	}
	return
}
