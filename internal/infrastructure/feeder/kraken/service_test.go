package krakenfeeder

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/stretchr/testify/require"
)

var (
	interval = 1000 // 1s interval
	tickers  = []string{"XBT/USDT", "XBT/EUR"}
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

	count := 0
	for priceFeed := range feederSvc.FeedChan() {
		count++
		require.NotNil(t, priceFeed.GetMarket())
		require.NotNil(t, priceFeed.GetPrice())
		require.NotEmpty(t, priceFeed.GetMarket().GetBaseAsset())
		require.NotEmpty(t, priceFeed.GetMarket().GetQuoteAsset())
		require.NotEmpty(t, priceFeed.GetMarket().Ticker())
		require.NotEmpty(t, priceFeed.GetPrice().GetBasePrice())
		require.NotEmpty(t, priceFeed.GetPrice().GetQuotePrice())
	}
	require.Greater(t, count, 0)
}

func newTestService() (ports.PriceFeeder, error) {
	markets := mockedMarkets(tickers)
	svc, err := NewKrakenPriceFeeder(interval)
	if err != nil {
		return nil, err
	}
	if err := svc.SubscribeMarkets(markets); err != nil {
		return nil, err
	}
	return svc, nil
}

func mockedMarkets(tickers []string) []ports.Market {
	markets := make([]ports.Market, 0, len(tickers))
	for _, ticker := range tickers {
		markets = append(markets, newMockedMarket(ticker))
	}
	return markets
}

type mockMarket struct {
	baseAsset  string
	quoteAsset string
	ticker     string
}

func newMockedMarket(ticker string) ports.Market {
	return &mockMarket{
		baseAsset:  randomHex(32),
		quoteAsset: randomHex(32),
		ticker:     ticker,
	}
}

func (m *mockMarket) GetBaseAsset() string {
	return m.baseAsset
}

func (m *mockMarket) GetQuoteAsset() string {
	return m.quoteAsset
}

func (m *mockMarket) Ticker() string {
	return m.ticker
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(len))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	rand.Read(b)
	return b
}
