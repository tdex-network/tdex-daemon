package ports

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

// RepoManager interface defines the methods for swap, price and unspent.
type RepoManager interface {
	VaultRepository() domain.VaultRepository
	MarketRepository() domain.MarketRepository
	UnspentRepository() domain.UnspentRepository
	TradeRepository() domain.TradeRepository
	DepositRepository() domain.DepositRepository
	WithdrawalRepository() domain.WithdrawalRepository

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
