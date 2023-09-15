package pricefeeder_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	pricefeeder "github.com/tdex-network/tdex-daemon/internal/infrastructure/price-feeder"
	pricefeederstore "github.com/tdex-network/tdex-daemon/internal/infrastructure/price-feeder/store/badger"
)

var (
	ctx = context.Background()

	tickerBySource = map[string]string{
		"bitfinex": "BTCUST",
		"coinbase": "BTC-USD",
		"kraken":   "XBT/USDT",
	}
)

func TestPriceFeedService(t *testing.T) {
	store, err := pricefeederstore.NewPriceFeedStore("", nil)
	require.NoError(t, err)
	require.NotNil(t, store)

	priceFeedSvc := pricefeeder.NewService(store)

	wg := &sync.WaitGroup{}
	wg.Add(6)

	go func() {
		defer wg.Done()

		time.Sleep(10 * time.Second)
		priceFeedSvc.Close()
	}()

	feeds := make([]string, 0)
	for source, ticker := range tickerBySource {
		market := randomMarket()
		id, err := priceFeedSvc.AddPriceFeed(ctx, market, source, ticker)
		require.NoError(t, err)
		require.NotEmpty(t, id)
		feeds = append(feeds, id)
	}

	for _, id := range feeds {
		feedCh, err := priceFeedSvc.StartPriceFeed(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, feedCh)

		go func(ch chan ports.PriceFeed) {
			defer wg.Done()

			for feed := range ch {
				require.NotNil(t, feed)
				require.NotNil(t, feed.GetMarket())
				require.NotNil(t, feed.GetPrice())
				fmt.Printf("received feed %+v\n", feed)
			}
		}(feedCh)
	}

	time.Sleep(3 * time.Second)
	market := randomMarket()
	id, err := priceFeedSvc.AddPriceFeed(ctx, market, "kraken", "XBT/USDT")
	require.NoError(t, err)
	require.NotEmpty(t, id)

	feeds = append(feeds, id)

	ch, err := priceFeedSvc.StartPriceFeed(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, ch)

	go func() {
		defer wg.Done()
		for feed := range ch {
			require.NotNil(t, feed)
			require.NotNil(t, feed.GetMarket())
			require.NotNil(t, feed.GetPrice())
			fmt.Printf("received feed %+v\n", feed)
		}
	}()

	go func() {
		defer wg.Done()

		time.Sleep(3 * time.Second)
		for _, id := range feeds {
			err := priceFeedSvc.StopPriceFeed(ctx, id)
			require.NoError(t, err)
		}
	}()

	wg.Wait()
}

func randomMarket() ports.Market {
	return pricefeeder.Market{
		BaseAsset:  randomAsset(),
		QuoteAsset: randomAsset(),
	}
}

func randomAsset() string {
	b := make([]byte, 32)
	// nolint
	rand.Read(b)
	return hex.EncodeToString(b)
}
