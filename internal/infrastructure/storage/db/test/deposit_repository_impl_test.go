package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
)

func TestT(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4} {
		fmt.Printf("%p\n", &n)
	}
}

func TestDepositRepositoryImplementations(t *testing.T) {
	repositories := createDepositRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Run("testAddAndListDeposits", func(t *testing.T) {
				testAddAndListDeposits(t, repo)
			})
		})
	}
}

func testAddAndListDeposits(t *testing.T, repo depositRepository) {
	depositRepository := repo.Repository

	deposits := make([]domain.Deposit, 0, 10)
	for i := 0; i < 10; i++ {
		deposits = append(deposits, domain.Deposit{
			TxID:         fmt.Sprintf("%d", i),
			AccountIndex: 1,
			VOut:         1,
			Asset:        "dummy",
			Value:        400,
		})
	}
	count, err := depositRepository.AddDeposits(context.Background(), deposits)
	require.NoError(t, err)
	require.Equal(t, 10, count)

	count, err = depositRepository.AddDeposits(context.Background(), []domain.Deposit{{
		TxID:         "0",
		AccountIndex: 1,
		VOut:         1,
		Asset:        "dummy",
		Value:        400,
	}})
	require.NoError(t, err)
	require.Zero(t, count)

	deposits, err = depositRepository.ListDepositsForAccount(context.Background(), 0)
	require.NoError(t, err)
	require.Empty(t, deposits)

	deposits, err = depositRepository.ListDepositsForAccount(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, deposits, 10)

	deposits, err = depositRepository.ListDepositsForAccountAndPage(
		context.Background(), 1, domain.Page{Number: 1, Size: 5},
	)
	require.NoError(t, err)
	require.Len(t, deposits, 5)

	deposits, err = depositRepository.ListDepositsForAccountAndPage(
		context.Background(), 1, domain.Page{Number: 2, Size: 5},
	)
	require.NoError(t, err)
	require.Len(t, deposits, 5)

	deposits, err = depositRepository.ListAllDeposits(context.Background())
	require.NoError(t, err)
	require.Len(t, deposits, 10)

	deposits, err = depositRepository.ListAllDepositsForPage(
		context.Background(), domain.Page{Number: 1, Size: 6},
	)
	require.NoError(t, err)
	require.Len(t, deposits, 6)

	deposits, err = depositRepository.ListAllDepositsForPage(
		context.Background(), domain.Page{Number: 2, Size: 6},
	)
	require.NoError(t, err)
	require.Len(t, deposits, 4)
}

type depositRepository struct {
	Name       string
	DBManager  ports.RepoManager
	Repository domain.DepositRepository
}

func createDepositRepositories(t *testing.T) []depositRepository {
	inmemoryDBManager := inmemory.NewRepoManager()
	badgerDBManager, err := dbbadger.NewRepoManager("", nil)
	require.NoError(t, err)

	return []depositRepository{
		{
			Name:       "badger",
			DBManager:  badgerDBManager,
			Repository: badgerDBManager.DepositRepository(),
		},
		{
			Name:       "inmemory",
			DBManager:  inmemoryDBManager,
			Repository: inmemoryDBManager.DepositRepository(),
		},
	}
}
