package ports

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

// DbManager interface defines the methods for swap, price and unspent.
type DbManager interface {
	VaultRepository() domain.VaultRepository
	MarketRepository() domain.MarketRepository
	UnspentRepository() domain.UnspentRepository
	TradeRepository() domain.TradeRepository

	Close()

	NewTransaction() Transaction
	NewPricesTransaction() Transaction
	NewUnspentsTransaction() Transaction
	RunTransaction(
		ctx context.Context,
		readOnly bool,
		handler func(ctx context.Context) (interface{}, error),
	) (interface{}, error)
	RunUnspentsTransaction(
		ctx context.Context,
		readOnly bool,
		handler func(ctx context.Context) (interface{}, error),
	) (interface{}, error)
	RunPricesTransaction(
		ctx context.Context,
		readOnly bool,
		handler func(ctx context.Context) (interface{}, error),
	) (interface{}, error)
}

// Transaction interface defines the method to commit or discard a database transaction.
type Transaction interface {
	Commit() error
	Discard()
}
