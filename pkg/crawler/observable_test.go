package crawler_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"golang.org/x/time/rate"
)

var (
	observationInterval = 2 * time.Second
	rateLimiter         = newTestRateLimiter()
	wantUtxoSpent       = true
	wantTxConfirmed     = true
)

func TestObservables(t *testing.T) {
	tests := []struct {
		name                  string
		observable            crawler.Observable
		mockedExplorer        explorer.Service
		expectedObservableKey string
		expectedEventType     crawler.EventType
	}{
		{
			name:                  "AddressObservable emits AddressUnspents",
			observable:            crawler.NewAddressObservable("el1qqdhke3shlafa8xrmxrp9lwpw063w0w9eqg4m4w86t6u986xt890wcztxj80rm395mvdq7zja6yzt4z539knpw3g588su6mnkn", h2b("86967925d6603ac1bf839ca7f3c7bf7879918bf48104b5ba3493acba8b292817"), nil),
			mockedExplorer:        mockedExplorerForAddressObs(),
			expectedObservableKey: "el1qqdhke3shlafa8xrmxrp9lwpw063w0w9eqg4m4w86t6u986xt890wcztxj80rm395mvdq7zja6yzt4z539knpw3g588su6mnkn",
			expectedEventType:     crawler.AddressUnspents,
		},
		{
			name:                  fmt.Sprintf("TransactionObservable emits %s", crawler.TransactionUnconfirmed),
			observable:            crawler.NewTransactionObservable("560d912df33521da808dc1f7d43a894ba7221af352328cda3f3b2ec894510477", nil),
			mockedExplorer:        mockedExplorerForTransactionObs(!wantTxConfirmed),
			expectedObservableKey: "560d912df33521da808dc1f7d43a894ba7221af352328cda3f3b2ec894510477",
			expectedEventType:     crawler.TransactionConfirmed,
		},
		{
			name:                  fmt.Sprintf("TransactionObservable emits %s", crawler.TransactionConfirmed),
			observable:            crawler.NewTransactionObservable("69fe1192a74e9c9a874ac1a1f80c244ce801b359e86f3c8d08084f93844e3845", nil),
			mockedExplorer:        mockedExplorerForTransactionObs(wantTxConfirmed),
			expectedObservableKey: "69fe1192a74e9c9a874ac1a1f80c244ce801b359e86f3c8d08084f93844e3845",
			expectedEventType:     crawler.TransactionConfirmed,
		},
		{
			name:                  fmt.Sprintf("OutpointsObservable emits %s", crawler.OutpointsUnspent),
			observable:            crawler.NewOutpointsObservable(mockedOutpoints(2), nil),
			mockedExplorer:        mockedExplorerForOutpointsObs(!wantUtxoSpent, !wantTxConfirmed),
			expectedObservableKey: "",
			expectedEventType:     crawler.OutpointsUnspent,
		},
		{
			name:                  fmt.Sprintf("OutpointsObservable emits %s", crawler.OutpointsSpentAndUnconfirmed),
			observable:            crawler.NewOutpointsObservable(mockedOutpoints(2), nil),
			mockedExplorer:        mockedExplorerForOutpointsObs(wantUtxoSpent, !wantTxConfirmed),
			expectedObservableKey: "",
			expectedEventType:     crawler.OutpointsSpentAndUnconfirmed,
		},
		{
			name:                  fmt.Sprintf("OutpointsObservable emits %s", crawler.OutpointsSpentAndConfirmed),
			observable:            crawler.NewOutpointsObservable(mockedOutpoints(2), nil),
			mockedExplorer:        mockedExplorerForOutpointsObs(wantUtxoSpent, wantTxConfirmed),
			expectedObservableKey: "",
			expectedEventType:     crawler.OutpointsSpentAndConfirmed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checkObservableKey := len(tt.expectedObservableKey) > 0
			if checkObservableKey {
				require.Equal(t, tt.expectedObservableKey, tt.observable.Key())
			}

			wg := &sync.WaitGroup{}
			eventChan := make(chan crawler.Event)
			errChan := make(chan error)
			handler := crawler.NewObservableHandler(
				tt.observable, tt.mockedExplorer,
				wg, observationInterval, eventChan, errChan, rateLimiter,
			)

			go handler.Start()
			defer handler.Stop()

			event, err := listenToChannels(eventChan, errChan)
			require.NoError(t, err)
			require.NotNil(t, event)
			require.Equal(t, tt.expectedEventType, event.Type())
		})
	}
}

func listenToChannels(
	eventChan chan crawler.Event, errChan chan error,
) (crawler.Event, error) {
	select {
	case err := <-errChan:
		return nil, err
	case event := <-eventChan:
		return event, nil
	}
}

func mockedExplorerForAddressObs() explorer.Service {
	svc := &mockExplorer{}
	svc.On("GetUnspents", mock.Anything, mock.Anything).Return([]explorer.Utxo{}, nil)
	return svc
}

func mockedExplorerForTransactionObs(wantTxConfirmed bool) explorer.Service {
	status := newMockedTxStatus(wantTxConfirmed)

	svc := &mockExplorer{}
	svc.On("GetTransactionStatus", mock.Anything).Return(status, nil)
	return svc
}

func mockedExplorerForOutpointsObs(wantUtxoSpent, wantTxConfirmed bool) explorer.Service {
	utxoStatus := newMockedUtxoStatus(wantUtxoSpent)
	txStatus := newMockedTxStatus(wantTxConfirmed)

	svc := &mockExplorer{}
	svc.On("GetUnspentStatus", mock.Anything, mock.Anything).Return(utxoStatus, nil)
	svc.On("GetTransactionStatus", mock.Anything).Return(txStatus, nil)
	svc.On("GetTransactionHex", mock.Anything).Return(randomHex(100), nil)
	return svc
}

func newTestRateLimiter() *rate.Limiter {
	rt := rate.Every(100 * time.Millisecond)
	return rate.NewLimiter(rt, 1)
}

func newMockedTxStatus(wantTxConfirmed bool) explorer.TransactionStatus {
	confirmed := false
	blockHash := ""
	blockTime, blockHeight := -1, -1
	if wantTxConfirmed {
		confirmed = true
		blockHash = randomHex(32)
		blockTime = 1627421914
		blockHeight = randomIntInRange(100, 1000)
	}
	status := &mockTxStatus{}
	status.On("Confirmed").Return(confirmed)
	status.On("BlockHash").Return(blockHash)
	status.On("BlockTime").Return(blockTime)
	status.On("BlockHeight").Return(blockHeight)
	return status
}

func newMockedUtxoStatus(wantUtxoSpent bool) explorer.UtxoStatus {
	spent := false
	hash := ""
	index := -1
	if wantUtxoSpent {
		spent = true
		hash = randomHex(32)
		index = randomIntInRange(0, 15)
	}
	status := &mockUtxoStatus{}
	status.On("Spent").Return(spent)
	status.On("Hash").Return(hash)
	status.On("Index").Return(index)
	return status
}

func unspentUtxoStatus() explorer.UtxoStatus {
	status := &mockUtxoStatus{}
	status.On("Spent").Return(false)
	return status
}

func spentUtxoStatus() explorer.UtxoStatus {
	status := &mockUtxoStatus{}
	status.On("Spent").Return(true)
	status.On("Hash").Return(randomHex(32))
	status.On("Index").Return(randomIntInRange(0, 15))
	return status
}

func mockedOutpoints(num int) []crawler.Outpoint {
	outs := make([]crawler.Outpoint, 0, num)
	for i := 0; i < num; i++ {
		out := &mockOutpoint{}
		out.On("Hash").Return(randomHex(32))
		out.On("Index").Return(randomIntInRange(0, 15))
		outs = append(outs, out)
	}
	return outs
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(len))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	rand.Read(b)
	return b
}

func randomIntInRange(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(int(n.Int64())) + min
}

func h2b(str string) []byte {
	buf, _ := hex.DecodeString(str)
	return buf
}
