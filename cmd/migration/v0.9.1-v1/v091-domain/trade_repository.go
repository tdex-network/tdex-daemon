package v091domain

import (
	"github.com/sekulicd/badgerhold/v2"
)

type TradeRepository interface {
	GetAllTrades() ([]*Trade, error)
}

type tradeRepositoryImpl struct {
	store *badgerhold.Store
}

func NewTradeRepositoryImpl(store *badgerhold.Store) TradeRepository {
	return &tradeRepositoryImpl{store}
}

func (t tradeRepositoryImpl) GetAllTrades() ([]*Trade, error) {
	var tr []Trade
	if err := t.store.Find(&tr, nil); err != nil {
		return nil, err
	}

	trades := make([]*Trade, 0, len(tr))
	for i := range tr {
		trade := tr[i]
		trades = append(trades, &trade)
	}

	return trades, nil
}
