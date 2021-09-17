package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
)

func TestWithdrawalRepositoryImplementations(t *testing.T) {
	repositories := createWithdrawalRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Run("testAddAndListWithdrawals", func(t *testing.T) {
				testAddAndListWithdrawals(t, repo)
			})
		})
	}
}

func testAddAndListWithdrawals(t *testing.T, repo withdrawalRepository) {
	depositRepository := repo.Repository
	withdrawals := make([]domain.Withdrawal, 0)
	for i := 0; i < 10; i++ {
		withdrawals = append(withdrawals, domain.Withdrawal{
			TxID:            fmt.Sprintf("%d", i),
			AccountIndex:    1,
			BaseAmount:      20,
			QuoteAmount:     20,
			MillisatPerByte: 10,
			Address:         "dwd",
		})
	}

	count, err := depositRepository.AddWithdrawals(
		context.Background(), withdrawals,
	)
	require.NoError(t, err)
	require.Equal(t, 10, count)

	count, err = depositRepository.AddWithdrawals(
		context.Background(), []domain.Withdrawal{
			{
				TxID:            "0",
				AccountIndex:    1,
				BaseAmount:      20,
				QuoteAmount:     20,
				MillisatPerByte: 10,
				Address:         "dwd",
			},
		},
	)
	require.NoError(t, err)
	require.Zero(t, count)

	withdrawals, err = depositRepository.ListWithdrawalsForAccount(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, withdrawals, 10)

	withdrawals, err = depositRepository.ListWithdrawalsForAccount(context.Background(), 0)
	require.NoError(t, err)
	require.Zero(t, withdrawals)

	withdrawals, err = depositRepository.ListWithdrawalsForAccountAndPage(
		context.Background(), 1, domain.Page{Number: 1, Size: 5},
	)
	require.NoError(t, err)
	require.Len(t, withdrawals, 5)

	withdrawals, err = depositRepository.ListWithdrawalsForAccountAndPage(
		context.Background(), 1, domain.Page{Number: 2, Size: 5},
	)
	require.NoError(t, err)
	require.Len(t, withdrawals, 5)

	withdrawals, err = depositRepository.ListAllWithdrawals(context.Background())
	require.NoError(t, err)
	require.Len(t, withdrawals, 10)

	withdrawals, err = depositRepository.ListAllWithdrawalsForPage(
		context.Background(), domain.Page{Number: 1, Size: 6},
	)
	require.NoError(t, err)
	require.Len(t, withdrawals, 6)

	withdrawals, err = depositRepository.ListAllWithdrawalsForPage(
		context.Background(), domain.Page{Number: 2, Size: 6},
	)
	require.NoError(t, err)
	require.Len(t, withdrawals, 4)
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
