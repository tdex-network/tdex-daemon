package pubsub

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sony/gobreaker"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/circuitbreaker"
	"github.com/tdex-network/tdex-daemon/pkg/securestore"
	"golang.org/x/sync/errgroup"
)

type service struct {
	store      store
	httpClient *client
	cb         *gobreaker.CircuitBreaker
}

func NewService(
	secureStore securestore.SecureStorage,
) (ports.SecurePubSub, error) {
	if secureStore == nil {
		return nil, fmt.Errorf("missing secure store")
	}

	return &service{
		store:      store{secureStore},
		httpClient: newHTTPClient(15 * time.Second),
		cb:         circuitbreaker.NewCircuitBreaker(),
	}, nil
}

func (ws *service) Store() ports.PubSubStore {
	return ws.store
}

func (ws *service) Subscribe(topic, endpoint, secret string) (string, error) {
	sub, err := NewSubscription(topic, endpoint, secret)
	if err != nil {
		return "", err
	}

	return ws.addSubscription(sub)
}

func (ws *service) Unsubscribe(_, id string) error {
	return ws.removeSubscription(id)
}

func (ws *service) ListSubscriptionsForTopic(topic string) []ports.Subscription {
	return ws.listSubscriptionsForTopic(topic).toPortable()
}

func (ws *service) Publish(topic string, message string) error {
	return ws.publishForTopic(topic, message)
}

func (ws *service) addSubscription(sub *Subscription) (string, error) {
	subID := []byte(sub.ID)
	ss, err := ws.store.db().GetFromBucket(subsBucket, subID)
	if err != nil {
		return "", err
	}
	if ss != nil {
		return sub.ID, nil
	}

	if err := ws.store.db().AddToBucket(subsBucket, subID, sub.Serialize()); err != nil {
		return "", err
	}
	if err := ws.addSubscriptionForTopic(sub); err != nil {
		return "", err
	}
	return sub.ID, nil
}

func (ws *service) removeSubscription(subID string) error {
	buf, err := ws.store.db().GetFromBucket(subsBucket, []byte(subID))
	if err != nil {
		return err
	}
	if buf == nil {
		return fmt.Errorf("webhook not found")
	}

	if err := ws.store.db().RemoveFromBucket(subsBucket, []byte(subID)); err != nil {
		return err
	}

	sub, _ := NewSubscriptionFromBytes(buf)
	return ws.removeSubscriptionForTopic(sub)
}

func (ws *service) listSubscriptionsForTopic(topic string) subscriptions {
	subs := ws.getSubscriptionsForTopic(topic)
	if topic != ports.AnyTopic && topic != ports.UnspecifiedTopic {
		subsForAnyTopic := ws.getSubscriptionsForTopic(ports.AnyTopic)
		subs = append(subs, subsForAnyTopic...)
	}
	return subs
}

func (ws *service) publishForTopic(topic, message string) error {
	subs := ws.listSubscriptionsForTopic(topic)

	eg := &errgroup.Group{}
	for i := range subs {
		sub := subs[i]
		eg.Go(func() error { return ws.doRequest(sub, message) })
	}
	return eg.Wait()
}

func (ws *service) addSubscriptionForTopic(sub *Subscription) error {
	key := []byte(sub.Event)
	subs := ws.getSerializedSubscriptions(sub.Event)
	subs = append(subs, sub.Serialize())
	updatedHooks := bytes.Join(subs, separator)
	return ws.store.db().AddToBucket(subsByEventBucket, key, updatedHooks)
}

func (ws *service) removeSubscriptionForTopic(sub *Subscription) error {
	subs := ws.getSerializedSubscriptions(sub.Event)

	var index int
	for i, buf := range subs {
		ss, _ := NewSubscriptionFromBytes(buf)
		if ss.ID == sub.ID {
			index = i
			break
		}
	}

	key := []byte(sub.Event)
	subs = append(subs[:index], subs[index+1:]...)

	if len(subs) <= 0 {
		return ws.store.db().RemoveFromBucket(subsByEventBucket, key)
	}

	updatedHooks := bytes.Join(subs, separator)
	return ws.store.db().AddToBucket(subsByEventBucket, key, updatedHooks)
}

func (ws *service) getSubscriptionsForTopic(topic string) subscriptions {
	rawSubs := ws.getSerializedSubscriptions(topic)
	subs := make(subscriptions, 0, len(rawSubs))
	for _, buf := range rawSubs {
		sub, _ := NewSubscriptionFromBytes(buf)
		subs = append(subs, *sub)
	}
	sort.SliceStable(subs, func(i, j int) bool {
		return subs[i].ID < subs[j].ID
	})
	return subs
}

func (ws *service) getSerializedSubscriptions(topic string) [][]byte {
	if topic == ports.UnspecifiedTopic {
		subs := make([][]byte, 0)
		subsByTopic, _ := ws.store.db().GetAllFromBucket(subsByEventBucket)
		for _, list := range subsByTopic {
			subs = append(subs, bytes.Split(list, separator)...)
		}
		return subs
	}

	key := []byte(topic)
	subs, _ := ws.store.db().GetFromBucket(subsByEventBucket, key)
	if len(subs) <= 0 {
		return nil
	}
	return bytes.Split(subs, separator)
}

func (ws *service) doRequest(sub Subscription, payload string) error {
	_, err := ws.cb.Execute(func() (interface{}, error) {
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		if sub.IsSecured() {
			token := jwt.New(jwt.SigningMethodHS256)
			secret := []byte(sub.Secret)
			tokenString, _ := token.SignedString(secret)
			headers["Authorization"] = fmt.Sprintf("Bearer %s", tokenString)
		}

		status, resp, err := ws.httpClient.post(sub.Endpoint, payload, headers)
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf(resp)
		}
		return nil, nil
	})

	return err
}
