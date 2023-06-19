package mapper

import (
	v091webhook "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-webhook"
	v1subscription "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-subscription"
)

func (m *mapperService) FromV091WebhooksToV1Subscriptions(
	webhooks []*v091webhook.Webhook,
) ([]v1subscription.Subscription, error) {
	res := make([]v1subscription.Subscription, 0, len(webhooks))
	for _, v := range webhooks {
		subscription, err := m.fromV091WebhookToV1Subscription(v)
		if err != nil {
			return nil, err
		}
		res = append(res, subscription)
	}

	return res, nil
}

func (m *mapperService) fromV091WebhookToV1Subscription(
	webhook v091webhook.Webhook,
) (v1subscription.Subscription, error) {
	event := parseActionType(webhook.ActionType)
	return v1subscription.Subscription{
		ID:       webhook.ID,
		Event:    event,
		Endpoint: webhook.Endpoint,
		Secret:   webhook.Secret,
	}, nil
}

func parseActionType(actionType v091webhook.WebhookAction) string {
	switch actionType {
	case v091webhook.TradeSettled:
		return v1subscription.EventTradeSettled
	case v091webhook.AccountLowBalance:
		return v1subscription.EventAccountLowBalance
	case v091webhook.AccountWithdraw:
		return v1subscription.EventAccountWithdraw
	case v091webhook.AllActions:
		return v1subscription.AnyTopic
	}

	return ""
}
