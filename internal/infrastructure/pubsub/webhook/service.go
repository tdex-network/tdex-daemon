package webhookpubsub

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/prometheus/common/log"
	"github.com/sony/gobreaker"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/tdex-network/tdex-daemon/pkg/securestore"
	"golang.org/x/sync/errgroup"
)

type webhookService struct {
	store      webhookStore
	httpClient *esplora.Client
	cb         *gobreaker.CircuitBreaker
}

func NewWebhookPubSubService(
	store securestore.SecureStorage,
	httpClient *esplora.Client,
) (application.SecurePubSub, error) {
	if store == nil {
		return nil, ErrNullSecureStore
	}
	if httpClient == nil {
		return nil, ErrNullHTTPClient
	}

	return &webhookService{
		store:      webhookStore{store},
		httpClient: httpClient,
		cb:         newCircuitBreaker(),
	}, nil
}

func (ws *webhookService) Store() application.PubSubStore {
	return ws.store
}

func (ws *webhookService) Subscribe(topic string, args ...interface{}) (string, error) {
	actionType, ok := stringToAction[topic]
	if !ok {
		return "", ErrInvalidTopic
	}
	if len(args) != 2 {
		return "", ErrInvalidArgs
	}
	endpoint, ok := args[0].(string)
	if !ok {
		return "", ErrInvalidArgType
	}
	secret, ok := args[1].(string)
	if !ok {
		return "", ErrInvalidArgType
	}

	hook, err := NewWebhook(actionType, endpoint, secret)
	if err != nil {
		return "", err
	}

	return ws.addWebhook(hook)
}

func (ws *webhookService) Unsubscribe(_, id string) error {
	return ws.removeWebhook(id)
}

func (ws *webhookService) ListSubscriptionsForTopic(topic string) []application.Subscription {
	actionType, ok := WebhookActionFromString(topic)
	if !ok {
		return nil
	}
	return ws.listWebhooksForAction(actionType)
}

func (ws *webhookService) Publish(topic string, message string) error {
	actionType, ok := WebhookActionFromString(topic)
	if !ok {
		return ErrUnknownWebhookAction
	}
	return ws.invokeWebhooksForAction(actionType, message)
}

func (ws *webhookService) TopicsByCode() map[int]application.Topic {
	topics := make(map[int]application.Topic)
	for action := range actionToString {
		topics[int(action)] = action
	}
	return topics
}

func (ws *webhookService) TopicsByLabel() map[string]application.Topic {
	topics := make(map[string]application.Topic)
	for label, action := range stringToAction {
		topics[label] = action
	}
	return topics
}

// AddWebhook adds the provided hook to those managed by the handler.
// If another hook with the same id already exists, the method returns
// preventing overwrites/duplications.
// NOTE: The generation of the hook ID can be assumed enough random to infer
// that if 2 hooks have the same id, then they are the same.
func (ws *webhookService) addWebhook(hook *Webhook) (string, error) {
	hookID := []byte(hook.ID)
	hh, err := ws.store.db().GetFromBucket(hooksBucket, hookID)
	if err != nil {
		return "", err
	}
	// there's already a webhook with the same id, so let's avoid duplicates.
	if hh != nil {
		return hook.ID, nil
	}

	if err := ws.store.db().AddToBucket(hooksBucket, hookID, hook.Serialize()); err != nil {
		return "", err
	}
	ws.addHookByAction(hook)
	return hook.ID, nil
}

// RemoveWebhhìook attempts to remove the hook identified by an ID from those
// managed by the handler. Nothing is done in case the hook does not actually
// exist in the handler's store.db().
func (ws *webhookService) removeWebhook(hookID string) error {
	buf, err := ws.store.db().GetFromBucket(hooksBucket, []byte(hookID))
	if err != nil {
		return err
	}
	if buf == nil {
		return nil
	}

	if err := ws.store.db().RemoveFromBucket(hooksBucket, []byte(hookID)); err != nil {
		return err
	}

	hook, _ := NewWebhookFromBytes(buf)
	ws.removeHookByAction(hook)
	return nil
}

func (ws *webhookService) listWebhooksForAction(actionType WebhookAction) []application.Subscription {
	hooks := ws.getHooksByAction(actionType)
	if actionType != AllActions {
		hooksForAllActions := ws.getHooksByAction(AllActions)
		hooks = append(hooks, hooksForAllActions...)
	}
	subs := make([]application.Subscription, len(hooks))
	for i, h := range hooks {
		subs[i] = h
	}
	return subs
}

// InvokeWebhooksByAction makes a POST request to every webhook endpoint
// registered for the given action.
// This method adopts a circuit breaker approach in order to maximize the
// chances that every webhook gets invoked without errors.
func (ws *webhookService) invokeWebhooksForAction(actionType WebhookAction, message string) error {
	hooks := ws.getHooksByAction(actionType)
	if actionType != AllActions {
		hooksForAllActions := ws.getHooksByAction(AllActions)
		hooks = append(hooks, hooksForAllActions...)
	}

	eg := &errgroup.Group{}
	for i := range hooks {
		hook := hooks[i]
		eg.Go(func() error { return ws.doRequest(hook, message) })
	}
	return eg.Wait()
}

func (ws *webhookService) addHookByAction(hook *Webhook) {
	key := []byte{byte(hook.ActionType)}
	hooks := ws.getSerializedHooks(hook.ActionType)
	hooks = append(hooks, hook.Serialize())
	updatedHooks := bytes.Join(hooks, separator)
	ws.store.db().AddToBucket(hooksByActionBucket, key, updatedHooks)
}

func (ws *webhookService) removeHookByAction(hook *Webhook) {
	rawHooks := ws.getSerializedHooks(hook.ActionType)

	var index int
	for i, buf := range rawHooks {
		hh, _ := NewWebhookFromBytes(buf)
		if hh.ID == hook.ID {
			index = i
			break
		}
	}

	key := []byte{byte(hook.ActionType)}
	rawHooks = append(rawHooks[:index], rawHooks[index+1:]...)

	if len(rawHooks) <= 0 {
		ws.store.db().RemoveFromBucket(hooksByActionBucket, key)
		return
	}

	updatedHooks := bytes.Join(rawHooks, separator)
	ws.store.db().AddToBucket(hooksByActionBucket, key, updatedHooks)
}

func (ws *webhookService) getHooksByAction(actionType WebhookAction) []*Webhook {
	rawHooks := ws.getSerializedHooks(actionType)
	hooks := make([]*Webhook, 0, len(rawHooks))
	for _, buf := range rawHooks {
		hook, _ := NewWebhookFromBytes(buf)
		hooks = append(hooks, hook)
	}
	return hooks
}

func (ws *webhookService) getSerializedHooks(actionType WebhookAction) [][]byte {
	key := []byte{byte(actionType)}
	hooks, _ := ws.store.db().GetFromBucket(hooksByActionBucket, key)
	if len(hooks) <= 0 {
		return nil
	}
	return bytes.Split(hooks, separator)
}

func (ws *webhookService) doRequest(hook *Webhook, payload string) error {
	_, err := ws.cb.Execute(func() (interface{}, error) {
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		if hook.IsSecured() {
			token := jwt.New(jwt.SigningMethodHS256)
			secret := []byte(hook.Secret)
			tokenString, _ := token.SignedString(secret)
			headers["Authorization"] = fmt.Sprintf("Bearer %s", tokenString)
		}

		status, resp, err := ws.httpClient.NewHTTPRequest("POST", hook.Endpoint, payload, headers)
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

func newCircuitBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "explorer",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests > 20 && failureRatio >= 0.7
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			if to == gobreaker.StateOpen {
				log.Warn("explorer seems down, stop allowing requests")
			}
			if from == gobreaker.StateOpen && to == gobreaker.StateHalfOpen {
				log.Info("checking explorer status")
			}
			if from == gobreaker.StateHalfOpen && to == gobreaker.StateClosed {
				log.Info("explorer seems ok, restart allowing requests")
			}
		},
	})
}
