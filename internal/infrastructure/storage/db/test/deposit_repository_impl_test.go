package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
)

func TestDepositRepositoryImplementations(t *testing.T) {
	repositories := createDepositRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("testAddAndListDeposits", func(t *testing.T) {
				t.Parallel()
				testAddAndListDeposits(t, repo)
			})
		})
	}
}

func testAddAndListDeposits(t *testing.T, repo depositRepository) {
	depositRepository := repo.Repository
	for i := 0; i < 100; i++ {
		err := depositRepository.AddDeposit(
			context.Background(),
			domain.Deposit{
				TxID:         "3232",
				AccountIndex: 1,
				VOut:         i,
				Asset:        "dummy",
				Value:        400,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	withdrawals, err := depositRepository.ListDepositsForAccountIdAndPage(
		context.Background(),
		1,
		domain.Page{
			Number: 2,
			Size:   10,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 10, len(withdrawals))

	withdrawals, err = depositRepository.ListDepositsForAccountIdAndPage(
		context.Background(),
		1,
		domain.Page{
			Number: 4,
			Size:   10,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 10, len(withdrawals))
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
