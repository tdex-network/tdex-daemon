package dbbadger

import (
	"github.com/magiconair/properties/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"testing"
)

func TestAddUnspents(t *testing.T) {
	before()
	defer after()
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
			Address:      "b",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
	}
	err := unspentRepository.AddUnspents(ctx, unspents)
	if err != nil {
		t.Fatal(err)
	}

	u := unspentRepository.GetAllUnspents(ctx)

	assert.Equal(t, len(u), 2)
	assert.Equal(t, u[0].TxID == "1", true)
	assert.Equal(t, u[1].TxID == "2", true)
}
