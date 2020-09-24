package trade

import (
	"context"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/storageutil/uow"
)

// Repository defines the abstraction for Trade
type Repository interface {
	GetOrCreateTrade(ctx context.Context, tradeID *uuid.UUID) (*Trade, error)
	GetAllTrades(ctx context.Context) ([]*Trade, error)
	GetAllTradesByMarket(ctx context.Context, marketQuoteAsset string) ([]*Trade, error)
	GetAllTradesByTrader(ctx context.Context, traderID string) ([]*Trade, error)
	GetTradeBySwapAcceptID(ctx context.Context, swapAcceptID string) (*Trade, error)
	UpdateTrade(
		ctx context.Context,
		tradeID *uuid.UUID,
		updateFn func(t *Trade) (*Trade, error),
	) error
	Begin() (uow.Tx, error)
	ContextKey() interface{}
}
