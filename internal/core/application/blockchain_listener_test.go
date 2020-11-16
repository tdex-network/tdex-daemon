package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
)

func TestUpdateUnspentsForAddress(t *testing.T) {
	dbManager := newTestDb()
	unspentRepository := inmemory.NewUnspentRepositoryImpl(dbManager)
	ctx := context.Background()

	l := newBlockchainListener(
		unspentRepository,
		nil,
		nil,
		nil,
		nil,
		dbManager)

	unspents := []domain.Unspent{
		{
			TxID:         "1",
			VOut:         1,
			Value:        0,
			AssetHash:    "",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
		{
			TxID:         "2",
			VOut:         2,
			Value:        0,
			AssetHash:    "",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
	}

	err := l.updateUnspentsForAddress(
		ctx,
		unspents,
		"a",
	)
	if err != nil {
		t.Fatal(err)
	}

	unsp, err := unspentRepository.GetAllUnspentsForAddresses(
		ctx,
		[]string{"a"},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(unsp))

	unspents = []domain.Unspent{
		{
			TxID:         "1",
			VOut:         1,
			Value:        0,
			AssetHash:    "",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
		{
			TxID:         "4",
			VOut:         2,
			Value:        0,
			AssetHash:    "",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
	}

	err = l.updateUnspentsForAddress(
		ctx,
		unspents,
		"a",
	)
	if err != nil {
		t.Fatal(err)
	}

	unsp, err = unspentRepository.GetAllUnspentsForAddresses(
		ctx,
		[]string{"a"},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(unsp))

	count := 0
	for _, v := range unsp {
		if v.IsSpent() == true {
			count++
		}
	}

	assert.Equal(t, 1, count)
}
