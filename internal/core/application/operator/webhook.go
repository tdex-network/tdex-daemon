package operator

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

func (s *service) AddWebhook(
	ctx context.Context, webhook ports.Webhook,
) (string, error) {
	return s.pubsub.AddWebhook(ctx, webhook)
}

func (s *service) RemoveWebhook(ctx context.Context, id string) error {
	return s.pubsub.RemoveWebhook(ctx, id)
}

func (s *service) ListWebhooks(
	ctx context.Context, event ports.WebhookEvent,
) ([]ports.WebhookInfo, error) {
	return s.pubsub.ListWebhooks(ctx, event)
}
