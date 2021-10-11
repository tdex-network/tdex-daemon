package crawler_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

func TestCrawler(t *testing.T) {
	explorerSvc := &mockExplorer{}
	explorerSvc.On("GetUnspents", mock.Anything, mock.Anything).Return([]explorer.Utxo{}, nil)
	explorerSvc.On("GetTransactionStatus", mock.Anything).Return(newMockedTxStatus(wantTxConfirmed), nil)
	explorerSvc.On("GetTransactionHex", mock.Anything).Return(randomHex(100), nil)

	crawlSvc := crawler.NewService(crawler.Opts{
		ExplorerSvc: explorerSvc,
		ErrorHandler: func(err error) {
			if err != nil {
				fmt.Println(err)
			}
		},
		CrawlerInterval: 500,
	})

	go crawlSvc.Start()
	go addObservableAfterTimeout(crawlSvc)
	go func() {
		removeObservableAfterTimeout(crawlSvc)
		crawlSvc.Stop()
	}()

	addrEvents, txEvents := listen(t, crawlSvc)
	require.Greater(t, len(addrEvents), 0)
	require.Greater(t, len(txEvents), 0)
}

func listen(
	t *testing.T, crawlSvc crawler.Service,
) (addrEvents, txEvents []crawler.Event) {
	eventChan := crawlSvc.GetEventChannel()
	addrEvents = make([]crawler.Event, 0)
	txEvents = make([]crawler.Event, 0)
	for {
		event := <-eventChan
		switch tt := event.Type(); tt {
		case crawler.CloseSignal:
			return addrEvents, txEvents
		case crawler.AddressUnspents:
			addrEvents = append(addrEvents, event)
		case crawler.TransactionConfirmed, crawler.TransactionUnconfirmed:
			txEvents = append(txEvents, event)
		}
	}
}

func removeObservableAfterTimeout(crawlerSvc crawler.Service) {
	time.Sleep(3 * time.Second)
	crawlerSvc.RemoveObservable(&crawler.AddressObservable{
		Address: "101",
	})
	crawlerSvc.RemoveObservable(crawler.NewTransactionObservable("102", nil))
}

func addObservableAfterTimeout(crawlerSvc crawler.Service) {
	time.Sleep(time.Second)
	crawlerSvc.AddObservable(crawler.NewAddressObservable("101", nil, 0))
	crawlerSvc.AddObservable(crawler.NewTransactionObservable("102", nil))
}
