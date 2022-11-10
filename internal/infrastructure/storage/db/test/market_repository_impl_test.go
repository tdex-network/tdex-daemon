package db_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
)

func TestMarketRepositoryImplementations(t *testing.T) {
	repositories := createMarketRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("add_and_get_market", func(t *testing.T) {
				testAddAndGetMarket(t, repo.Repository)
			})

			t.Run("update_market", func(t *testing.T) {
				testUpdateMarket(t, repo.Repository)
			})

			t.Run("update_market_price", func(t *testing.T) {
				testUpdateMarketPrice(t, repo.Repository)
			})

			t.Run("open_and_close_market", func(t *testing.T) {
				testOpenCloseMarket(t, repo.Repository)
			})

			t.Run("delete_market", func(t *testing.T) {
				testDeleteMarket(t, repo.Repository)
			})
		})
	}
}

func testAddAndGetMarket(t *testing.T, repo domain.MarketRepository) {
	ctx := context.Background()
	market := makeRandomMarket()

	allMarkets, err := repo.GetAllMarkets(ctx)
	require.NoError(t, err)
	require.Empty(t, allMarkets)

	err = repo.AddMarket(ctx, market)
	require.NoError(t, err)

	allMarkets, err = repo.GetAllMarkets(ctx)
	require.NoError(t, err)
	require.Len(t, allMarkets, 1)

	err = repo.AddMarket(ctx, market)
	require.Error(t, err)

	foundMarket, err := repo.GetMarketByName(ctx, market.Name)
	require.NoError(t, err)
	require.Exactly(t, *market, *foundMarket)

	foundMarket, err = repo.GetMarketByName(ctx, randomHex(32))
	require.Error(t, err)
	require.Nil(t, foundMarket)

	foundMarket, err = repo.GetMarketByAssets(
		ctx, market.BaseAsset, market.QuoteAsset,
	)
	require.NoError(t, err)
	require.Exactly(t, *market, *foundMarket)

	foundMarket, err = repo.GetMarketByAssets(
		ctx, randomHex(32), market.QuoteAsset,
	)
	require.Error(t, err)
	require.Nil(t, foundMarket)

	foundMarket, err = repo.GetMarketByAssets(
		ctx, market.BaseAsset, randomHex(32),
	)
	require.Error(t, err)
	require.Nil(t, foundMarket)

	foundMarket, err = repo.GetMarketByAssets(
		ctx, randomHex(32), randomHex(32),
	)
	require.Error(t, err)
	require.Nil(t, foundMarket)
}

func testUpdateMarket(t *testing.T, repo domain.MarketRepository) {
	ctx := context.Background()

	markets, err := repo.GetAllMarkets(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, markets)

	market := markets[0]

	err = repo.UpdateMarket(
		ctx, market.Name, func(_ *domain.Market) (*domain.Market, error) {
			return nil, fmt.Errorf("something went wrong")
		},
	)
	require.Error(t, err)
	require.EqualError(t, err, "something went wrong")

	err = repo.UpdateMarket(
		ctx, market.Name, func(mkt *domain.Market) (*domain.Market, error) {
			if err := mkt.ChangePercentageFee(100); err != nil {
				return nil, err
			}
			return mkt, nil
		},
	)
	require.NoError(t, err)

	foundMarket, err := repo.GetMarketByName(ctx, market.Name)
	require.NoError(t, err)
	require.NotNil(t, foundMarket)
	require.Equal(t, 100, int(foundMarket.PercentageFee))
}

func testUpdateMarketPrice(t *testing.T, repo domain.MarketRepository) {
	ctx := context.Background()

	markets, err := repo.GetAllMarkets(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, markets)

	market := markets[0]
	require.True(t, market.Price.IsZero())

	err = repo.UpdateMarketPrice(
		ctx, market.Name, domain.MarketPrice{
			BasePrice:  decimal.NewFromFloat(0.00002).String(),
			QuotePrice: decimal.NewFromInt(50000).String(),
		},
	)
	require.NoError(t, err)

	foundMarket, err := repo.GetMarketByName(ctx, market.Name)
	require.NoError(t, err)
	require.NotNil(t, market)
	require.False(t, foundMarket.Price.IsZero())
}

func testOpenCloseMarket(t *testing.T, repo domain.MarketRepository) {
	ctx := context.Background()

	markets, err := repo.GetAllMarkets(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, markets)
	market := markets[0]
	require.False(t, market.Tradable)

	tradableMarkets, err := repo.GetTradableMarkets(ctx)
	require.NoError(t, err)
	require.Empty(t, tradableMarkets)

	err = repo.OpenMarket(ctx, randomHex(20))
	require.Error(t, err)
	err = repo.OpenMarket(ctx, market.Name)
	require.NoError(t, err)
	err = repo.OpenMarket(ctx, market.Name)
	require.NoError(t, err)

	tradableMarkets, err = repo.GetTradableMarkets(ctx)
	require.NoError(t, err)
	require.Len(t, tradableMarkets, 1)
	require.Equal(t, market.BaseAsset, tradableMarkets[0].BaseAsset)
	require.Equal(t, market.QuoteAsset, tradableMarkets[0].QuoteAsset)
	require.True(t, tradableMarkets[0].Tradable)

	err = repo.CloseMarket(ctx, randomHex(20))
	require.Error(t, err)
	err = repo.CloseMarket(ctx, market.Name)
	require.NoError(t, err)
	err = repo.CloseMarket(ctx, market.Name)
	require.NoError(t, err)

	tradableMarkets, err = repo.GetTradableMarkets(ctx)
	require.NoError(t, err)
	require.Empty(t, tradableMarkets)
}

func testDeleteMarket(t *testing.T, repo domain.MarketRepository) {
	ctx := context.Background()

	markets, err := repo.GetAllMarkets(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, markets)

	market := markets[0]

	err = repo.DeleteMarket(ctx, market.Name)
	require.NoError(t, err)

	err = repo.DeleteMarket(ctx, market.Name)
	require.Error(t, err)
}

func createMarketRepositories(t *testing.T) []marketRepository {
	inmemoryDBManager := inmemory.NewRepoManager()
	badgerDBManager, err := dbbadger.NewRepoManager("", nil)
	require.NoError(t, err)

	return []marketRepository{
		{
			Name:       "badger",
			Repository: badgerDBManager.MarketRepository(),
		},
		{
			Name:       "inmemory",
			Repository: inmemoryDBManager.MarketRepository(),
		},
	}
}

type marketRepository struct {
	Name       string
	Repository domain.MarketRepository
}
