package trade

import (
	"context"
)

// Repository defines the abstraction for Trade
type Repository interface {
	GetOrCreateTrade(ctx context.Context, swapID string) (*Trade, error)
	GetAllTrades(ctx context.Context) ([]*Trade, error)
	GetAllTradesByMarket(ctx context.Context, marketIndex int) ([]*Trade, error)
	GetAllTradesByTrader(ctx context.Context, traderID string) ([]*Trade, error)
	UpdateTrade(
		ctx context.Context,
		tradeID string,
		updateFn func(t *Trade) (*Trade, error),
	) error
}
