package pricefeederinfra

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceFeedStore(t *testing.T) {
	repo, err := NewPriceFeedStoreImpl("", nil)
	require.NoError(t, err)

	//happy path
	t.Run("AddAndDeletePriceFeed", testAddAndDeletePriceFeed(repo))
	t.Run("GetPriceFeed", testGetPriceFeed(repo))
	t.Run("GetPriceFeedsByMarket", testGetPriceFeedsByMarket(repo))
	t.Run("UpdatePriceFeed", testUpdatePriceFeed(repo))
	t.Run("GetAll", testGetAll(repo))
	t.Run("GetStartedPriceFeeds", testGetStartedPriceFeeds(repo))

	//check errors
	t.Run("GetPriceFeedNotFound", testGetPriceFeedNotFound(repo))
	t.Run("PriceFeedAlreadyExists", testPriceFeedAlreadyExists(repo))
	t.Run("UpdateExistingPriceFeedMarket", testUpdateExistingPriceFeedMarket(repo))
}

func testAddAndDeletePriceFeed(repo PriceFeedStore) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		id := uuid.New().String()
		err := repo.AddPriceFeed(ctx, PriceFeed{
			ID: id,
			Market: Market{
				BaseAsset:  "BA2",
				QuoteAsset: "QA2",
				Ticker:     "XBT/USD",
			},
			Source:  "kraken",
			Started: false,
		})
		require.NoError(t, err)

		err = repo.RemovePriceFeed(ctx, id)
		require.NoError(t, err)
	}
}

func testGetPriceFeed(repo PriceFeedStore) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		priceFeed := createTestPriceFeed()

		err := repo.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		fetchedPriceFeed, err := repo.GetPriceFeed(ctx, priceFeed.ID)
		require.NoError(t, err)
		assert.NotNil(t, fetchedPriceFeed)
		assert.Equal(t, priceFeed.ID, fetchedPriceFeed.ID)
	}
}

func testGetPriceFeedsByMarket(repo PriceFeedStore) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		priceFeed := createTestPriceFeed()

		err := repo.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		market := Market{
			BaseAsset:  priceFeed.Market.BaseAsset,
			QuoteAsset: priceFeed.Market.QuoteAsset,
		}

		fetchedPriceFeed, err := repo.GetPriceFeedsByMarket(ctx, market)
		require.NoError(t, err)
		assert.NotNil(t, fetchedPriceFeed)
		assert.Equal(t, priceFeed.ID, fetchedPriceFeed.ID)
	}
}

func testUpdatePriceFeed(repo PriceFeedStore) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		priceFeed := createTestPriceFeed()

		err := repo.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		pf, err := repo.GetPriceFeed(ctx, priceFeed.ID)
		require.NoError(t, err)
		assert.Equal(t, false, pf.Started)

		err = repo.UpdatePriceFeed(
			ctx,
			priceFeed.ID,
			func(pf *PriceFeed) (*PriceFeed, error) {
				pf.Started = true
				return pf, nil
			},
		)
		require.NoError(t, err)

		updatedPriceFeed, err := repo.GetPriceFeed(ctx, priceFeed.ID)
		require.NoError(t, err)
		assert.NotNil(t, updatedPriceFeed)
		assert.Equal(t, priceFeed.ID, updatedPriceFeed.ID)
		assert.Equal(t, true, updatedPriceFeed.Started)
	}
}

func testGetAll(repo PriceFeedStore) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		priceFeed := createTestPriceFeed()

		err := repo.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		priceFeed = createTestPriceFeed()
		err = repo.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		priceFeeds, err := repo.GetAllPriceFeeds(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, priceFeeds)
	}
}

func testGetStartedPriceFeeds(repo PriceFeedStore) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		priceFeeds, err := repo.GetAllPriceFeeds(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, priceFeeds)

		priceFeeds, err = repo.GetStartedPriceFeeds(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, len(priceFeeds))
	}
}

func testGetPriceFeedNotFound(repo PriceFeedStore) func(*testing.T) {
	return func(t *testing.T) {
		pf, err := repo.GetPriceFeed(context.Background(), "dummy")
		assert.Nil(t, pf)
		assert.EqualError(t, err, ErrPriceFeedNotFound.Error())
	}
}

func testPriceFeedAlreadyExists(repo PriceFeedStore) func(*testing.T) {
	return func(t *testing.T) {
		err := repo.AddPriceFeed(
			context.Background(),
			PriceFeed{
				ID: uuid.New().String(),
				Market: Market{
					BaseAsset:  "BA1",
					QuoteAsset: "QA1",
					Ticker:     "XBT/USD",
				},
				Source:  "kraken",
				Started: false,
			},
		)
		require.NoError(t, err)

		err = repo.AddPriceFeed(
			context.Background(),
			PriceFeed{
				ID: uuid.New().String(),
				Market: Market{
					BaseAsset:  "BA1",
					QuoteAsset: "QA1",
					Ticker:     "XBT/EUR",
				},
				Source:  "coinbase",
				Started: false,
			},
		)
		require.EqualError(t, err, ErrPriceFeedAlreadyExists.Error())
	}
}

func testUpdateExistingPriceFeedMarket(repo PriceFeedStore) func(*testing.T) {
	return func(t *testing.T) {
		id := uuid.New().String()
		err := repo.AddPriceFeed(
			context.Background(),
			PriceFeed{
				ID: id,
				Market: Market{
					BaseAsset:  "BA",
					QuoteAsset: "QA",
					Ticker:     "XBT/USD",
				},
				Source:  "kraken",
				Started: false,
			},
		)
		require.NoError(t, err)

		err = repo.UpdatePriceFeed(
			context.Background(),
			id,
			func(pf *PriceFeed) (*PriceFeed, error) {
				pf.Market.BaseAsset = "dummy"
				return pf, nil
			},
		)
		require.EqualError(t, err, ErrPriceFeedMarketCannotBeChanged.Error())

		err = repo.UpdatePriceFeed(
			context.Background(),
			id,
			func(pf *PriceFeed) (*PriceFeed, error) {
				pf.Market.QuoteAsset = "dummy"
				return pf, nil
			},
		)
		require.EqualError(t, err, ErrPriceFeedMarketCannotBeChanged.Error())
	}
}

func createTestPriceFeed() *PriceFeed {
	return &PriceFeed{
		ID: uuid.New().String(),
		Market: Market{
			BaseAsset:  randAsset(),
			QuoteAsset: randAsset(),
			Ticker:     "XBT/USD",
		},
		Source:  "kraken",
		Started: false,
	}
}

func randAsset() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return string(b)
}
