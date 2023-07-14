package v0webhook

import (
	"encoding/json"

	"github.com/tdex-network/tdex-daemon/pkg/securestore"
	boltsecurestore "github.com/tdex-network/tdex-daemon/pkg/securestore/bolt"
)

var (
	hooksBucket  = []byte("hooks")
	pubSubDbFile = "pubsub.db"
)

type Repository interface {
	GetAllWebhooks() ([]*Webhook, error)
	Unlock(password string) error
}

type webhookRepository struct {
	store securestore.SecureStorage
}

func NewWebhookRepository(datadir string) (Repository, error) {
	store, err := boltsecurestore.NewSecureStorage(datadir, pubSubDbFile)
	if err != nil {
		return nil, err
	}
	return &webhookRepository{
		store: store,
	}, nil
}

func (w *webhookRepository) GetAllWebhooks() ([]*Webhook, error) {
	valuesByKey, err := w.store.GetAllFromBucket(hooksBucket)
	if err != nil {
		return nil, err
	}

	webhooks := make([]*Webhook, 0, len(valuesByKey))
	for _, value := range valuesByKey {
		webhook, err := newWebhookFromBytes(value)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, webhook)
	}

	return webhooks, nil
}

func (w *webhookRepository) Unlock(password string) error {
	pwd := []byte(password)
	return w.store.CreateUnlock(&pwd)
}

func newWebhookFromBytes(buf []byte) (*Webhook, error) {
	h := &Webhook{}
	if err := json.Unmarshal(buf, h); err != nil {
		return nil, err
	}
	return h, nil
}
