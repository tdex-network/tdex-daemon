package pubsub

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type Subscription struct {
	ID       string `json:"id"`
	Event    string `json:"event"`
	Endpoint string `json:"endpoint"`
	Secret   string `json:"secret"`
}

type subscriptions []Subscription

func (s subscriptions) toPortable() []ports.Subscription {
	subs := make([]ports.Subscription, 0, len(s))
	for i := range s {
		sub := s[i]
		subs = append(subs, &sub)
	}
	return subs
}

func NewSubscription(event, endpoint, secret string) (*Subscription, error) {
	if len(event) <= 0 {
		return nil, fmt.Errorf("missing event")
	}
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return nil, fmt.Errorf("invalid webhook endpoint, must be a vald URI")
	}
	id := uuid.New().String()
	return &Subscription{id, event, endpoint, secret}, nil
}

func NewSubscriptionFromBytes(buf []byte) (*Subscription, error) {
	sub := &Subscription{}
	if err := json.Unmarshal(buf, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (h *Subscription) Topic() string {
	return h.Event
}

func (h *Subscription) Id() string {
	return h.ID
}

func (h *Subscription) NotifyAt() string {
	return h.Endpoint
}

func (h *Subscription) IsSecured() bool {
	return len(h.Secret) > 0
}

func (h *Subscription) Serialize() []byte {
	b, _ := json.Marshal(*h)
	return b
}
