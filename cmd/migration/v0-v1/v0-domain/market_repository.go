package v0domain

import "github.com/timshannon/badgerhold/v2"

type MarketRepository interface {
	GetMarketByAccount(accountIndex int) (*Market, error)
	GetMarketByAssets(baseAsset, quoteAsset string) (*Market, error)
	GetAllMarkets() ([]*Market, error)
}

type marketRepositoryImpl struct {
	mainDb  *badgerhold.Store
	priceDb *badgerhold.Store
}

func NewMarketRepositoryImpl(mainDb, priceDb *badgerhold.Store) MarketRepository {
	return &marketRepositoryImpl{
		mainDb:  mainDb,
		priceDb: priceDb,
	}
}

func (m *marketRepositoryImpl) GetMarketByAccount(
	accountIndex int,
) (*Market, error) {
	var market Market

	if err := m.mainDb.Get(accountIndex, &market); err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	price, err := m.getPriceByAccountIndex(accountIndex)
	if err != nil {
		return nil, err
	}
	market.Price = *price

	return &market, nil
}

func (m *marketRepositoryImpl) GetMarketByAssets(
	baseAsset, quoteAsset string,
) (*Market, error) {
	query := badgerhold.
		Where("BaseAsset").Eq(baseAsset).And("QuoteAsset").Eq(quoteAsset)
	markets, err := m.findMarkets(query)
	if err != nil {
		return nil, err
	}

	var market *Market
	if len(markets) > 0 {
		market = markets[0]
	}

	return market, nil
}

func (m *marketRepositoryImpl) GetAllMarkets() ([]*Market, error) {
	query := badgerhold.Where("AccountIndex").Ge(FeeAccount)

	return m.findMarkets(query)
}

func (m *marketRepositoryImpl) getPriceByAccountIndex(
	accountIndex int,
) (*Prices, error) {
	var prices Prices

	if err := m.priceDb.Get(accountIndex, &prices); err != nil {
		return nil, err
	}

	return &prices, nil
}

func (m *marketRepositoryImpl) findMarkets(
	query *badgerhold.Query,
) ([]*Market, error) {
	var markets []Market
	var err error

	if err = m.mainDb.Find(&markets, query); err != nil {
		return nil, err
	}

	for i, mkt := range markets {
		price, err := m.getPriceByAccountIndex(mkt.AccountIndex)
		if err != nil {
			return nil, err
		}
		mkt.Price = *price
		markets[i] = mkt
	}

	resp := make([]*Market, 0, len(markets))
	for i := range markets {
		mkt := markets[i]
		resp = append(resp, &mkt)
	}

	return resp, err
}
