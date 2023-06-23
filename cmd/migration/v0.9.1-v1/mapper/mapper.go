package mapper

import (
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"
	v091webhook "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-webhook"
	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-domain"
	v1subscription "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-subscription"
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
	FromV091WebhooksToV1Subscriptions(
		webhooks []*v091webhook.Webhook,
	) ([]v1subscription.Subscription, error)
	FromV091MarketsToV1Markets(
		markets []*v091domain.Market,
	) ([]domain.Market, error)
	FromV091UnspentsToV1Utxos(
		unspents []*v091domain.Unspent,
	) ([]*v1domain.Utxo, error)
	GetUnspentStatus(txid string, index uint32) (*UtxoStatus, error)
}

type mapperService struct {
	v091RepoManager v091domain.Repository
	esploraUrl      string
}

func NewService(
	v091RepoManager v091domain.Repository, esploraUrl string,
) Service {
	return &mapperService{
		v091RepoManager: v091RepoManager,
		esploraUrl:      esploraUrl,
	}
}
