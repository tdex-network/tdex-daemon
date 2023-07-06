package mapper

import (
	v0webhook "github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1/v0-webhook"
	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1/v1-domain"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	v0domain "github.com/tdex-network/tdex-daemon/old-v0"
)

type Service interface {
	FromV091VaultToV1Wallet(
		vault v0domain.Vault, walletPass string,
	) (*v1domain.Wallet, error)
	FromV091TradesToV1Trades(
		trades []*v0domain.Trade,
	) ([]*domain.Trade, error)
	FromV091DepositsToV1Deposits(
		deposits []*v0domain.Deposit,
	) ([]domain.Deposit, error)
	FromV091WithdrawalsToV1Withdrawals(
		withdrawals []*v0domain.Withdrawal,
	) ([]domain.Withdrawal, error)
	FromV091WebhooksToV1Subscriptions(
		webhooks []*v0webhook.Webhook,
	) ([]ports.Webhook, error)
	FromV091MarketsToV1Markets(
		markets []*v0domain.Market,
	) ([]*domain.Market, error)
	FromV091UnspentsToV1Utxos(
		unspents []*v0domain.Unspent,
	) ([]*v1domain.Utxo, error)
	FromV091TransactionsToV1Transactions(
		trades []*domain.Trade, deposits []*domain.Deposit,
		withdrawals []*domain.Withdrawal, accountsByLabel map[string]string,
	) map[string]*v1domain.Transaction
}

type mapperService struct {
	v0RepoManager v0domain.TdexRepoManager
}

func NewService(
	v0RepoManager v0domain.TdexRepoManager,
) Service {
	return &mapperService{
		v0RepoManager: v0RepoManager,
	}
}
