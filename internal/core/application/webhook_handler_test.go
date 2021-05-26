package application_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/pkg/securestore"
	boltsecurestore "github.com/tdex-network/tdex-daemon/pkg/securestore/bolt"
)

var (
	password         = []byte("password")
	serverPort       = "8888"
	serverURL        = fmt.Sprintf("http://localhost:%s", serverPort)
	payloadForAction = `{"txid":"0000000000000000","swap":{"amount_p":10000,"asset_p":"LBTC","amount_r":450000000,"asset_r":"USDT"},"price":{"base_price":"0.000025","quote_price":"40000"}}`

	tradesettleEndpoint = fmt.Sprintf("%s/tradesettle", serverURL)
	allactionsEndpoint  = fmt.Sprintf("%s/allactions", serverURL)
)

func TestWebhookHandler(t *testing.T) {
	handler, clean, err := newTestWebhookHandler()
	require.NoError(t, err)

	server := newTestWebServer(t)

	t.Cleanup(func() {
		server.Shutdown(context.TODO())
		clean()
	})

	// Start the webserver whose endpoints will be invoked by the handler.
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

	testHooks := newTestHooks()
	for _, hook := range testHooks {
		err := handler.AddWebhook(hook)
		require.NoError(t, err)
	}

	hooks := handler.ListWebhooksForAction(application.TradeSettled)
	require.Len(t, hooks, len(testHooks))
	require.ElementsMatch(t, hooks, testHooks)

	handler.InvokeWebhooksByAction(application.TradeSettled, payloadForAction)

	for i, hook := range testHooks {
		err := handler.RemoveWebhook(hook.Id)
		require.NoError(t, err)

		if hook.ActionType == application.AllActions {
			hooks := handler.ListWebhooksForAction(application.AllActions)
			require.Len(t, hooks, 0)
		}
		hooks := handler.ListWebhooksForAction(hook.ActionType)
		require.Len(t, hooks, len(testHooks)-1-i)
	}
}

func newTestWebhookHandler() (application.WebhookHandler, func(), error) {
	datadir, filename := "webhooktest", "test.db"
	store, err := newTestSecureStorage(datadir, filename)
	if err != nil {
		return nil, nil, err
	}

	webhookHandler, err := application.NewWebhookHandler(store)
	if err != nil {
		return nil, nil, err
	}
	clean := func() {
		webhookHandler.CloseStore()
		os.RemoveAll(datadir)
	}
	return webhookHandler, clean, nil
}

func newTestSecureStorage(datadir, filename string) (securestore.SecureStorage, error) {
	store, err := boltsecurestore.NewSecureStorage(datadir, filename)
	if err != nil {
		return nil, err
	}
	if err := store.CreateUnlock(&password); err != nil {
		return nil, err
	}
	return store, nil
}

func newTestHooks() []*application.Hook {
	hooksDetails := []struct {
		actionType int
		endpoint   string
		secret     string
	}{
		{application.TradeSettled, tradesettleEndpoint, randomSecret()},
		{application.TradeSettled, tradesettleEndpoint, randomSecret()},
		{application.TradeSettled, tradesettleEndpoint, randomSecret()},
		{application.AllActions, allactionsEndpoint, randomSecret()},
	}
	hooks := make([]*application.Hook, 0, len(hooksDetails))
	for _, d := range hooksDetails {
		hook, _ := application.NewWebhook(d.actionType, d.endpoint, d.secret)
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
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "Missing bearer token", http.StatusUnauthorized)
			return
		}
		// Return response
		fmt.Fprintf(w, "Done")

		// Log request
		defer r.Body.Close()
		payload, _ := ioutil.ReadAll(r.Body)
		h := map[string]string{
			"Authorization": r.Header.Get("Authorization"),
			"Content-Type":  r.Header.Get("Content-Type"),
		}
		headers, _ := json.Marshal(h)
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
	rand.Read(b)
	return hex.EncodeToString(b)
}
