package storage

import (
	"github.com/magiconair/properties/assert"
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"testing"
)

func TestAddUnspentAndBalance(t *testing.T) {
	repo := NewInMemoryUnspentRepository()

	u1 := unspent.NewUnspent(
		"1",
		"lbtc",
		"adr",
		0,
		1,
		false,
		false,
		nil,
	)

	u2 := unspent.NewUnspent(
		"2",
		"lbtc",
		"adr",
		0,
		0,
		false,
		false,
		nil,
	)

	unspents := []unspent.Unspent{u1, u2}
	repo.AddUnspent(unspents)

	allUnspent := repo.GetAllUnspent()
	allSpent := repo.GetAllSpent()
	assert.Equal(t, len(allUnspent), 2)
	assert.Equal(t, len(allSpent), 0)

	unspents = []unspent.Unspent{u2}
	repo.AddUnspent(unspents)

	allUnspent = repo.GetAllUnspent()
	allSpent = repo.GetAllSpent()
	assert.Equal(t, len(allUnspent), 1)
	assert.Equal(t, len(allSpent), 1)

	u3 := unspent.NewUnspent(
		"3",
		"lbtc",
		"adr",
		0,
		3,
		false,
		false,
		nil,
	)

	u4 := unspent.NewUnspent(
		"4",
		"lbtc",
		"adr",
		0,
		2,
		false,
		false,
		nil,
	)

	u5 := unspent.NewUnspent(
		"5",
		"lbtc",
		"adr",
		0,
		2,
		false,
		false,
		nil,
	)

	unspents = []unspent.Unspent{u3, u4, u5}
	repo.AddUnspent(unspents)

	allUnspent = repo.GetAllUnspent()
	allSpent = repo.GetAllSpent()

	assert.Equal(t, len(allUnspent), 3)
	assert.Equal(t, len(allSpent), 2)

	balance := repo.GetBalance("adr", "lbtc")

	assert.Equal(t, balance, uint64(7))

}
