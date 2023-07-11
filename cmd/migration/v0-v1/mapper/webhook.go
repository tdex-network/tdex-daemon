package mapper

import (
	v0webhook "github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1/v0-webhook"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

func (m *mapperService) FromV0WebhooksToV1Subscriptions(
	webhooks []*v0webhook.Webhook,
) ([]ports.Webhook, error) {
	res := make([]ports.Webhook, 0, len(webhooks))
	for _, v := range webhooks {
		subscription, err := m.fromV0WebhookToV1Subscription(*v)
		if err != nil {
			return nil, err
		}
		res = append(res, subscription)
	}

	return res, nil
}

func (m *mapperService) fromV0WebhookToV1Subscription(
	webhook v0webhook.Webhook,
) (ports.Webhook, error) {
	return webhookV1(webhook), nil
}

type webhookV1 v0webhook.Webhook

func (w webhookV1) GetId() string {
	return w.ID
}
func (w webhookV1) GetEndpoint() string {
	return w.Endpoint
}
func (w webhookV1) GetSecret() string {
	return w.Secret
}
func (w webhookV1) GetEvent() ports.WebhookEvent {
	return webhookEvent(w.ActionType)
}

type webhookEvent int

func (e webhookEvent) IsUnspecified() bool {
	return false
}
func (e webhookEvent) IsTradeSettled() bool {
	return int(e) == int(v0webhook.TradeSettled)
}
func (e webhookEvent) IsAccountLowBalance() bool {
	return int(e) == int(v0webhook.AccountLowBalance)
}
func (e webhookEvent) IsAccountWithdraw() bool {
	return int(e) == int(v0webhook.AccountWithdraw)
}
func (e webhookEvent) IsAccountDeposit() bool {
	return false
}
func (e webhookEvent) IsAny() bool {
	return int(e) == int(v0webhook.AllActions)
}
