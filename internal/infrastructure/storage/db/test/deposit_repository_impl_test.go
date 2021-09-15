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
			t.Run("testAddAndListDeposits", func(t *testing.T) {
				testAddAndListDeposits(t, repo)
			})

			t.Run("testDepositDuplicateKeyInsertion", func(t *testing.T) {
				testDepositDuplicateKeyInsertion(t, repo)
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

	deposits, err := depositRepository.ListDepositsForAccountId(
		context.Background(),
		1,
		&domain.Page{
			Number: 2,
			Size:   10,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 10, len(deposits))

	deposits, err = depositRepository.ListDepositsForAccountId(
		context.Background(),
		1,
		&domain.Page{
			Number: 4,
			Size:   10,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 10, len(deposits))

	deposits, err = depositRepository.ListAllDeposits(
		context.Background(), nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 100, len(deposits))
}

func testDepositDuplicateKeyInsertion(t *testing.T, repo depositRepository) {
	depositRepository := repo.Repository

	err := depositRepository.AddDeposit(
		context.Background(),
		domain.Deposit{
			TxID:         "tx",
			AccountIndex: 1,
			VOut:         1,
			Asset:        "dummy",
			Value:        400,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = depositRepository.AddDeposit(
		context.Background(),
		domain.Deposit{
			TxID:         "tx",
			AccountIndex: 1,
			VOut:         1,
			Asset:        "dummy",
			Value:        500,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	deposits, err := depositRepository.ListAllDeposits(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	//length increased by 1 and not two
	assert.Equal(t, 101, len(deposits))
	//first inserted is not updated
	assert.Equal(t, 400, int(deposits[0].Value))
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
