package v1domain

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v4"
)

type TradeRepository interface {
	InsertTrades(trades []*domain.Trade) error
}

type tradeRepositoryImpl struct {
	store *badgerhold.Store
}

func NewTradeRepositoryImpl(store *badgerhold.Store) TradeRepository {
	return &tradeRepositoryImpl{store}
}

func (t *tradeRepositoryImpl) InsertTrades(trades []*domain.Trade) error {
	for _, trade := range trades {
		if err := t.store.Insert(trade.Id, *trade); err != nil {
			return err
		}
	}

	return nil
}
