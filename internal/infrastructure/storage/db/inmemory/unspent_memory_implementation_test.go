package inmemory

import (
	"context"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestAddUnspentAndBalance(t *testing.T) {
	repo := NewUnspentRepositoryImpl()
	ctx := context.Background()

	u1 := domain.NewUnspent(
		"1",
		"lbtc",
		"adr",
		0,
		1,
		false,
		false,
		nil,
		nil,
		true,
	)

	u2 := domain.NewUnspent(
		"2",
		"lbtc",
		"adr",
		0,
		0,
		false,
		false,
		nil,
		nil,
		true,
	)

	unspents := []domain.Unspent{u1, u2}
	repo.AddUnspents(ctx, unspents)

	allUnspent := repo.GetAllUnspents(ctx)
	allSpent := repo.GetAllSpents(ctx)
	assert.Equal(t, len(allUnspent), 2)
	assert.Equal(t, len(allSpent), 0)

	unspents = []domain.Unspent{u2}
	repo.AddUnspents(ctx, unspents)

	allUnspent = repo.GetAllUnspents(ctx)
	allSpent = repo.GetAllSpents(ctx)
	assert.Equal(t, len(allUnspent), 1)
	assert.Equal(t, len(allSpent), 1)

	u3 := domain.NewUnspent(
		"3",
		"lbtc",
		"adr",
		0,
		3,
		false,
		false,
		nil,
		nil,
		true,
	)

	u4 := domain.NewUnspent(
		"4",
		"lbtc",
		"adr",
		0,
		2,
		false,
		false,
		nil,
		nil,
		true,
	)

	u5 := domain.NewUnspent(
		"5",
		"lbtc",
		"adr",
		0,
		2,
		false,
		false,
		nil,
		nil,
		true,
	)

	unspents = []domain.Unspent{u3, u4, u5}
	repo.AddUnspents(ctx, unspents)

	allUnspent = repo.GetAllUnspents(ctx)
	allSpent = repo.GetAllSpents(ctx)

	assert.Equal(t, len(allUnspent), 3)
	assert.Equal(t, len(allSpent), 2)

	balance := repo.GetBalance(ctx, "adr", "lbtc")

	assert.Equal(t, balance, uint64(7))

}
