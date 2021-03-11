package db_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
)

var (
	readOnly = true
	ctx      = context.Background()
)

func TestMarketRepositoryImplementations(t *testing.T) {
	repositories, cancel := createMarketRepositories(t)
	t.Cleanup(cancel)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("testGetOrCreateMarket", func(t *testing.T) {
				t.Parallel()
				testGetOrCreateMarket(t, repo)
			})

			t.Run("testGetMarketByAccount", func(t *testing.T) {
				t.Parallel()
				testGetMarketByAccount(t, repo)
			})

			t.Run("testGetMarketByAsset", func(t *testing.T) {
				t.Parallel()
				testGetMarketByAsset(t, repo)
			})

			t.Run("testGetLatestMarket", func(t *testing.T) {
				t.Parallel()
				testGetLatestMarket(t, repo)
			})

			t.Run("testGetAllMarkets", func(t *testing.T) {
				t.Parallel()
				testGetAllMarkets(t, repo)
			})

			t.Run("testOpenCloseMarket", func(t *testing.T) {
				t.Parallel()
				testOpenCloseMarket(t, repo)
			})

			// TODO: uncomment - the following test demonstrate that in case of error,
			// any change to the db is rolled back. Currently, the transactional
			// component is not properly implemented for inmemory DbManager and the
			// test below would fail for te inmemory implementation.

			// t.Run("testWrite_rollback", func(t *testing.T) {
			// 	t.Parallel()
			// 	testWriteRollback(t, repo)
			// })
		})
	}
}

func testGetOrCreateMarket(t *testing.T, repo marketRepository) {
	// to create a market is mandatory to specify the AccountIndex and Fee
	accountIndex := domain.MarketAccountStart
	fee := int64(25)

	iNewMarket, err := repo.write(
		func(ctx context.Context) (interface{}, error) {
			return repo.Repository.GetOrCreateMarket(
				ctx,
				&domain.Market{AccountIndex: accountIndex, Fee: fee},
			)
		},
	)
	require.NoError(t, err)

	newMarket, ok := iNewMarket.(*domain.Market)
	require.True(t, ok)
	require.NotNil(t, newMarket)
	require.Equal(t, accountIndex, newMarket.AccountIndex)
	require.Equal(t, fee, newMarket.Fee)

	// to retrieve an existing market is enough to specify just the AccountIndex
	iExistingMarket, err := repo.read(
		func(ctx context.Context) (interface{}, error) {
			return repo.Repository.GetOrCreateMarket(
				ctx,
				&domain.Market{AccountIndex: accountIndex},
			)
		},
	)
	require.NoError(t, err)

	existingMarket, ok := iExistingMarket.(*domain.Market)
	require.True(t, ok)
	require.NotNil(t, existingMarket)
	require.Exactly(t, newMarket, existingMarket)
}

func testGetMarketByAccount(t *testing.T, repo marketRepository) {
	accountIndex := domain.MarketAccountStart + 1

	market, err := repo.read(
		func(ctx context.Context) (interface{}, error) {
			return repo.Repository.GetMarketByAccount(ctx, accountIndex)
		},
	)
	require.NoError(t, err)
	require.Nil(t, market)

	_, err = repo.write(
		func(ctx context.Context) (interface{}, error) {
			return repo.Repository.GetOrCreateMarket(
				ctx,
				&domain.Market{AccountIndex: accountIndex, Fee: 25},
			)
		},
	)
	require.NoError(t, err)

	market, err = repo.read(
		func(ctx context.Context) (interface{}, error) {
			return repo.Repository.GetMarketByAccount(ctx, accountIndex)
		},
	)
	require.NoError(t, err)
	require.NotNil(t, market)
}

func testGetMarketByAsset(t *testing.T, repo marketRepository) {
	accountIndex := domain.MarketAccountStart + 2
	baseAsset := "0000000000000000000000000000000000000000000000000000000000000000"
	quoteAsset := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	mockedOutpoints := mockMarketFunds(baseAsset, quoteAsset)

	type response struct {
		market       *domain.Market
		accountIndex int
	}
	resp, err := repo.read(
		func(ctx context.Context) (interface{}, error) {
			mkt, accountIndex, err := repo.Repository.GetMarketByAsset(ctx, quoteAsset)
			if err != nil {
				return nil, err
			}
			return response{mkt, accountIndex}, nil
		},
	)
	require.NoError(t, err)
	require.Nil(t, resp.(response).market)
	require.Equal(t, -1, resp.(response).accountIndex)

	_, err = repo.write(
		func(ctx context.Context) (interface{}, error) {
			_, err := repo.Repository.GetOrCreateMarket(
				ctx,
				&domain.Market{AccountIndex: accountIndex, Fee: 25},
			)
			if err != nil {
				return nil, err
			}

			return nil, repo.Repository.UpdateMarket(
				ctx,
				accountIndex,
				func(market *domain.Market) (*domain.Market, error) {
					err := market.FundMarket(
						mockedOutpoints,
						baseAsset,
					)
					return market, err
				},
			)
		},
	)
	require.NoError(t, err)

	resp, err = repo.read(
		func(ctx context.Context) (interface{}, error) {
			mkt, accountIndex, err := repo.Repository.GetMarketByAsset(ctx, quoteAsset)
			if err != nil {
				return nil, err
			}
			return response{mkt, accountIndex}, nil
		},
	)
	require.NoError(t, err)
	require.NotNil(t, resp.(response).market)
	require.Equal(t, accountIndex, resp.(response).accountIndex)
}

func testGetLatestMarket(t *testing.T, repo marketRepository) {
	resp, err := repo.read(func(ctx context.Context) (interface{}, error) {
		market, accountIndex, err := repo.Repository.GetLatestMarket(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"market":       market,
			"accountIndex": accountIndex,
		}, nil
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func testGetAllMarkets(t *testing.T, repo marketRepository) {
	_, err := repo.read(func(ctx context.Context) (interface{}, error) {
		return repo.Repository.GetAllMarkets(ctx)
	})
	require.NoError(t, err)
}

func testOpenCloseMarket(t *testing.T, repo marketRepository) {
	accountIndex := domain.MarketAccountStart + 4
	baseAsset := "0000000000000000000000000000000000000000000000000000000000000000"
	quoteAsset := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	mockedOutpoints := mockMarketFunds(baseAsset, quoteAsset)
	iOpenMarkets, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.GetOrCreateMarket(
			ctx,
			&domain.Market{AccountIndex: accountIndex, Fee: 25},
		); err != nil {
			return nil, err
		}
		if err := repo.Repository.UpdateMarket(
			ctx,
			accountIndex,
			func(m *domain.Market) (*domain.Market, error) {
				if err := m.FundMarket(mockedOutpoints, baseAsset); err != nil {
					return nil, err
				}
				return m, nil
			},
		); err != nil {
			return nil, err
		}

		if err := repo.Repository.OpenMarket(ctx, quoteAsset); err != nil {
			return nil, err
		}

		return repo.Repository.GetTradableMarkets(ctx)
	})
	require.NoError(t, err)
	require.NotNil(t, iOpenMarkets)
	markets, ok := iOpenMarkets.([]domain.Market)
	require.True(t, ok)
	require.True(t, len(markets) > 0)

	iClosedMarkets, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if err := repo.Repository.CloseMarket(ctx, quoteAsset); err != nil {
			return nil, err
		}
		return repo.Repository.GetTradableMarkets(ctx)
	})
	require.NoError(t, err)
	markets, ok = iClosedMarkets.([]domain.Market)
	require.True(t, ok)
	require.Len(t, markets, 0)
}

func testUpdatePrices(t *testing.T, repo marketRepository) {
	accountIndex := domain.MarketAccountStart + 5
	iNewMarket, err := repo.write(func(ctx context.Context) (interface{}, error) {
		return repo.Repository.GetOrCreateMarket(
			ctx,
			&domain.Market{AccountIndex: accountIndex, Fee: 25},
		)
	})
	require.NoError(t, err)
	require.NotNil(t, iNewMarket)
	market, ok := iNewMarket.(*domain.Market)
	require.True(t, ok)
	require.True(t, market.Price.AreZero())

	_, err = repo.writePrice(func(ctx context.Context) (interface{}, error) {
		return nil, repo.Repository.UpdatePrices(
			ctx,
			accountIndex,
			domain.Prices{
				BasePrice:  decimal.NewFromFloat(0.00002),
				QuotePrice: decimal.NewFromInt(50000),
			},
		)
	})
	require.NoError(t, err)

	iUpdatedMarket, err := repo.read(func(ctx context.Context) (interface{}, error) {
		return repo.Repository.GetOrCreateMarket(
			ctx,
			&domain.Market{AccountIndex: accountIndex},
		)
	})
	require.NoError(t, err)
	require.NotNil(t, iUpdatedMarket)
	market, ok = iUpdatedMarket.(*domain.Market)
	require.True(t, ok)
	require.False(t, market.Price.AreZero())
}

func testWriteRollback(t *testing.T, repo marketRepository) {
	accountIndex := domain.MarketAccountStart + 4
	mockedError := errors.New("somehting went wrong")

	market, err := repo.read(func(ctx context.Context) (interface{}, error) {
		return repo.Repository.GetMarketByAccount(ctx, accountIndex)
	})
	require.NoError(t, err)
	require.Nil(t, market)

	_, err = repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.GetOrCreateMarket(
			ctx,
			&domain.Market{AccountIndex: accountIndex, Fee: 25},
		); err != nil {
			return nil, err
		}

		if err := repo.Repository.UpdateMarket(
			ctx,
			accountIndex,
			func(_ *domain.Market) (*domain.Market, error) {
				return nil, mockedError
			},
		); err != nil {
			return nil, err
		}

		return market, nil
	})
	require.EqualError(t, err, mockedError.Error())

	market, err = repo.read(func(ctx context.Context) (interface{}, error) {
		return repo.Repository.GetMarketByAccount(ctx, accountIndex)
	})
	require.NoError(t, err)
	require.Nil(t, market)
}

func createMarketRepositories(t *testing.T) ([]marketRepository, func()) {
	datadir := "marketdb"
	err := os.Mkdir(datadir, os.ModePerm)
	require.NoError(t, err)

	inmemoryDBManager := inmemory.NewDbManager()
	badgerDBManager, err := dbbadger.NewDbManager(datadir, nil)
	require.NoError(t, err)

	return []marketRepository{
			{
				Name:       "badger",
				DBManager:  badgerDBManager,
				Repository: newBadgerMarketRepository(badgerDBManager),
			},
			{
				Name:       "inmemory",
				DBManager:  inmemoryDBManager,
				Repository: newInMemoryMarketRepository(inmemoryDBManager),
			},
		}, func() {
			os.RemoveAll(datadir)
		}
}

func newBadgerMarketRepository(dbmanager *dbbadger.DbManager) domain.MarketRepository {
	return dbbadger.NewMarketRepositoryImpl(dbmanager)
}

func newInMemoryMarketRepository(dbmanager *inmemory.DbManager) domain.MarketRepository {
	return inmemory.NewMarketRepositoryImpl(dbmanager)
}

type marketRepository struct {
	Name       string
	DBManager  ports.DbManager
	Repository domain.MarketRepository
}

func (r marketRepository) read(query func(context.Context) (interface{}, error)) (interface{}, error) {
	return r.DBManager.RunTransaction(context.Background(), true, query)
}

func (r marketRepository) write(query func(context.Context) (interface{}, error)) (interface{}, error) {
	return r.DBManager.RunTransaction(context.Background(), false, query)
}

func (r marketRepository) writePrice(query func(context.Context) (interface{}, error)) (interface{}, error) {
	return r.DBManager.RunPricesTransaction(context.Background(), false, query)
}

func mockMarketFunds(baseAsset, quoteAsset string) []domain.OutpointWithAsset {
	return []domain.OutpointWithAsset{
		{
			Asset: baseAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  0,
		},
		{
			Asset: quoteAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  1,
		},
	}
}
