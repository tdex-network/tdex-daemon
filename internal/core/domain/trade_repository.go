package domain

import (
	"context"

	"github.com/google/uuid"
)

// TradeRepository is the abstraction for any kind of database intended to
// persist Trades.
type TradeRepository interface {
	// // GetOrCreateVault returns the trade with the given tradeID, or create a
	// new empty one if not found.
	GetOrCreateTrade(ctx context.Context, tradeID *uuid.UUID) (*Trade, error)
	// GetAllTrades returns all the trades stored in the repository.
	GetAllTrades(ctx context.Context) ([]*Trade, error)
	// GetAllTradesByMarket returns all the trades filtered by a market
	// identified by its quote asset.
	GetAllTradesForMarket(ctx context.Context, marketQuoteAsset string) ([]*Trade, error)
	// GetCompletedTradesForMarket returns all the Completed or Settled trades
	// for the provided market identified by its quote asset.
	GetCompletedTradesForMarket(ctx context.Context, marketQuoteAsset string) ([]*Trade, error)
	// GetTradeWithSwapAcceptID returns the trade that contains the SwapAccept
	// message matching the given id.
	GetTradeWithSwapAcceptID(ctx context.Context, swapAcceptID string) (*Trade, error)
	// GetTradeWithTxID returns the trade which transaction matches the given
	// transaction id.
	GetTradeWithTxID(ctx context.Context, txID string) (*Trade, error)
	// UpdateTrade allowa to commit multiple changes to the same trade in a
	// transactional way.
	UpdateTrade(
		ctx context.Context,
		tradeID *uuid.UUID,
		updateFn func(t *Trade) (*Trade, error),
	) error
}
