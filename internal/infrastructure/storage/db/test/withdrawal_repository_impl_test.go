package db

import (
	"context"
	"strconv"
	"testing"

	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"

	"github.com/stretchr/testify/assert"
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

			t.Run("testWithdrawalDuplicateKeyInsertion", func(t *testing.T) {
				testWithdrawalDuplicateKeyInsertion(t, repo)
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

	withdrawals, err = depositRepository.ListAllWithdrawals(
		context.Background(),
	)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 100, len(withdrawals))
}

func testWithdrawalDuplicateKeyInsertion(t *testing.T, repo withdrawalRepository) {
	withdrawalRepository := repo.Repository

	err := withdrawalRepository.AddWithdrawal(
		context.Background(),
		domain.Withdrawal{
			TxID:            "tx",
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

	err = withdrawalRepository.AddWithdrawal(
		context.Background(),
		domain.Withdrawal{
			TxID:            "tx",
			AccountIndex:    1,
			BaseAmount:      20,
			QuoteAmount:     20,
			MillisatPerByte: 10,
			Address:         "dwd1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	withdrwals, err := withdrawalRepository.ListAllWithdrawals(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	//length increased by 1 and not two
	assert.Equal(t, 101, len(withdrwals))
	//first inserted is not updated
	assert.Equal(t, "dwd", withdrwals[0].Address)
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
