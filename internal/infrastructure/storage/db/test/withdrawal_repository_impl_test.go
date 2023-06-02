package db_test

import (
	"context"
	"testing"

	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
)

func TestWithdrawalRepositoryPgImplementation(t *testing.T) {
	if err := SetupPgDb(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := TearDownPgDb()
		if err != nil {
			t.Fatal(err)
		}
	}()

	testAddAndGetWithdrawals(t, repoManager.WithdrawalRepository())
}

func TestWithdrawalRepositoryImplementations(t *testing.T) {
	repositories := createWithdrawalRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("add_and_get_withdrawals", func(t *testing.T) {
				testAddAndGetWithdrawals(t, repo.Repository)
			})
		})
	}
}

func testAddAndGetWithdrawals(
	t *testing.T, withdrawalRepository domain.WithdrawalRepository,
) {
	ctx := context.Background()
	withdrawals := makeRandomWithdrawals(20)

	allWithdrawals, err := withdrawalRepository.GetAllWithdrawals(ctx, nil)
	require.NoError(t, err)
	require.Empty(t, allWithdrawals)

	count, err := withdrawalRepository.AddWithdrawals(
		context.Background(), withdrawals,
	)
	require.NoError(t, err)
	require.Equal(t, 20, count)

	allWithdrawals, err = withdrawalRepository.GetAllWithdrawals(ctx, nil)
	require.NoError(t, err)
	require.Len(t, allWithdrawals, 20)

	count, err = withdrawalRepository.AddWithdrawals(
		context.Background(), withdrawals,
	)
	require.NoError(t, err)
	require.Zero(t, count)

	// Test that pagination is correct by getting all 20 withdrawals in 4 pages,
	// each including 5 items. The concatenation of all pages must match the
	// non-paginated list item per item.
	allPagedWithdrawals := make([]domain.Withdrawal, 0)
	for i := 1; i < 5; i++ {
		pagedWithdrawals, err := withdrawalRepository.GetAllWithdrawals(
			ctx, page{int64(i), 5},
		)
		require.NoError(t, err)
		require.NotEmpty(t, pagedWithdrawals)
		allPagedWithdrawals = append(allPagedWithdrawals, pagedWithdrawals...)
	}
	require.Exactly(t, allPagedWithdrawals, allWithdrawals)

	withdrawalsByMarket, err := withdrawalRepository.GetWithdrawalsForAccount(
		ctx, withdrawals[0].AccountName, nil,
	)
	require.NoError(t, err)
	require.Len(t, withdrawalsByMarket, 1)

	withdrawalsByMarket, err = withdrawalRepository.GetWithdrawalsForAccount(
		ctx, randomHex(32), nil,
	)
	require.NoError(t, err)
	require.Empty(t, withdrawalsByMarket)
}

type withdrawalRepository struct {
	Name       string
	DBManager  ports.RepoManager
	Repository domain.WithdrawalRepository
}

func createWithdrawalRepositories(t *testing.T) []withdrawalRepository {
	inmemoryDBManager := inmemory.NewRepoManager()
	badgerDBManager, err := dbbadger.NewRepoManager("", nil)
	require.NoError(t, err)

	return []withdrawalRepository{
		{
			Name:       "badger",
			DBManager:  badgerDBManager,
			Repository: badgerDBManager.WithdrawalRepository(),
		},
		{
			Name:       "inmemory",
			DBManager:  inmemoryDBManager,
			Repository: inmemoryDBManager.WithdrawalRepository(),
		},
	}
}
