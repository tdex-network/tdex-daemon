package v091domain

import (
	"github.com/sekulicd/badgerhold/v2"
)

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
	var tr []Trade
	if err := t.tradeDb.Find(&tr, nil); err != nil {
		return nil, err
	}

	trades := make([]*Trade, 0, len(tr))
	for i := range tr {
		trade := tr[i]
		trades = append(trades, &trade)
	}

	return trades, nil
}
