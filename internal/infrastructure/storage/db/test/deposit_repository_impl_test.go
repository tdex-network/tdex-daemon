package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
)

func TestDepositRepositoryImplementations(t *testing.T) {
	repositories := createDepositRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Run("add_and_get_deposits", func(t *testing.T) {
				testAddAndGetDeposits(t, repo)
			})
		})
	}
}

func testAddAndGetDeposits(t *testing.T, repo depositRepository) {
	depositRepository := repo.Repository
	ctx := context.Background()
	deposits := makeRandomDeposits(20)

	allDeposits, err := depositRepository.GetAllDeposits(ctx, nil)
	require.NoError(t, err)
	require.Empty(t, allDeposits)

	count, err := depositRepository.AddDeposits(ctx, deposits)
	require.NoError(t, err)
	require.Equal(t, 20, count)

	allDeposits, err = depositRepository.GetAllDeposits(ctx, nil)
	require.NoError(t, err)
	require.Len(t, allDeposits, 20)

	count, err = depositRepository.AddDeposits(ctx, deposits)
	require.NoError(t, err)
	require.Zero(t, count)

	// Test that pagination is correct by getting all 20 deposits in 4 pages,
	// each including 5 items. The concatenation of all pages must match the
	// non-paginated list item per item.
	allPagedDeposits := make([]domain.Deposit, 0)
	for i := 1; i < 5; i++ {
		pagedDeposits, err := depositRepository.GetAllDeposits(
			ctx, page{int64(i), 5},
		)
		require.NoError(t, err)
		require.Len(t, pagedDeposits, 5)
		allPagedDeposits = append(allPagedDeposits, pagedDeposits...)
	}
	require.Exactly(t, allPagedDeposits, allDeposits)

	depositsByMarket, err := depositRepository.GetDepositsForAccount(
		ctx, deposits[0].AccountName, nil,
	)
	require.NoError(t, err)
	require.Len(t, depositsByMarket, 1)

	depositsByMarket, err = depositRepository.GetDepositsForAccount(
		ctx, randomHex(20), nil,
	)
	require.NoError(t, err)
	require.Empty(t, depositsByMarket)
}

type depositRepository struct {
	Name       string
	Repository domain.DepositRepository
}

func createDepositRepositories(t *testing.T) []depositRepository {
	inmemoryDBManager := inmemory.NewRepoManager()
	badgerDBManager, err := dbbadger.NewRepoManager("", nil)
	require.NoError(t, err)

	return []depositRepository{
		{
			Name:       "badger",
			Repository: badgerDBManager.DepositRepository(),
		},
		{
			Name:       "inmemory",
			Repository: inmemoryDBManager.DepositRepository(),
		},
	}
}
