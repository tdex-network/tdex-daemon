package pricefeederinfra_test

import (
	"context"
	"testing"
	"time"

	coinbasefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/coinbase"
	krakenfeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/kraken"

	"github.com/stretchr/testify/require"
	pricefeederinfra "github.com/tdex-network/tdex-daemon/internal/infrastructure/price-feeder"
	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"
	bitfinexfeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/bitfinex"
)

var (
	ctx = context.Background()

	interval = 1000 // 1s interval
	//bitfinexTickers = []string{"BTCUST", "BTCEUT"}
	//coinbaseTickers = []string{"BTC-USD", "BTC-EUR"}
	//krakenTickers   = []string{"XBT/USDT", "XBT/EUR"}
)

func TestPriceFeedService(t *testing.T) {
	feederSvcBySource, priceFeedStore, err := prepare()
	require.NoError(t, err)

	priceFeedSvc := pricefeederinfra.NewService(feederSvcBySource, priceFeedStore)

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

func prepare() (map[string]pricefeeder.PriceFeeder,
	pricefeederinfra.PriceFeedStore, error) {
	bitfinexSvc, err := bitfinexfeeder.NewService(interval)
	if err != nil {
		return nil, nil, err
	}

	coinbaseSvc, err := coinbasefeeder.NewService(interval)
	if err != nil {
		return nil, nil, err
	}

	krakenSvc, err := krakenfeeder.NewService(interval)
	if err != nil {
		return nil, nil, err
	}

	feederSvcBySource := map[string]pricefeeder.PriceFeeder{
		"bitfinex": bitfinexSvc,
		"coinbase": coinbaseSvc,
		"kraken":   krakenSvc,
	}

	priceFeedStore, err := mockPriceFeedStoreStore()
	if err != nil {
		return nil, nil, err
	}

	return feederSvcBySource, priceFeedStore, nil
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
