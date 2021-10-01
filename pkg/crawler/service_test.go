package crawler_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
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
	go listen(t, crawlSvc)

	addObservableAfterTimeout(crawlSvc)
	time.Sleep(2 * time.Second)
	removeObservableAfterTimeout(crawlSvc)
	crawlSvc.Stop()
}

func listen(t *testing.T, crawlSvc crawler.Service) {
	eventChan := crawlSvc.GetEventChannel()
	for {
		event := <-eventChan
		switch tt := event.Type(); tt {
		case crawler.CloseSignal:
			return
		default:
			t.Logf("received event %s\n", tt)
		}
	}
}

func removeObservableAfterTimeout(crawlerSvc crawler.Service) {
	crawlerSvc.RemoveObservable(&crawler.AddressObservable{
		Address: "101",
	})
	time.Sleep(400 * time.Millisecond)
	crawlerSvc.RemoveObservable(crawler.NewTransactionObservable("102", nil))
}

func addObservableAfterTimeout(crawlerSvc crawler.Service) {
	time.Sleep(2 * time.Second)
	crawlerSvc.AddObservable(crawler.NewAddressObservable("101", nil, 0))
	time.Sleep(300 * time.Millisecond)
	crawlerSvc.AddObservable(crawler.NewTransactionObservable("102", nil))
}
