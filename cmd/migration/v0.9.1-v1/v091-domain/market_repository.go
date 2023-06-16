package v091domain

import (
	"github.com/sekulicd/badgerhold/v2"
)

type MarketRepository interface {
	GetMarketByAccount(accountIndex int) (*Market, error)
	GetMarketByAssets(baseAsset, quoteAsset string) (*Market, error)
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
	restoreStrategy(&market)

	return &market, nil
}

func (m *marketRepositoryImpl) GetMarketByAssets(
	baseAsset, quoteAsset string,
) (*Market, error) {
	//TODO implement me
	panic("implement me")
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

func restoreStrategy(market *Market) {
	var strategy MakingStrategy
	switch market.Strategy.Type {
	case BalancedReservesType:
		strategy = NewStrategyFromFormula(BalancedReserves{})
	default:
		strategy = NewStrategyFromFormula(PluggableStrategy{})
	}
	market.Strategy = strategy
}
