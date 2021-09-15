package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
)

func TestTradeRepositoryImplementations(t *testing.T) {
	repositories := createTradeRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("testGetOrCreateTrade", func(t *testing.T) {
				t.Parallel()
				testGetOrCreateTrade(t, repo)
			})

			t.Run("testGetAllTrades", func(t *testing.T) {
				t.Parallel()
				testGetAllTrades(t, repo)
			})

			t.Run("testGetAllTradesForMarket", func(t *testing.T) {
				t.Parallel()
				testGetAllTradesByMarket(t, repo)
			})

			t.Run("testGetCompletedTradesForMarket", func(t *testing.T) {
				t.Parallel()
				testGetCompletedTradesByMarket(t, repo)
			})

			t.Run("testGetTradeWithSwapAcceptID", func(t *testing.T) {
				t.Parallel()
				testGetTradeBySwapAcceptID(t, repo)
			})

			// TODO: uncomment - the following test demonstrate that in case of error,
			// any change to the db is rolled back. Currently, the transactional
			// component is not properly implemented for inmemory RepoManager and the
			// test below would fail for te inmemory implementation.

			// t.Run("testUpdateTrade_rollback", func(t *testing.T) {
			// 	t.Parallel()
			// 	testUpdateTradeRollback(t, repo)
			// })
		})
	}
}

func testGetOrCreateTrade(t *testing.T, repo tradeRepository) {
	iTrade, err := repo.write(func(ctx context.Context) (interface{}, error) {
		return repo.Repository.GetOrCreateTrade(ctx, nil)
	})
	require.NoError(t, err)
	trade, ok := iTrade.(*domain.Trade)
	require.True(t, ok)
	require.NotNil(t, trade)
}

func testGetAllTrades(t *testing.T, repo tradeRepository) {
	iTrades, err := repo.write(func(ctx context.Context) (interface{}, error) {
		_, err := repo.Repository.GetOrCreateTrade(ctx, nil)
		if err != nil {
			return nil, err
		}
		return repo.Repository.GetAllTrades(ctx, nil)
	})
	require.NoError(t, err)
	trades, ok := iTrades.([]*domain.Trade)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(trades), 1)
}

func testGetAllTradesByMarket(t *testing.T, repo tradeRepository) {
	marketAsset := randomString(32)
	var tradeID uuid.UUID

	iTrades, err := repo.write(func(ctx context.Context) (interface{}, error) {
		trade, err := repo.Repository.GetOrCreateTrade(ctx, nil)
		if err != nil {
			return nil, err
		}
		tradeID = trade.ID
		return repo.Repository.GetAllTradesByMarket(ctx, marketAsset, nil)
	})
	require.NoError(t, err)
	trades, ok := iTrades.([]*domain.Trade)
	require.True(t, ok)
	require.Len(t, trades, 0)

	iTrades, err = repo.write(func(ctx context.Context) (interface{}, error) {
		if err := repo.Repository.UpdateTrade(
			ctx,
			&tradeID,
			func(trade *domain.Trade) (*domain.Trade, error) {
				trade.MarketQuoteAsset = marketAsset
				return trade, nil
			},
		); err != nil {
			return nil, err
		}
		return repo.Repository.GetAllTradesByMarket(ctx, marketAsset, nil)
	})
	require.NoError(t, err)
	trades, ok = iTrades.([]*domain.Trade)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(trades), 1)
}

func testGetCompletedTradesByMarket(t *testing.T, repo tradeRepository) {
	marketAsset := randomString(32)
	var tradeID uuid.UUID

	iTrades, err := repo.write(func(ctx context.Context) (interface{}, error) {
		trade, err := repo.Repository.GetOrCreateTrade(ctx, nil)
		if err != nil {
			return nil, err
		}
		if err := repo.Repository.UpdateTrade(
			ctx,
			&trade.ID,
			func(trade *domain.Trade) (*domain.Trade, error) {
				trade.MarketQuoteAsset = marketAsset
				return trade, nil
			},
		); err != nil {
			return nil, err
		}
		tradeID = trade.ID
		return repo.Repository.GetCompletedTradesByMarket(ctx, marketAsset, nil)
	})
	require.NoError(t, err)
	trades, ok := iTrades.([]*domain.Trade)
	require.True(t, ok)
	require.Len(t, trades, 0)

	iTrades, err = repo.write(func(ctx context.Context) (interface{}, error) {
		if err := repo.Repository.UpdateTrade(
			ctx,
			&tradeID,
			func(trade *domain.Trade) (*domain.Trade, error) {
				trade.Status = domain.CompletedStatus
				return trade, nil
			},
		); err != nil {
			return nil, err
		}
		return repo.Repository.GetCompletedTradesByMarket(ctx, marketAsset, nil)
	})
	require.NoError(t, err)
	trades, ok = iTrades.([]*domain.Trade)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(trades), 1)

	iTrades, err = repo.write(func(ctx context.Context) (interface{}, error) {
		trade, err := repo.Repository.GetOrCreateTrade(ctx, nil)
		if err != nil {
			return nil, err
		}
		if err := repo.Repository.UpdateTrade(
			ctx,
			&trade.ID,
			func(trade *domain.Trade) (*domain.Trade, error) {
				trade.MarketQuoteAsset = marketAsset
				trade.Status = domain.SettledStatus
				return trade, nil
			},
		); err != nil {
			return nil, err
		}
		return repo.Repository.GetCompletedTradesByMarket(ctx, marketAsset, nil)
	})
	require.NoError(t, err)
	trades, ok = iTrades.([]*domain.Trade)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(trades), 2)
}

func testGetTradeBySwapAcceptID(t *testing.T, repo tradeRepository) {
	swapAcceptID := uuid.New().String()

	iTrades, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.GetOrCreateTrade(ctx, nil); err != nil {
			return nil, err
		}
		return repo.Repository.GetTradeBySwapAcceptID(ctx, swapAcceptID)
	})
	require.NoError(t, err)
	require.Nil(t, iTrades)
}

func testGetTradeByTxID(t *testing.T, repo tradeRepository) {
	txId := randomString(32)
	var tradeId uuid.UUID

	iTrades, err := repo.write(func(ctx context.Context) (interface{}, error) {
		trade, err := repo.Repository.GetOrCreateTrade(ctx, nil)
		if err != nil {
			return nil, err
		}
		tradeId = trade.ID
		return repo.Repository.GetTradeByTxID(ctx, txId)
	})
	require.NoError(t, err)
	trades, ok := iTrades.([]*domain.Trade)
	require.True(t, ok)
	require.Len(t, trades, 0)

	iTrades, err = repo.write(func(ctx context.Context) (interface{}, error) {
		if err := repo.Repository.UpdateTrade(
			ctx,
			&tradeId,
			func(trade *domain.Trade) (*domain.Trade, error) {
				trade.TxID = txId
				return trade, nil
			},
		); err != nil {
			return nil, err
		}
		return repo.Repository.GetTradeByTxID(ctx, txId)
	})
	require.NoError(t, err)
	trades, ok = iTrades.([]*domain.Trade)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(trades), 1)
}

func testUpdateTradeRollback(t *testing.T, repo tradeRepository) {
	expectedErr := errors.New("something went wrong")
	var tradeID uuid.UUID

	trade, err := repo.write(func(ctx context.Context) (interface{}, error) {
		trade, err := repo.Repository.GetOrCreateTrade(ctx, nil)
		if err != nil {
			return nil, err
		}
		tradeID = trade.ID

		if err := repo.Repository.UpdateTrade(
			ctx,
			&trade.ID,
			func(v *domain.Trade) (*domain.Trade, error) {
				return nil, expectedErr
			},
		); err != nil {
			return nil, err
		}

		return repo.Repository.GetOrCreateTrade(ctx, &tradeID)
	})
	require.EqualError(t, err, expectedErr.Error())
	require.Nil(t, trade)

	trade, err = repo.read(func(ctx context.Context) (interface{}, error) {
		return repo.Repository.GetOrCreateTrade(ctx, &tradeID)
	})
	require.Error(t, err)
	require.Nil(t, trade)
}

func createTradeRepositories(t *testing.T) []tradeRepository {
	inmemoryDBManager := inmemory.NewRepoManager()
	badgerDBManager, err := dbbadger.NewRepoManager("", nil)
	require.NoError(t, err)

	return []tradeRepository{
		{
			Name:       "badger",
			DBManager:  badgerDBManager,
			Repository: badgerDBManager.TradeRepository(),
		},
		{
			Name:       "inmemory",
			DBManager:  inmemoryDBManager,
			Repository: inmemoryDBManager.TradeRepository(),
		},
	}
}

type tradeRepository struct {
	Name       string
	DBManager  ports.RepoManager
	Repository domain.TradeRepository
}

func (r tradeRepository) read(query func(context.Context) (interface{}, error)) (interface{}, error) {
	return r.DBManager.RunTransaction(context.Background(), true, query)
}

func (r tradeRepository) write(query func(context.Context) (interface{}, error)) (interface{}, error) {
	return r.DBManager.RunTransaction(context.Background(), false, query)
}
