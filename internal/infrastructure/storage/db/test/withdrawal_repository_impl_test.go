package db

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"

	"github.com/stretchr/testify/assert"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
)

func TestWithdrawalRepositoryImplementations(t *testing.T) {
	repositories := createWithdrawalRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("testAddAndListWithdrawals", func(t *testing.T) {
				t.Parallel()
				testAddAndListWithdrawals(t, repo)
			})
		})
	}
}

func testAddAndListWithdrawals(t *testing.T, repo withdrawalRepository) {

	depositRepository := repo.Repository
	for i := 0; i < 100; i++ {
		err := depositRepository.AddWithdrawal(
			context.Background(),
			domain.Withdrawal{
				TxID:            strconv.Itoa(i),
				AccountIndex:    1,
				BaseAmount:      20,
				QuoteAmount:     20,
				MillisatPerByte: 10,
				Address:         "dwd",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	withdrawals, err := depositRepository.ListWithdrawalsForAccountIdAndPage(
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

	withdrawals, err = depositRepository.ListWithdrawalsForAccountIdAndPage(
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
