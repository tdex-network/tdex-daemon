package application

import (
	"github.com/tdex-network/tdex-daemon/internal/core/application/pubsub"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type PubSubService interface {
	SecurePubSub() ports.SecurePubSub
	PublisAccountLowBalanceTopic(
		accountName string, accountBalance map[string]ports.Balance,
		market ports.Market,
	) error
	PublisAccountWithdrawTopic(
		accountName string, accountBalance map[string]ports.Balance,
		withdrawal domain.Withdrawal, market ports.Market,
	) error
	PublishTradeSettledTopic(
		accountName string, accountBalance map[string]ports.Balance,
		trade domain.Trade,
	) error
	Close()
}

func NewPubSubService(pubsubSvc ports.SecurePubSub) PubSubService {
	return pubsub.NewService(pubsubSvc)
}
