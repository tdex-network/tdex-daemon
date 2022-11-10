package ports

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

// RepoManager interface defines the methods for swap, price and unspent.
type RepoManager interface {
	MarketRepository() domain.MarketRepository
	TradeRepository() domain.TradeRepository
	DepositRepository() domain.DepositRepository
	WithdrawalRepository() domain.WithdrawalRepository

	Close()
}
