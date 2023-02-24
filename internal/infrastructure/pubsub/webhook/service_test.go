package webhookpubsub_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	webhookpubsub "github.com/tdex-network/tdex-daemon/internal/infrastructure/pubsub/webhook"
	"github.com/tdex-network/tdex-daemon/pkg/securestore"
	boltsecurestore "github.com/tdex-network/tdex-daemon/pkg/securestore/bolt"
)

var (
	datadir     = "webhooktest"
	filename    = "test.db"
	password    = "password"
	serverPort  = "8888"
	serverURL   = fmt.Sprintf("http://localhost:%s", serverPort)
	testMessage = `{"txid":"0000000000000000000000000000000000000000000000000000000000000000","swap":{"amount_p":10000,"asset_p":"LBTC","amount_r":450000000,"asset_r":"USDT"},"price":{"base_price":"0.000025","quote_price":"40000"}}`

	tradesettleEndpoint = fmt.Sprintf("%s/tradesettle", serverURL)
	allactionsEndpoint  = fmt.Sprintf("%s/allactions", serverURL)
)

func TestWebhookPubSubService(t *testing.T) {
	pubsubSvc, err := newTestService()
	require.NoError(t, err)

	server := newTestWebServer(t)

	t.Cleanup(func() {
		//nolint
		server.Shutdown(context.TODO())
		//nolint
		pubsubSvc.Store().Close()
		os.RemoveAll(datadir)
	})

	// Start the webserver whose endpoints will be invoked by the pubsubSvc.
	go func() {
		t.Logf("starting web server on port %s", serverPort)
		err := server.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				t.Log("closed web server")
				return
			}
			t.Error(err)
		}
	}()

	// Ensures precondition: if not initialized, the store is also locked.
	require.True(t, pubsubSvc.Store().IsLocked())

	err = pubsubSvc.Store().Init(password)
	require.NoError(t, err)

	// Ensures Init() initializes and locks the store
	require.True(t, pubsubSvc.Store().IsLocked())

	err = pubsubSvc.Store().Unlock(password)
	require.NoError(t, err)

	require.False(t, pubsubSvc.Store().IsLocked())

	testHooks := newTestHooks()
	for _, hook := range testHooks {
		hookID, err := pubsubSvc.Subscribe(hook.ActionType.String(), hook.Endpoint, hook.Secret)
		require.NoError(t, err)
		require.NotNil(t, hookID)
	}

	subs := pubsubSvc.ListSubscriptionsForTopic(webhookpubsub.TradeSettled.String())
	require.Len(t, subs, len(testHooks))
	require.Condition(t, func() bool {
		for i, expectedHook := range testHooks {
			sub := subs[i]
			if sub.Id() == "" {
				return false
			}
			if sub.NotifyAt() != expectedHook.Endpoint {
				return false
			}
			if len(expectedHook.Secret) > 0 && !sub.IsSecured() {
				return false
			}
		}
		return true
	})

	// Should invoke all hooks.
	err = pubsubSvc.Publish(webhookpubsub.TradeSettled.String(), testMessage)
	require.NoError(t, err)

	for i, s := range subs {
		err := pubsubSvc.Unsubscribe(s.Topic().Label(), s.Id())
		require.NoError(t, err)

		if s.Topic().Code() == webhookpubsub.AllActions.Code() {
			subs := pubsubSvc.ListSubscriptionsForTopic(webhookpubsub.AllActions.String())
			require.Len(t, subs, 0)
		}
		subs := pubsubSvc.ListSubscriptionsForTopic(s.Topic().Label())
		require.Len(t, subs, len(testHooks)-1-i)
	}

	// Checks that it's all ok if there are no hooks to invoke.
	err = pubsubSvc.Publish(webhookpubsub.AccountLowBalance.String(), testMessage)
	require.NoError(t, err)
}

func newTestService() (ports.SecurePubSub, error) {
	store, err := newTestSecureStorage(datadir, filename)
	if err != nil {
		return nil, err
	}
	return webhookpubsub.NewWebhookPubSubService(store)
}

func newTestSecureStorage(datadir, filename string) (securestore.SecureStorage, error) {
	store, err := boltsecurestore.NewSecureStorage(datadir, filename)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func newTestHooks() []*webhookpubsub.Webhook {
	hooksDetails := []struct {
		actionType webhookpubsub.WebhookAction
		endpoint   string
		secret     string
	}{
		{webhookpubsub.TradeSettled, tradesettleEndpoint, randomSecret()},
		{webhookpubsub.TradeSettled, tradesettleEndpoint, randomSecret()},
		{webhookpubsub.TradeSettled, tradesettleEndpoint, randomSecret()},
		{webhookpubsub.AllActions, allactionsEndpoint, ""},
	}
	hooks := make([]*webhookpubsub.Webhook, 0, len(hooksDetails))
	for _, d := range hooksDetails {
		hook, _ := webhookpubsub.NewWebhook(d.actionType, d.endpoint, d.secret)
		hook.ID = ""
		hooks = append(hooks, hook)
	}
	return hooks
}

func newTestWebServer(t *testing.T) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%s", serverPort)}
	handleFn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Bad method", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") == "" {
			http.Error(w, "Missing Content-Type header", http.StatusUnsupportedMediaType)
			return
		}
		// Return response
		fmt.Fprintf(w, "Done")

		// Log request
		defer r.Body.Close()
		payload, _ := io.ReadAll(r.Body)
		headers, _ := json.Marshal(r.Header)
		info := struct {
			method   string
			endpoint string
			payload  string
			headers  string
		}{r.Method, r.URL.String(), string(payload), string(headers)}
		t.Logf("request info: %+v", info)
	}
	http.HandleFunc("/tradesettle", handleFn)
	http.HandleFunc("/allactions", handleFn)
	return srv
}

func randomSecret() string {
	b := make([]byte, 32)
	//nolint
	rand.Read(b)
	return hex.EncodeToString(b)
}
