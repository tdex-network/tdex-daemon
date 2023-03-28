package pricefeederinfra

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceFeedRepository(t *testing.T) {
	repo, err := NewPriceFeedRepositoryImpl("", nil)
	require.NoError(t, err)

	//happy path
	t.Run("AddAndDeletePriceFeed", testAddAndDeletePriceFeed(repo))
	t.Run("GetPriceFeed", testGetPriceFeed(repo))
	t.Run("GetPriceFeedsByMarket", testGetPriceFeedsByMarket(repo))
	t.Run("UpdatePriceFeed", testUpdatePriceFeed(repo))

	//check errors
	t.Run("GetPriceFeedNotFound", testGetPriceFeedNotFound(repo))
	t.Run("PriceFeedAlreadyExists", testPriceFeedAlreadyExists(repo))
	t.Run("UpdateExistingPriceFeedMarket", testUpdateExistingPriceFeedMarket(repo))
}

func testAddAndDeletePriceFeed(repo PriceFeedRepository) func(*testing.T) {
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
			Source: "kraken",
			On:     false,
		})
		require.NoError(t, err)

		err = repo.RemovePriceFeed(ctx, id)
		require.NoError(t, err)
	}
}

func testGetPriceFeed(repo PriceFeedRepository) func(*testing.T) {
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

func testGetPriceFeedsByMarket(repo PriceFeedRepository) func(*testing.T) {
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

func testUpdatePriceFeed(repo PriceFeedRepository) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		priceFeed := createTestPriceFeed()

		err := repo.AddPriceFeed(ctx, *priceFeed)
		require.NoError(t, err)

		pf, err := repo.GetPriceFeed(ctx, priceFeed.ID)
		require.NoError(t, err)
		assert.Equal(t, false, pf.On)

		err = repo.UpdatePriceFeed(
			ctx,
			priceFeed.ID,
			func(pf *PriceFeed) (*PriceFeed, error) {
				pf.On = true
				return pf, nil
			},
		)
		require.NoError(t, err)

		updatedPriceFeed, err := repo.GetPriceFeed(ctx, priceFeed.ID)
		require.NoError(t, err)
		assert.NotNil(t, updatedPriceFeed)
		assert.Equal(t, priceFeed.ID, updatedPriceFeed.ID)
		assert.Equal(t, true, updatedPriceFeed.On)
	}
}

func testGetPriceFeedNotFound(repo PriceFeedRepository) func(*testing.T) {
	return func(t *testing.T) {
		pf, err := repo.GetPriceFeed(context.Background(), "dummy")
		assert.Nil(t, pf)
		assert.EqualError(t, err, ErrPriceFeedNotFound.Error())
	}
}

func testPriceFeedAlreadyExists(repo PriceFeedRepository) func(*testing.T) {
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
				Source: "kraken",
				On:     false,
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
				Source: "coinbase",
				On:     false,
			},
		)
		require.EqualError(t, err, ErrPriceFeedAlreadyExists.Error())
	}
}

func testUpdateExistingPriceFeedMarket(repo PriceFeedRepository) func(*testing.T) {
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
				Source: "kraken",
				On:     false,
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
		Source: "kraken",
		On:     false,
	}
}

func randAsset() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return string(b)
}
