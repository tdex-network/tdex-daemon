package mapper

import (
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"
	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-domain"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

type Service interface {
	FromV091VaultToV1Wallet(
		vault v091domain.Vault, walletPass string,
	) (*v1domain.Wallet, error)
	FromV091TradesToV1Trades(
		trades []*v091domain.Trade,
	) ([]*domain.Trade, error)
	FromV091DepositsToV1Deposits(
		deposits []*v091domain.Deposit,
	) ([]*domain.Deposit, error)
	FromV091WithdrawalsToV1Withdrawals(
		withdrawals []*v091domain.Withdrawal,
	) ([]*domain.Withdrawal, error)
}

type mapperService struct {
	v091RepoManager v091domain.Repository
}

func NewService(v091RepoManager v091domain.Repository) Service {
	return &mapperService{
		v091RepoManager: v091RepoManager,
	}
}
