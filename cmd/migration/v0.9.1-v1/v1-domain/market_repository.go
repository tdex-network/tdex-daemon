package v1domain

import (
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v4"
)

type MarketRepository interface {
	InsertMarkets(markets []domain.Market) error
}

type marketRepositoryImpl struct {
	marketDb *badgerhold.Store
	priceDb  *badgerhold.Store
}

func NewMarketRepositoryImpl(
	marketDb, priceDb *badgerhold.Store,
) MarketRepository {
	return &marketRepositoryImpl{
		marketDb: marketDb,
		priceDb:  priceDb,
	}
}

func (m *marketRepositoryImpl) InsertMarkets(markets []domain.Market) error {
	for _, v := range markets {
		if err := m.insertMarket(v); err != nil {
			return err
		}
	}

	return nil
}

func (m *marketRepositoryImpl) insertMarket(market domain.Market) error {
	if err := m.marketDb.Insert(market.Name, market); err != nil {
		if err == badgerhold.ErrKeyExists {
			return fmt.Errorf(
				"market with assets %s %s already exists",
				market.BaseAsset, market.QuoteAsset,
			)
		}
	}

	return m.updateMarketPrice(market.Name, market.Price)
}

func (m *marketRepositoryImpl) updateMarketPrice(
	marketName string, price domain.MarketPrice,
) (err error) {
	if err := m.priceDb.Upsert(marketName, price); err != nil {
		return fmt.Errorf(
			"failed to update price for market %s: %s", marketName, err,
		)
	}

	return nil
}
