package operator

import (
	"context"
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

func (s *service) AddWebhook(
	ctx context.Context, hook ports.Webhook,
) (string, error) {
	topics := s.pubsub.SecurePubSub().TopicsByCode()
	topic, ok := topics[hook.GetActionType()]
	if !ok {
		return "", fmt.Errorf("unknown action type")
	}

	return s.pubsub.SecurePubSub().Subscribe(
		topic.Label(), hook.GetEndpoint(), hook.GetSecret(),
	)
}

func (s *service) RemoveWebhook(ctx context.Context, id string) error {
	return s.pubsub.SecurePubSub().Unsubscribe("", id)
}

func (s *service) ListWebhooks(
	ctx context.Context, actionType int,
) ([]ports.WebhookInfo, error) {
	topics := s.pubsub.SecurePubSub().TopicsByCode()
	topic, ok := topics[actionType]
	if !ok {
		return nil, fmt.Errorf("unknown action type")
	}

	subs := s.pubsub.SecurePubSub().ListSubscriptionsForTopic(topic.Label())
	return webhookList(subs).toPortableList(), nil
}

type webhookInfo struct {
	ports.Subscription
}

func (i webhookInfo) GetId() string {
	return i.Subscription.Id()
}
func (i webhookInfo) GetActionType() int {
	return i.Subscription.Topic().Code()
}
func (i webhookInfo) GetEndpoint() string {
	return i.Subscription.NotifyAt()
}

func (i webhookInfo) IsSecured() bool {
	return i.Subscription.IsSecured()
}

type webhookList []ports.Subscription

func (l webhookList) toPortableList() []ports.WebhookInfo {
	list := make([]ports.WebhookInfo, 0, len(l))
	for _, w := range l {
		list = append(list, webhookInfo{w})
	}
	return list
}
