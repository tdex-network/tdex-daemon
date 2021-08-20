package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
)

func TestAddAndListWithdrawals(t *testing.T) {
	badgerDBManager, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		t.Fatal(err)
	}

	statsRepository := badgerDBManager.StatsRepository()
	for i := 0; i < 100; i++ {
		err = statsRepository.AddWithdrawal(
			context.Background(),
			domain.Withdrawal{
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

	withdrawals, err := statsRepository.ListWithdrawalsForAccountIdAndPage(
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
	assert.Equal(t, int(withdrawals[0].ID), 11)

	withdrawals, err = statsRepository.ListWithdrawalsForAccountIdAndPage(
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
	assert.Equal(t, int(withdrawals[0].ID), 31)
}

func TestAddAndListDeposits(t *testing.T) {
	badgerDBManager, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		t.Fatal(err)
	}

	statsRepository := badgerDBManager.StatsRepository()
	for i := 0; i < 100; i++ {
		err = statsRepository.AddDeposit(
			context.Background(),
			domain.Deposit{
				AccountIndex: 1,
				TxID:         "3232",
				VOut:         i,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	withdrawals, err := statsRepository.ListDepositsForAccountIdAndPage(
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
	assert.Equal(t, int(withdrawals[0].VOut), 11)

	withdrawals, err = statsRepository.ListDepositsForAccountIdAndPage(
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
	assert.Equal(t, int(withdrawals[0].VOut), 31)
}
