package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

const (
	eventTradeSettled      = "TRADE_SETTLED"
	eventAccountLowBalance = "ACCOUNT_LOW_BALANCE"
	eventAccountWithdraw   = "ACCOUNT_WITHDRAW"
	eventAccountDeposit    = "ACCOUNT_DEPOSIT"
)

type Service struct {
	pubsub ports.SecurePubSub
}

func NewService(pubsub ports.SecurePubSub) *Service {
	return &Service{pubsub}
}

func (s *Service) SecurePubSub() ports.SecurePubSub {
	return s.pubsub
}

func (s *Service) AddWebhook(
	_ context.Context, webhook ports.Webhook,
) (string, error) {
	if webhook.GetEvent().IsUnspecified() {
		return "", fmt.Errorf("invalid webhook event type")
	}
	topic := topicForEvent(webhook.GetEvent())
	return s.pubsub.Subscribe(topic, webhook.GetEndpoint(), webhook.GetSecret())
}

func (s *Service) RemoveWebhook(_ context.Context, id string) error {
	return s.pubsub.Unsubscribe(ports.UnspecifiedTopic, id)
}

func (s *Service) ListWebhooks(
	_ context.Context, event ports.WebhookEvent,
) ([]ports.WebhookInfo, error) {
	topic := topicForEvent(event)
	subs := s.pubsub.ListSubscriptionsForTopic(topic)
	webhooks := make([]ports.WebhookInfo, 0, len(subs))
	for _, s := range subs {
		webhooks = append(webhooks, webhookInfo{s})
	}
	return webhooks, nil
}

func (s *Service) PublisAccountLowBalanceEvent(
	accountName string, accountBalance map[string]ports.Balance,
	market ports.Market,
) error {
	event := eventAccountLowBalance
	account := getAccountPayload(accountName, market)
	balance := getBalancePayload(accountBalance)
	payload := map[string]interface{}{
		"event":   event,
		"account": account,
		"balance": balance,
	}
	message, _ := json.Marshal(payload)

	return s.pubsub.Publish(event, string(message))
}

func (s *Service) PublisAccountDepositEvent(
	accountName string, accountBalance map[string]ports.Balance,
	deposit domain.Deposit, market ports.Market,
) error {
	event := eventAccountDeposit
	account := getAccountPayload(accountName, market)
	balance := getBalancePayload(accountBalance)
	payload := map[string]interface{}{
		"event":            event,
		"account":          account,
		"balance":          balance,
		"txid":             deposit.TxID,
		"amount_deposited": deposit.TotAmountPerAsset,
	}
	message, _ := json.Marshal(payload)
	if err := s.pubsub.Publish(event, string(message)); err != nil {
		return err
	}
	return nil
}

func (s *Service) PublisAccountWithdrawEvent(
	accountName string, accountBalance map[string]ports.Balance,
	withdrawal domain.Withdrawal, market ports.Market,
) error {
	event := eventAccountWithdraw
	account := getAccountPayload(accountName, market)
	balance := getBalancePayload(accountBalance)
	payload := map[string]interface{}{
		"event":            event,
		"account":          account,
		"balance":          balance,
		"txid":             withdrawal.TxID,
		"amount_withdrawn": withdrawal.TotAmountPerAsset,
	}
	message, _ := json.Marshal(payload)
	if err := s.pubsub.Publish(event, string(message)); err != nil {
		return err
	}
	return nil
}

func (s *Service) PublishTradeSettledEvent(
	accountName string, accountBalance map[string]ports.Balance,
	trade domain.Trade,
) error {
	event := eventTradeSettled
	market := getMarketPayload(trade)
	balance := getBalancePayload(accountBalance)
	swap := getSwapPayload(trade.SwapRequestMessage())
	payload := map[string]interface{}{
		"event":   event,
		"market":  market,
		"balance": balance,
		"swap":    swap,
		"price": map[string]string{
			"base_price":  trade.MarketPrice.BasePrice,
			"quote_price": trade.MarketPrice.QuotePrice,
		},
		"txid":                 trade.TxId,
		"settlement_timestamp": trade.SettlementTime,
		"settlement_date":      time.Unix(int64(trade.SettlementTime), 0).Format(time.RFC3339),
		"trading_fees": map[string]interface{}{
			"asset":  trade.FeeAsset,
			"amount": trade.FeeAmount,
		},
	}
	message, _ := json.Marshal(payload)
	if err := s.pubsub.Publish(event, string(message)); err != nil {
		return err
	}
	return nil
}

func (s *Service) Close() {
	s.pubsub.Store().Close()
}

type webhookInfo struct {
	ports.Subscription
}

func (i webhookInfo) GetId() string {
	return i.Subscription.Id()
}
func (i webhookInfo) GetEvent() ports.WebhookEvent {
	return webhookEventInfo(i.Subscription.Topic())
}
func (i webhookInfo) GetEndpoint() string {
	return i.Subscription.NotifyAt()
}
func (i webhookInfo) IsSecured() bool {
	return i.Subscription.IsSecured()
}

type webhookEventInfo string

func (i webhookEventInfo) IsUnspecified() bool {
	return i == ports.UnspecifiedTopic
}
func (i webhookEventInfo) IsTradeSettled() bool {
	return i == eventTradeSettled
}
func (i webhookEventInfo) IsAccountLowBalance() bool {
	return i == eventAccountLowBalance
}
func (i webhookEventInfo) IsAccountWithdraw() bool {
	return i == eventAccountWithdraw
}
func (i webhookEventInfo) IsAccountDeposit() bool {
	return i == eventAccountDeposit
}
func (i webhookEventInfo) IsAny() bool {
	return i == ports.AnyTopic
}
