package v0domain

import "github.com/timshannon/badgerhold/v2"

type TradeRepository interface {
	GetAllTrades() ([]*Trade, error)
}

type tradeRepositoryImpl struct {
	tradeDb *badgerhold.Store
}

func NewTradeRepositoryImpl(tradeDb *badgerhold.Store) TradeRepository {
	return &tradeRepositoryImpl{tradeDb}
}

func (t *tradeRepositoryImpl) GetAllTrades() ([]*Trade, error) {
	var trades []*Trade
	if err := t.tradeDb.Find(&trades, &badgerhold.Query{}); err != nil {
		return nil, err
	}

	return trades, nil
}
