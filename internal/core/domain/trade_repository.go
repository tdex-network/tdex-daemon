package domain

import (
	"context"
	"github.com/google/uuid"
)

// TradeRepository defines the abstraction for Trade
type TradeRepository interface {
	GetOrCreateTrade(ctx context.Context, tradeID *uuid.UUID) (*Trade, error)
	GetAllTrades(ctx context.Context) ([]*Trade, error)
	GetAllTradesByMarket(ctx context.Context, marketQuoteAsset string) ([]*Trade, error)
	GetTradeBySwapAcceptID(ctx context.Context, swapAcceptID string) (*Trade, error)
	UpdateTrade(
		ctx context.Context,
		tradeID *uuid.UUID,
		updateFn func(t *Trade) (*Trade, error),
	) error
}
