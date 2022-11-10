package inmemory

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type RepoManager struct {
	marketRepository      domain.MarketRepository
	tradeRepository       domain.TradeRepository
	depositRepository     domain.DepositRepository
	withdrawalsRepository domain.WithdrawalRepository
}

func NewRepoManager() ports.RepoManager {
	marketRepo := NewMarketRepositoryImpl()
	tradeRepo := NewTradeRepositoryImpl()
	depositRepo := NewDepositRepositoryImpl()
	withdrawalRepo := NewWithdrawalRepositoryImpl()

	return &RepoManager{
		marketRepository:      marketRepo,
		tradeRepository:       tradeRepo,
		depositRepository:     depositRepo,
		withdrawalsRepository: withdrawalRepo,
	}
}

func (d *RepoManager) MarketRepository() domain.MarketRepository {
	return d.marketRepository
}

func (d *RepoManager) TradeRepository() domain.TradeRepository {
	return d.tradeRepository
}

func (d *RepoManager) DepositRepository() domain.DepositRepository {
	return d.depositRepository
}

func (d *RepoManager) WithdrawalRepository() domain.WithdrawalRepository {
	return d.withdrawalsRepository
}

func (d *RepoManager) Close() {}
