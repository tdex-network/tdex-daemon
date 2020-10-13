package inmemory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/vulpemventures/go-elements/network"
)

func TestAddUnspentAndBalance(t *testing.T) {
	repo := NewUnspentRepositoryImpl()
	ctx := context.Background()

	u1 := domain.NewUnspent(
		"0000000000000000000000000000000000000000000000000000000000000000", // txid
		0,                       // vout
		100000000,               // value
		network.Regtest.AssetID, // asset
		make([]byte, 22),        // script
		"080000000000000000000000000000000000000000000000000000000000000000", // value commitment
		"090000000000000000000000000000000000000000000000000000000000000000", // asset commitment
		make([]byte, 33),   // nonce
		make([]byte, 4174), // range proof
		make([]byte, 64),   // surjection proof
		"el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc", // address
		true, // confirmed
	)

	u2 := domain.NewUnspent(
		"0000000000000000000000000000000000000000000000000000000000000001", // txid
		1,                       // vout
		150000000,               // value
		network.Regtest.AssetID, // asset
		make([]byte, 22),        // script
		"080000000000000000000000000000000000000000000000000000000000000000", // value commitment
		"090000000000000000000000000000000000000000000000000000000000000000", // asset commitment
		make([]byte, 33),   // nonce
		make([]byte, 4174), // range proof
		make([]byte, 64),   // surjection proof
		"el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc", // address
		true, // confirmed
	)

	unspents := []domain.Unspent{u1, u2}
	repo.AddUnspents(ctx, unspents)

	allUnspent := repo.GetAllUnspents(ctx)
	allSpent := repo.GetAllSpents(ctx)
	assert.Equal(t, 2, len(allUnspent))
	assert.Equal(t, 0, len(allSpent))

	unspents = []domain.Unspent{u2}
	repo.AddUnspents(ctx, unspents)

	allUnspent = repo.GetAllUnspents(ctx)
	allSpent = repo.GetAllSpents(ctx)

	assert.Equal(t, 1, len(allUnspent))
	assert.Equal(t, 1, len(allSpent))

	u3 := domain.NewUnspent(
		"0000000000000000000000000000000000000000000000000000000000000002", // txid
		0,           // vout
		10000000000, // value
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // asset
		make([]byte, 22), // script
		"080000000000000000000000000000000000000000000000000000000000000000", // value commitment
		"090000000000000000000000000000000000000000000000000000000000000000", // asset commitment
		make([]byte, 33),   // nonce
		make([]byte, 4174), // range proof
		make([]byte, 64),   // surjection proof
		"el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc", // address
		true, // confirmed
	)

	u4 := domain.NewUnspent(
		"0000000000000000000000000000000000000000000000000000000000000003", // txid
		1,            // vout
		650000000000, // value
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", // asset
		make([]byte, 22), // script
		"080000000000000000000000000000000000000000000000000000000000000000", // value commitment
		"090000000000000000000000000000000000000000000000000000000000000000", // asset commitment
		make([]byte, 33),   // nonce
		make([]byte, 4174), // range proof
		make([]byte, 64),   // surjection proof
		"el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc", // address
		true, // confirmed
	)

	u5 := domain.NewUnspent(
		"0000000000000000000000000000000000000000000000000000000000000004", // txid
		1,           // vout
		30000000000, // value
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // asset
		make([]byte, 22), // script
		"080000000000000000000000000000000000000000000000000000000000000000", // value commitment
		"090000000000000000000000000000000000000000000000000000000000000000", // asset commitment
		make([]byte, 33),   // nonce
		make([]byte, 4174), // range proof
		make([]byte, 64),   // surjection proof
		"el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc", // address
		true, // confirmed
	)

	unspents = []domain.Unspent{u3, u4, u5}
	repo.AddUnspents(ctx, unspents)

	allUnspent = repo.GetAllUnspents(ctx)
	allSpent = repo.GetAllSpents(ctx)

	assert.Equal(t, 3, len(allUnspent))
	assert.Equal(t, 2, len(allSpent))

	balance := repo.GetBalance(
		ctx,
		"el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	)

	assert.Equal(t, 40000000000, int(balance))
}
