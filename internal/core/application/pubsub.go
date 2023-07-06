package application

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application/pubsub"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type PubSubService interface {
	SecurePubSub() ports.SecurePubSub
	AddWebhook(ctx context.Context, hook ports.Webhook) (string, error)
	AddWebhookWithID(ctx context.Context, hook ports.Webhook) (string, error)
	RemoveWebhook(ctx context.Context, id string) error
	ListWebhooks(
		ctx context.Context, event ports.WebhookEvent,
	) ([]ports.WebhookInfo, error)
	PublisAccountLowBalanceEvent(
		accountName string, accountBalance map[string]ports.Balance,
		market ports.Market,
	) error
	PublisAccountWithdrawEvent(
		accountName string, accountBalance map[string]ports.Balance,
		withdrawal domain.Withdrawal, market ports.Market,
	) error
	PublishTradeSettledEvent(
		accountName string, accountBalance map[string]ports.Balance,
		trade domain.Trade,
	) error
	Close()
}

func NewPubSubService(pubsubSvc ports.SecurePubSub) PubSubService {
	return pubsub.NewService(pubsubSvc)
}
