package application

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/tdex-network/tdex-daemon/pkg/securestore"
	"golang.org/x/sync/errgroup"
)

var (
	hooksBucket         = []byte("hooks")
	hooksByActionBucket = []byte("hooksbyaction")

	// separator equivalent character is ÿ.
	// Should be fine to use such value  since it's not used for Secret (jwt
	// base64-encoded token), nor for Endpoint (http url).
	separator = []byte{255}
)

type Hook struct {
	Id         string `json:"id"`
	ActionType int    `json:"actionType"`
	Endpoint   string `json:"endpoint"`
	Secret     string `json:"secret"`
}

func NewWebhook(actionType int, endpoint, secret string) (*Hook, error) {
	if actionType < TradeSettled || actionType > AllActions {
		return nil, fmt.Errorf("action is of unknown type")
	}
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return nil, fmt.Errorf("webhook endpoint must be a vald URI")
	}
	id := uuid.New().String()
	return &Hook{id, actionType, endpoint, secret}, nil
}

func NewWebhookFromBytes(buf []byte) (*Hook, error) {
	h := &Hook{}
	if err := json.Unmarshal(buf, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Hook) IsSecured() bool {
	return len(h.Secret) > 0
}

func (h *Hook) Serialize() []byte {
	b, _ := json.Marshal(*h)
	return b
}

type WebhookHandler interface {
	// Init initialize the internal store
	Init(password string) error
	// LockStore locks the internal store.
	LockStore()
	// CloseStore closes the connections with the internal store.
	CloseStore() error
	// UnlockStore unlocks the internal store.
	UnlockStore(password string) error
	ChangePassword(oldPwd, newPwd string) error

	// AddWebhook adds the provided webhook to the interal store.
	AddWebhook(hook *Hook) error
	// RemoveWebhook removes the hook identified by an ID from the internal store
	// if existing.
	RemoveWebhook(hookID string) error
	// ListWebhooksForAction returns the list of all webhooks registerd for the
	// give action type, included those registered for AllActions.
	ListWebhooksForAction(actionType int) []*Hook
	// InvokeWebhookByAction makes a POST request with the given stringified JSON
	// payload for any webhook registered for the given action, included those
	// registered for AllActions.
	InvokeWebhooksByAction(actionType int, payload string) error
}

type webhookHandler struct {
	store      securestore.SecureStorage
	httpClient *esplora.Client
	cb         *gobreaker.CircuitBreaker
}

func NewWebhookHandler(store securestore.SecureStorage) (WebhookHandler, error) {
	return &webhookHandler{
		store:      store,
		httpClient: esplora.NewHTTPClient(15 * time.Second),
		cb:         newCircuitBreaker(),
	}, nil
}

func (h *webhookHandler) Init(password string) error {
	if err := h.UnlockStore(password); err != nil {
		return err
	}
	defer h.LockStore()

	if err := h.store.CreateBucket(hooksBucket); err != nil {
		return err
	}
	if err := h.store.CreateBucket(hooksByActionBucket); err != nil {
		return err
	}
	return nil
}

func (h *webhookHandler) LockStore() {
	h.store.Lock()
}

func (h *webhookHandler) UnlockStore(password string) error {
	pwd := []byte(password)
	return h.store.CreateUnlock(&pwd)
}

func (h *webhookHandler) CloseStore() error {
	return h.store.Close()
}

func (h *webhookHandler) ChangePassword(oldPwd, newPwd string) error {
	old := []byte(oldPwd)
	new := []byte(newPwd)
	return h.store.ChangePassword(old, new)
}

// AddWebhook adds the provided hook to those managed by the handler.
// If another hook with the same id already exists, the method returns
// preventing overwrites/duplications.
// NOTE: The generation of the hook ID can be assumed enough random to infer
// that if 2 hooks have the same id, then they are the same.
func (h *webhookHandler) AddWebhook(hook *Hook) error {
	hookID := []byte(hook.Id)
	hh, err := h.store.GetFromBucket(hooksBucket, hookID)
	if err != nil {
		return err
	}
	// there's already a webhook with the same id, so let's avoid duplicates.
	if hh != nil {
		return nil
	}

	if err := h.store.AddToBucket(hooksBucket, hookID, hook.Serialize()); err != nil {
		return err
	}
	h.addHookByAction(hook)
	return nil
}

// RemoveWebhhìook attempts to remove the hook identified by an ID from those
// managed by the handler. Nothing is done in case the hook does not actually
// exist in the handler's store.
func (h *webhookHandler) RemoveWebhook(hookID string) error {
	buf, err := h.store.GetFromBucket(hooksBucket, []byte(hookID))
	if err != nil {
		return err
	}
	if buf == nil {
		return nil
	}

	if err := h.store.RemoveFromBucket(hooksBucket, []byte(hookID)); err != nil {
		return err
	}

	hook, _ := NewWebhookFromBytes(buf)
	h.removeHookByAction(hook)
	return nil
}

func (h *webhookHandler) ListWebhooksForAction(actionType int) []*Hook {
	hooks := h.getHooksByAction(actionType)
	if actionType != AllActions {
		hooksForAllActions := h.getHooksByAction(AllActions)
		hooks = append(hooks, hooksForAllActions...)
	}
	return hooks
}

// InvokeWebhooksByAction makes a POST request to every webhook endpoint
// registered for the given action.
// This method adopts a circuit breaker approach in order to maximize the
// chances that every webhook gets invoked without errors.
func (h *webhookHandler) InvokeWebhooksByAction(actionType int, payload string) error {
	hooks := h.getHooksByAction(actionType)
	if actionType != AllActions {
		hooksForAllActions := h.getHooksByAction(AllActions)
		hooks = append(hooks, hooksForAllActions...)
	}

	eg := &errgroup.Group{}
	for i := range hooks {
		hook := hooks[i]
		eg.Go(func() error { return h.doRequest(hook, payload) })
	}
	return eg.Wait()
}

func (h *webhookHandler) addHookByAction(hook *Hook) {
	key := []byte{byte(hook.ActionType)}
	hooks := h.getSerializedHooks(hook.ActionType)
	hooks = append(hooks, hook.Serialize())
	updatedHooks := bytes.Join(hooks, separator)
	h.store.AddToBucket(hooksByActionBucket, key, updatedHooks)
}

func (h *webhookHandler) removeHookByAction(hook *Hook) {
	rawHooks := h.getSerializedHooks(hook.ActionType)

	var index int
	for i, buf := range rawHooks {
		hh, _ := NewWebhookFromBytes(buf)
		if hh.Id == hook.Id {
			index = i
			break
		}
	}

	key := []byte{byte(hook.ActionType)}
	rawHooks = append(rawHooks[:index], rawHooks[index+1:]...)

	if len(rawHooks) <= 0 {
		h.store.RemoveFromBucket(hooksByActionBucket, key)
		return
	}

	updatedHooks := bytes.Join(rawHooks, separator)
	h.store.AddToBucket(hooksByActionBucket, key, updatedHooks)
}

func (h *webhookHandler) getHooksByAction(actionType int) []*Hook {
	rawHooks := h.getSerializedHooks(actionType)
	hooks := make([]*Hook, 0, len(rawHooks))
	for _, buf := range rawHooks {
		hook, _ := NewWebhookFromBytes(buf)
		hooks = append(hooks, hook)
	}
	return hooks
}

func (h *webhookHandler) getSerializedHooks(actionType int) [][]byte {
	key := []byte{byte(actionType)}
	hooks, _ := h.store.GetFromBucket(hooksByActionBucket, key)
	if len(hooks) <= 0 {
		return nil
	}
	return bytes.Split(hooks, separator)
}

func (h *webhookHandler) doRequest(hook *Hook, payload string) error {
	_, err := h.cb.Execute(func() (interface{}, error) {
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		if hook.IsSecured() {
			token := jwt.New(jwt.SigningMethodHS256)
			secret, _ := hex.DecodeString(hook.Secret)
			tokenString, _ := token.SignedString(secret)
			headers["Authorization"] = fmt.Sprintf("Bearer %s", tokenString)
		}

		status, resp, err := h.httpClient.NewHTTPRequest("POST", hook.Endpoint, payload, headers)
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

var webhookManager WebhookHandler

func InitWebhookManager(store securestore.SecureStorage) (clean func(), err error) {
	webhookManager, err = NewWebhookHandler(store)
	if err != nil {
		return
	}
	clean = func() {
		webhookManager.CloseStore()
	}
	return
}
