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
	GetAllTradesByMarket(ctx context.Context, marketQuoteAsset string) ([]*Trade, error)
	// GetCompletedTradesByMarket returns all the Completed or Settled trades
	// for the provided market identified by its quote asset.
	GetCompletedTradesByMarket(ctx context.Context, marketQuoteAsset string) ([]*Trade, error)
	// GetTradeBySwapAcceptID returns the trade that contains the SwapAccept
	// message matching the given id.
	GetTradeBySwapAcceptID(ctx context.Context, swapAcceptID string) (*Trade, error)
	// GetTradeByTxID returns the trade which transaction matches the given
	// transaction id.
	GetTradeByTxID(ctx context.Context, txID string) (*Trade, error)
	// UpdateTrade allowa to commit multiple changes to the same trade in a
	// transactional way.
	UpdateTrade(
		ctx context.Context,
		tradeID *uuid.UUID,
		updateFn func(t *Trade) (*Trade, error),
	) error
}
