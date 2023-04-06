package pricefeederinfra_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	pricefeederinfra "github.com/tdex-network/tdex-daemon/internal/infrastructure/price-feeder"
)

var (
	ctx = context.Background()

	//bitfinexTickers = []string{"BTCUST", "BTCEUT"}
	//coinbaseTickers = []string{"BTC-USD", "BTC-EUR"}
	//krakenTickers   = []string{"XBT/USDT", "XBT/EUR"}
)

func TestPriceFeedService(t *testing.T) {
	priceFeedStore, err := prepare()
	require.NoError(t, err)

	priceFeedSvc := pricefeederinfra.NewService(priceFeedStore)

	priceFeedChan, err := priceFeedSvc.Start(ctx)
	require.NoError(t, err)

	go func() {
		time.Sleep(7 * time.Second)
		priceFeedSvc.Stop(ctx)
	}()

	go func() {
		time.Sleep(3 * time.Second)
		err := priceFeedSvc.StopFeed(ctx, "1")
		require.NoError(t, err)
		err = priceFeedSvc.UpdatePriceFeed(ctx, "1", "kraken", "XBT/EUR")
		require.NoError(t, err)
		err = priceFeedSvc.StartFeed(ctx, "1")
		require.NoError(t, err)
	}()

	go func() {
		time.Sleep(5 * time.Second)
		err := priceFeedSvc.StopFeed(ctx, "2")
		require.NoError(t, err)
		err = priceFeedSvc.StopFeed(ctx, "3")
	}()

	count := 0
	for pf := range priceFeedChan {
		count++
		require.NotEmpty(t, pf.GetMarket().GetBaseAsset())
		require.NotEmpty(t, pf.GetMarket().GetQuoteAsset())
		require.NotEmpty(t, pf.GetPrice().GetBasePrice())
		require.NotEmpty(t, pf.GetPrice().GetQuotePrice())

		t.Logf("market %s-%s, price %v-%v,", pf.GetMarket().GetBaseAsset(),
			pf.GetMarket().GetQuoteAsset(), pf.GetPrice().GetBasePrice(),
			pf.GetPrice().GetQuotePrice())
	}
	require.Greater(t, count, 0)

}

func prepare() (pricefeederinfra.PriceFeedStore, error) {
	return mockPriceFeedStoreStore()
}

type mkt struct {
	BaseAsset  string
	QuoteAsset string
}

func (m mkt) GetBaseAsset() string {
	return m.BaseAsset
}

func (m mkt) GetQuoteAsset() string {
	return m.QuoteAsset
}

func mockPriceFeedStoreStore() (pricefeederinfra.PriceFeedStore, error) {
	store, err := pricefeederinfra.NewPriceFeedStoreImpl("", nil)
	if err != nil {
		return nil, err
	}

	if err := store.AddPriceFeed(
		context.Background(),
		pricefeederinfra.PriceFeed{
			ID: "1",
			Market: pricefeederinfra.Market{
				BaseAsset:  "BA",
				QuoteAsset: "QA",
				Ticker:     "BTCUST",
			},
			Source:  "bitfinex",
			Started: true,
		},
	); err != nil {
		return nil, err
	}

	if err := store.AddPriceFeed(
		context.Background(),
		pricefeederinfra.PriceFeed{
			ID: "2",
			Market: pricefeederinfra.Market{
				BaseAsset:  "BA1",
				QuoteAsset: "QA1",
				Ticker:     "BTC-EUR",
			},
			Source:  "coinbase",
			Started: true,
		},
	); err != nil {
		return nil, err
	}

	if err := store.AddPriceFeed(
		context.Background(),
		pricefeederinfra.PriceFeed{
			ID: "3",
			Market: pricefeederinfra.Market{
				BaseAsset:  "BA2",
				QuoteAsset: "QA2",
				Ticker:     "XBT/CAD",
			},
			Source:  "kraken",
			Started: true,
		},
	); err != nil {
		return nil, err
	}

	return store, nil
}
