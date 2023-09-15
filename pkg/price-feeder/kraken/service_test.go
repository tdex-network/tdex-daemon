package krakenfeeder_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"
	krakenfeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/kraken"

	"github.com/stretchr/testify/require"
)

var (
	tickers = []string{"XBT/USDT", "XBT/EUR"}
)

func TestService(t *testing.T) {
	feederSvc, err := krakenfeeder.NewService()
	require.NoError(t, err)
	require.NotNil(t, feederSvc)

	markets := newMarkets(tickers)
	err = feederSvc.SubscribeMarkets(markets)
	require.NoError(t, err)

	feedCh := feederSvc.Start()
	require.NotNil(t, feedCh)

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()

		for feed := range feedCh {
			require.NotEmpty(t, feed.Market)
			require.NotEmpty(t, feed.Price)
			fmt.Printf("received feed %+v\n", feed)
		}
	}()

	go func() {
		defer wg.Done()

		time.Sleep(5 * time.Second)
		feederSvc.Stop()
	}()

	time.Sleep(1 * time.Second)

	mkts := newMarkets([]string{"ETH/USDT"})
	err = feederSvc.SubscribeMarkets(mkts)
	require.NoError(t, err)
	markets = append(markets, mkts...)

	go func(markets []pricefeeder.Market) {
		defer wg.Done()

		time.Sleep(3 * time.Second)
		err := feederSvc.UnsubscribeMarkets(markets)
		require.NoError(t, err)
	}(markets)

	wg.Wait()
}

func newMarkets(tickers []string) []pricefeeder.Market {
	markets := make([]pricefeeder.Market, 0, len(tickers))
	for _, ticker := range tickers {
		markets = append(markets, pricefeeder.Market{
			BaseAsset:  randomHex(32),
			QuoteAsset: randomHex(32),
			Ticker:     ticker,
		})
	}
	return markets
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(len))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	// nolint
	rand.Read(b)
	return b
}
