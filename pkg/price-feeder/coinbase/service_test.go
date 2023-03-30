package coinbasefeeder

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"

	"github.com/stretchr/testify/require"
)

var (
	interval = 1000 // 1s interval
	tickers  = []string{"BTC-USD", "BTC-EUR"}
)

func TestService(t *testing.T) {
	feederSvc, err := newTestService()
	require.NoError(t, err)

	go func() {
		err := feederSvc.Start()
		require.NoError(t, err)
	}()

	go func() {
		time.Sleep(5 * time.Second)
		feederSvc.Stop()
	}()

	go func() {
		time.Sleep(2 * time.Second)
		markets := mockedMarkets([]string{"ETH-USD"})
		err := feederSvc.SubscribeMarkets(markets)
		require.NoError(t, err)
	}()

	go func() {
		time.Sleep(3 * time.Second)
		markets := mockedMarkets(tickers)
		err := feederSvc.UnSubscribeMarkets(markets)
		require.NoError(t, err)
	}()

	count := 0
	for priceFeed := range feederSvc.FeedChan() {
		count++
		require.NotEmpty(t, priceFeed.Market.BaseAsset)
		require.NotEmpty(t, priceFeed.Market.QuoteAsset)
		require.NotEmpty(t, priceFeed.Market.Ticker)
		require.NotEmpty(t, priceFeed.Price.BasePrice)
		require.NotEmpty(t, priceFeed.Price.QuotePrice)
	}
	require.Greater(t, count, 0)
}

func newTestService() (pricefeeder.PriceFeeder, error) {
	markets := mockedMarkets(tickers)
	svc, err := NewCoinbasePriceFeeder(interval)
	if err != nil {
		return nil, err
	}
	if err := svc.SubscribeMarkets(markets); err != nil {
		return nil, err
	}
	return svc, nil
}

func mockedMarkets(tickers []string) []pricefeeder.Market {
	markets := make([]pricefeeder.Market, 0, len(tickers))
	for _, ticker := range tickers {
		markets = append(markets, newMockedMarket(ticker))
	}
	return markets
}

func newMockedMarket(ticker string) pricefeeder.Market {
	return pricefeeder.Market{
		BaseAsset:  randomHex(32),
		QuoteAsset: randomHex(32),
		Ticker:     ticker,
	}
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(len))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	rand.Read(b)
	return b
}
