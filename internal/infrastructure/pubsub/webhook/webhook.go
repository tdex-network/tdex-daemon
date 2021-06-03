package webhookpubsub

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
)

type Webhook struct {
	ID         string        `json:"id"`
	ActionType WebhookAction `json:"action_type"`
	Endpoint   string        `json:"endpoint"`
	Secret     string        `json:"secret"`
}

func NewWebhook(actionType WebhookAction, endpoint, secret string) (*Webhook, error) {
	if actionType < TradeSettled || actionType > AllActions {
		return nil, fmt.Errorf("action is of unknown type")
	}
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return nil, fmt.Errorf("webhook endpoint must be a vald URI")
	}
	id := uuid.New().String()
	return &Webhook{id, actionType, endpoint, secret}, nil
}

func NewWebhookFromBytes(buf []byte) (*Webhook, error) {
	h := &Webhook{}
	if err := json.Unmarshal(buf, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Webhook) Topic() application.Topic {
	return h.ActionType
}

func (h *Webhook) Id() string {
	return h.ID
}

func (h *Webhook) NotifyAt() string {
	return h.Endpoint
}

func (h *Webhook) IsSecured() bool {
	return len(h.Secret) > 0
}

func (h *Webhook) Serialize() []byte {
	b, _ := json.Marshal(*h)
	return b
}
