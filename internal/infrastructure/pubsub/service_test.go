package pubsub_test

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
	pubsub "github.com/tdex-network/tdex-daemon/internal/infrastructure/pubsub"
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
	alleventsEndpoint   = fmt.Sprintf("%s/allevents", serverURL)
)

func TestPubSubService(t *testing.T) {
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

	testSubs := newTestSubs()
	for _, sub := range testSubs {
		subID, err := pubsubSvc.Subscribe(sub.Topic(), sub.Endpoint, sub.Secret)
		require.NoError(t, err)
		require.NotNil(t, subID)
	}

	subs := pubsubSvc.ListSubscriptionsForTopic("test")
	require.Len(t, subs, len(testSubs))
	require.Condition(t, func() bool {
		for i, expectedSub := range testSubs {
			sub := subs[i]
			if sub.Id() == "" {
				return false
			}
			if sub.NotifyAt() != expectedSub.Endpoint {
				return false
			}
			if len(expectedSub.Secret) > 0 && !sub.IsSecured() {
				return false
			}
		}
		return true
	})

	// Should invoke all hooks.
	err = pubsubSvc.Publish("test", testMessage)
	require.NoError(t, err)

	for i, s := range subs {
		err := pubsubSvc.Unsubscribe(s.Topic(), s.Id())
		require.NoError(t, err)

		if s.Topic() == ports.AnyTopic {
			subs := pubsubSvc.ListSubscriptionsForTopic(ports.AnyTopic)
			require.Len(t, subs, 0)
		}
		subs := pubsubSvc.ListSubscriptionsForTopic(s.Topic())
		require.Len(t, subs, len(testSubs)-1-i)
	}

	// Checks that it's all ok if there are no hooks to invoke.
	err = pubsubSvc.Publish("test1", testMessage)
	require.NoError(t, err)
}

func newTestService() (ports.SecurePubSub, error) {
	store, err := newTestSecureStorage(datadir, filename)
	if err != nil {
		return nil, err
	}
	return pubsub.NewService(store)
}

func newTestSecureStorage(datadir, filename string) (securestore.SecureStorage, error) {
	store, err := boltsecurestore.NewSecureStorage(datadir, filename)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func newTestSubs() []*pubsub.Subscription {
	subsDetails := []struct {
		topic    string
		endpoint string
		secret   string
	}{
		{"test", tradesettleEndpoint, randomSecret()},
		{"test", tradesettleEndpoint, randomSecret()},
		{"test", tradesettleEndpoint, randomSecret()},
		{"*", alleventsEndpoint, ""},
	}
	subs := make([]*pubsub.Subscription, 0, len(subsDetails))
	for _, d := range subsDetails {
		sub, _ := pubsub.NewSubscription(d.topic, d.endpoint, d.secret)
		sub.ID = ""
		subs = append(subs, sub)
	}
	return subs
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
	http.HandleFunc("/allevents", handleFn)
	return srv
}

func randomSecret() string {
	b := make([]byte, 32)
	//nolint
	rand.Read(b)
	return hex.EncodeToString(b)
}
