package v1subscription

import (
	"bytes"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/securestore"
	boltsecurestore "github.com/tdex-network/tdex-daemon/pkg/securestore/bolt"
)

const (
	EventTradeSettled      = "TRADE_SETTLED"
	EventAccountLowBalance = "ACCOUNT_LOW_BALANCE"
	EventAccountWithdraw   = "ACCOUNT_WITHDRAW"
	EventAccountDeposit    = "ACCOUNT_DEPOSIT"
	AnyTopic               = "*"
)

var (
	subsBucket        = []byte("subscriptionRepository")
	subsByEventBucket = []byte("subscriptionsbyevent")

	separator = []byte{255}
)

type Repository interface {
	InsertSubscriptions(subscriptions []Subscription) error
}

type subscriptionRepository struct {
	store securestore.SecureStorage
}

func NewSubscriptionRepository(datadir string) (Repository, error) {
	store, err := boltsecurestore.NewSecureStorage(datadir, "pubsub.db")
	if err != nil {
		return nil, err
	}
	return &subscriptionRepository{
		store: store,
	}, nil
}

func (s *subscriptionRepository) InsertSubscriptions(subscriptions []Subscription) error {
	for _, sub := range subscriptions {
		if _, err := s.addSubscription(&sub); err != nil {
			return err
		}
	}

	return nil
}

func (s *subscriptionRepository) addSubscription(sub *Subscription) (string, error) {
	subID := []byte(sub.ID)
	ss, err := s.store.GetFromBucket(subsBucket, subID)
	if err != nil {
		return "", err
	}
	if ss != nil {
		return sub.ID, nil
	}

	if err := s.store.AddToBucket(subsBucket, subID, sub.Serialize()); err != nil {
		return "", err
	}
	if err := s.addSubscriptionForTopic(sub); err != nil {
		return "", err
	}
	return sub.ID, nil
}

func (s *subscriptionRepository) addSubscriptionForTopic(sub *Subscription) error {
	key := []byte(sub.Event)
	subs := s.getSerializedSubscriptions(sub.Event)
	subs = append(subs, sub.Serialize())
	updatedHooks := bytes.Join(subs, separator)
	return s.store.AddToBucket(subsByEventBucket, key, updatedHooks)
}

func (s *subscriptionRepository) getSerializedSubscriptions(topic string) [][]byte {
	if topic == ports.UnspecifiedTopic {
		subs := make([][]byte, 0)
		subsByTopic, _ := s.store.GetAllFromBucket(subsByEventBucket)
		for _, list := range subsByTopic {
			subs = append(subs, bytes.Split(list, separator)...)
		}
		return subs
	}

	key := []byte(topic)
	subs, _ := s.store.GetFromBucket(subsByEventBucket, key)
	if len(subs) <= 0 {
		return nil
	}
	return bytes.Split(subs, separator)
}
