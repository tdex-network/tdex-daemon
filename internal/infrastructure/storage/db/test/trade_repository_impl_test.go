package db_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
)

func TestTradeRepositoryImplementations(t *testing.T) {
	repositories := createTradeRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("add_get_trade", func(t *testing.T) {
				testAddAndGetTrade(t, repo.Repository)
			})

			t.Run("update_trade", func(t *testing.T) {
				testUpdateTrade(t, repo.Repository)
			})

			t.Run("get_completed_trades", func(t *testing.T) {
				testGetCompletedTrades(t, repo.Repository)
			})
		})
	}
}

func testAddAndGetTrade(t *testing.T, repo domain.TradeRepository) {
	ctx := context.Background()
	trade := makeRandomTrade()

	allTrades, err := repo.GetAllTrades(ctx, nil)
	require.NoError(t, err)
	require.Empty(t, allTrades)

	err = repo.AddTrade(ctx, trade)
	require.NoError(t, err)
	err = repo.AddTrade(ctx, trade)
	require.Error(t, err)

	allTrades, err = repo.GetAllTrades(ctx, nil)
	require.NoError(t, err)
	require.Len(t, allTrades, 1)

	trades, err := repo.GetAllTradesByMarket(ctx, trade.MarketName, nil)
	require.NoError(t, err)
	require.Len(t, trades, 1)
	require.Exactly(t, allTrades, trades)

	trades, err = repo.GetAllTradesByMarket(ctx, randomHex(20), nil)
	require.NoError(t, err)
	require.Empty(t, trades)
}

func testUpdateTrade(t *testing.T, repo domain.TradeRepository) {
	ctx := context.Background()

	trades, err := repo.GetAllTrades(ctx, nil)
	require.NoError(t, err)
	require.NotEmpty(t, trades)

	trade := trades[0]
	require.False(t, trade.IsCompleted())

	err = repo.UpdateTrade(
		ctx, trade.Id, func(tt *domain.Trade) (*domain.Trade, error) {
			return nil, fmt.Errorf("something went wrong")
		},
	)
	require.Error(t, err)
	require.EqualError(t, err, "something went wrong")

	err = repo.UpdateTrade(
		ctx, trade.Id, func(tt *domain.Trade) (*domain.Trade, error) {
			tt.Status.Code = domain.TradeStatusCodeCompleted
			tt.SwapAccept = &domain.Swap{Id: randomId()}
			tt.TxId = randomHex(32)
			return tt, nil
		},
	)
	require.NoError(t, err)

	trades, err = repo.GetAllTrades(ctx, nil)
	require.NoError(t, err)
	require.NotEmpty(t, trades)

	trade = trades[0]
	require.True(t, trade.IsCompleted())
}

func testGetCompletedTrades(t *testing.T, repo domain.TradeRepository) {
	ctx := context.Background()

	trades, err := repo.GetAllTrades(ctx, nil)
	require.NoError(t, err)
	require.NotEmpty(t, trades)

	trade := trades[0]

	completedTrades, err := repo.GetCompletedTradesByMarket(
		ctx, randomHex(20), nil,
	)
	require.NoError(t, err)
	require.Empty(t, completedTrades)

	completedTrades, err = repo.GetCompletedTradesByMarket(
		ctx, trade.MarketName, nil,
	)
	require.NoError(t, err)
	require.Len(t, completedTrades, 1)
	require.Exactly(t, trade, completedTrades[0])

	foundTrade, err := repo.GetTradeBySwapAcceptId(ctx, randomId())
	require.Error(t, err)
	require.Nil(t, foundTrade)

	foundTrade, err = repo.GetTradeBySwapAcceptId(ctx, trade.SwapAccept.Id)
	require.NoError(t, err)
	require.NotNil(t, foundTrade)
	require.Exactly(t, trade, *foundTrade)

	foundTrade, err = repo.GetTradeByTxId(ctx, randomHex(32))
	require.NoError(t, err)
	require.Nil(t, foundTrade)

	foundTrade, err = repo.GetTradeByTxId(ctx, trade.TxId)
	require.NoError(t, err)
	require.NotNil(t, foundTrade)
	require.Exactly(t, trade, *foundTrade)
}

func createTradeRepositories(t *testing.T) []tradeRepository {
	inmemoryDBManager := inmemory.NewRepoManager()
	badgerDBManager, err := dbbadger.NewRepoManager("", nil)
	require.NoError(t, err)

	return []tradeRepository{
		{
			Name:       "badger",
			Repository: badgerDBManager.TradeRepository(),
		},
		{
			Name:       "inmemory",
			Repository: inmemoryDBManager.TradeRepository(),
		},
	}
}

type tradeRepository struct {
	Name       string
	Repository domain.TradeRepository
}
