package pricefeederstore_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	pricefeeder "github.com/tdex-network/tdex-daemon/internal/infrastructure/price-feeder"
	pricefeederstore "github.com/tdex-network/tdex-daemon/internal/infrastructure/price-feeder/store/badger"
)

func TestPriceFeedStore(t *testing.T) {
	t.Run("AddAndDeletePriceFeed", testAddAndDeletePriceFeed())
	t.Run("GetPriceFeed", testGetPriceFeed())
	t.Run("UpdatePriceFeed", testUpdatePriceFeed())
	t.Run("GetAll", testGetAll())
}

func testAddAndDeletePriceFeed() func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		store, err := pricefeederstore.NewPriceFeedStore("", nil)
		require.NoError(t, err)

		priceFeed := createTestPriceFeed()
		err = store.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		err = store.AddPriceFeed(ctx, *priceFeed)
		require.Error(t, err)

		err = store.RemovePriceFeed(ctx, priceFeed.GetId())
		require.NoError(t, err)

		err = store.RemovePriceFeed(ctx, priceFeed.GetId())
		require.NoError(t, err)
	}
}

func testGetPriceFeed() func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		priceFeed := createTestPriceFeed()

		store, err := pricefeederstore.NewPriceFeedStore("", nil)
		require.NoError(t, err)

		err = store.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		fetchedPriceFeed, err := store.GetPriceFeed(ctx, priceFeed.ID)
		require.NoError(t, err)
		require.NotNil(t, fetchedPriceFeed)
		require.Equal(t, priceFeed.ID, fetchedPriceFeed.ID)
	}
}

func testUpdatePriceFeed() func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		priceFeed := createTestPriceFeed()

		store, err := pricefeederstore.NewPriceFeedStore("", nil)
		require.NoError(t, err)

		err = store.UpdatePriceFeed(ctx, priceFeed.ID, nil)
		require.Error(t, err)

		err = store.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		err = store.UpdatePriceFeed(
			ctx, priceFeed.ID, func(
				pf *pricefeeder.PriceFeedInfo,
			) (*pricefeeder.PriceFeedInfo, error) {
				pf.Started = true
				return pf, nil
			},
		)
		require.NoError(t, err)

		err = store.UpdatePriceFeed(
			ctx, priceFeed.ID, func(
				pf *pricefeeder.PriceFeedInfo,
			) (*pricefeeder.PriceFeedInfo, error) {
				return nil, fmt.Errorf("test error")
			},
		)
		require.Error(t, err)

		gotPriceFeed, err := store.GetPriceFeed(ctx, priceFeed.ID)
		require.NoError(t, err)
		require.NotNil(t, gotPriceFeed)
		require.Equal(t, priceFeed.ID, gotPriceFeed.ID)
		require.True(t, gotPriceFeed.Started)
	}
}

func testGetAll() func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		priceFeed := createTestPriceFeed()

		store, err := pricefeederstore.NewPriceFeedStore("", nil)
		require.NoError(t, err)

		err = store.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		priceFeed = createTestPriceFeed()
		err = store.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		priceFeeds, err := store.GetAllPriceFeeds(ctx)
		require.NoError(t, err)
		require.Len(t, priceFeeds, 2)
	}
}

func createTestPriceFeed() *pricefeeder.PriceFeedInfo {
	return &pricefeeder.PriceFeedInfo{
		ID: uuid.New().String(),
		Market: pricefeeder.Market{
			BaseAsset:  randAsset(),
			QuoteAsset: randAsset(),
		},
		Ticker:  "XBT/USD",
		Source:  "kraken",
		Started: false,
	}
}

func randAsset() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return string(b)
}
