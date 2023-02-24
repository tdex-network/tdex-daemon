package domain

import (
	"context"
)

// TradeRepository is the abstraction for any kind of database intended to
// persist Trades.
type TradeRepository interface {
	// AddTrade adds a new trade to the repository.
	AddTrade(ctx context.Context, trade *Trade) error
	// GetTradeById returns the trade with the given id if existing.
	GetTradeById(ctx context.Context, id string) (*Trade, error)
	// GetAllTrades returns all the trades stored in the repository.
	GetAllTrades(ctx context.Context, page Page) ([]Trade, error)
	// GetAllTradesByMarket returns all the trades filtered by a market
	// identified by its name.
	GetAllTradesByMarket(
		ctx context.Context, marketName string, page Page,
	) ([]Trade, error)
	// GetCompletedTradesByMarket returns all the Completed or Settled trades
	// for the provided market identified by its name.
	GetCompletedTradesByMarket(
		ctx context.Context, marketName string, page Page,
	) ([]Trade, error)
	// GetTradeBySwapAcceptId returns the trade that contains the SwapAccept
	// message matching the given id.
	GetTradeBySwapAcceptId(
		ctx context.Context, swapAcceptId string,
	) (*Trade, error)
	// GetTradeByTxid returns the trade which transaction matches the given
	// transaction id.
	GetTradeByTxId(ctx context.Context, txid string) (*Trade, error)
	// UpdateTrade allowa to commit multiple changes to the same trade in a
	// transactional way.
	UpdateTrade(
		ctx context.Context,
		tradeId string, updateFn func(t *Trade) (*Trade, error),
	) error
}
