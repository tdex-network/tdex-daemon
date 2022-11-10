package pubsub

import (
	"encoding/json"
	"time"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

const (
	tradeSettled = iota
	accountLowBalance
	accountWithdraw
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

func (s *Service) PublisAccountLowBalanceTopic(
	accountName string, accountBalance map[string]ports.Balance,
	market ports.Market,
) error {
	topics := s.pubsub.TopicsByCode()
	topic := topics[accountLowBalance]
	event := getEventPayload(topic)
	account := getAccountPayload(accountName, market)
	balance := getBalancePayload(accountBalance)
	payload := map[string]interface{}{
		"event":   event,
		"account": account,
		"balance": balance,
	}
	message, _ := json.Marshal(payload)

	if err := s.pubsub.Publish(topic.Label(), string(message)); err != nil {
		return err
	}
	return nil
}

func (s *Service) PublisAccountWithdrawTopic(
	accountName string, accountBalance map[string]ports.Balance,
	withdrawal domain.Withdrawal, market ports.Market,
) error {
	topics := s.pubsub.TopicsByCode()
	topic := topics[accountLowBalance]
	event := getEventPayload(topic)
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
	if err := s.pubsub.Publish(topic.Label(), string(message)); err != nil {
		return err
	}
	return nil
}

func (s *Service) PublishTradeSettledTopic(
	accountName string, accountBalance map[string]ports.Balance,
	trade domain.Trade,
) error {
	topics := s.pubsub.TopicsByCode()
	topic := topics[tradeSettled]
	event := getEventPayload(topic)
	market := getMarketPayload(trade)
	balance := getBalancePayload(accountBalance)
	swap := getSwapPayload(trade.SwapRequestMessage())
	payload := map[string]interface{}{
		"event":                event,
		"market":               market,
		"balance":              balance,
		"swap":                 swap,
		"price":                trade.MarketPrice,
		"txid":                 trade.TxId,
		"settlement_timestamp": trade.SettlementTime,
		"settlement_date":      time.Unix(int64(trade.SettlementTime), 0).Format(time.RFC3339),
	}
	message, _ := json.Marshal(payload)
	if err := s.pubsub.Publish(topic.Label(), string(message)); err != nil {
		return err
	}
	return nil
}

func (s *Service) Close() {
	s.pubsub.Store().Close()
}
