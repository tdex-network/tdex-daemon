package storage

import (
	"github.com/magiconair/properties/assert"
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"testing"
)

func TestAddUnspentAndBalance(t *testing.T) {
	repo := NewInMemoryUnspentRepository()

	u1 := unspent.Unspent{
		Txid:      "1",
		Vout:      0,
		Value:     1,
		AssetHash: "lbtc",
		Address:   "adr",
		Spent:     false,
		Locked:    false,
	}
	u2 := unspent.Unspent{
		Txid:      "2",
		Vout:      0,
		Value:     0,
		AssetHash: "lbtc",
		Address:   "adr",
		Spent:     false,
		Locked:    false,
	}
	unspents := []unspent.Unspent{u1, u2}
	repo.AddUnspent(unspents)

	allUnspent := repo.GetAllUnspent()
	assert.Equal(t, len(allUnspent), 2)

	unspents = []unspent.Unspent{u2}
	repo.AddUnspent(unspents)

	allUnspent = repo.GetAllUnspent()
	assert.Equal(t, len(allUnspent), 2)

	assert.Equal(t, allUnspent[0].Spent, true)
	assert.Equal(t, allUnspent[1].Spent, false)

	u3 := unspent.Unspent{
		Txid:      "3",
		Vout:      0,
		Value:     3,
		AssetHash: "lbtc",
		Address:   "adr",
		Spent:     false,
		Locked:    false,
	}
	u4 := unspent.Unspent{
		Txid:      "4",
		Vout:      0,
		Value:     2,
		AssetHash: "lbtc",
		Address:   "adr",
		Spent:     false,
		Locked:    false,
	}
	u5 := unspent.Unspent{
		Txid:      "5",
		Vout:      0,
		Value:     2,
		AssetHash: "lbtc",
		Address:   "adr",
		Spent:     false,
		Locked:    false,
	}
	unspents = []unspent.Unspent{u3, u4, u5}
	repo.AddUnspent(unspents)

	allUnspent = repo.GetAllUnspent()

	assert.Equal(t, len(allUnspent), 5)

	assert.Equal(t, allUnspent[0].Spent, true)
	assert.Equal(t, allUnspent[1].Spent, true)
	assert.Equal(t, allUnspent[2].Spent, false)
	assert.Equal(t, allUnspent[3].Spent, false)
	assert.Equal(t, allUnspent[4].Spent, false)

	balance := repo.GetBalance("adr", "lbtc")

	assert.Equal(t, balance, uint64(7))

}
