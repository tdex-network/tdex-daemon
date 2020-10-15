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

	u1 := domain.Unspent{
		TxID:            "0000000000000000000000000000000000000000000000000000000000000000",
		VOut:            0,
		Value:           100000000,
		AssetHash:       network.Regtest.AssetID,
		ValueCommitment: "080000000000000000000000000000000000000000000000000000000000000000",
		AssetCommitment: "090000000000000000000000000000000000000000000000000000000000000000",
		ScriptPubKey:    make([]byte, 22),
		Nonce:           make([]byte, 33),
		RangeProof:      make([]byte, 4174),
		SurjectionProof: make([]byte, 64),
		Address:         "el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc",
		Spent:           false,
		Locked:          false,
		LockedBy:        nil,
		Confirmed:       true,
	}

	u2 := domain.Unspent{
		TxID:            "0000000000000000000000000000000000000000000000000000000000000001",
		VOut:            1,
		Value:           150000000,
		AssetHash:       network.Regtest.AssetID,
		ValueCommitment: "080000000000000000000000000000000000000000000000000000000000000000",
		AssetCommitment: "090000000000000000000000000000000000000000000000000000000000000000",
		ScriptPubKey:    make([]byte, 22),
		Nonce:           make([]byte, 33),
		RangeProof:      make([]byte, 4174),
		SurjectionProof: make([]byte, 64),
		Address:         "el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc",
		Spent:           false,
		Locked:          false,
		LockedBy:        nil,
		Confirmed:       true,
	}

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

	u3 := domain.Unspent{
		TxID:            "0000000000000000000000000000000000000000000000000000000000000002",
		VOut:            0,
		Value:           10000000000,
		AssetHash:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ValueCommitment: "080000000000000000000000000000000000000000000000000000000000000000",
		AssetCommitment: "090000000000000000000000000000000000000000000000000000000000000000",
		ScriptPubKey:    make([]byte, 22),
		Nonce:           make([]byte, 33),
		RangeProof:      make([]byte, 4174),
		SurjectionProof: make([]byte, 64),
		Address:         "el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc",
		Spent:           false,
		Locked:          false,
		LockedBy:        nil,
		Confirmed:       true,
	}

	u4 := domain.Unspent{
		TxID:            "0000000000000000000000000000000000000000000000000000000000000003",
		VOut:            1,
		Value:           650000000000,
		AssetHash:       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ValueCommitment: "080000000000000000000000000000000000000000000000000000000000000000",
		AssetCommitment: "090000000000000000000000000000000000000000000000000000000000000000",
		ScriptPubKey:    make([]byte, 22),
		Nonce:           make([]byte, 33),
		RangeProof:      make([]byte, 4174),
		SurjectionProof: make([]byte, 64),
		Address:         "el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc",
		Spent:           false,
		Locked:          false,
		LockedBy:        nil,
		Confirmed:       true,
	}

	u5 := domain.Unspent{
		TxID:            "0000000000000000000000000000000000000000000000000000000000000004",
		VOut:            1,
		Value:           30000000000,
		AssetHash:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ValueCommitment: "080000000000000000000000000000000000000000000000000000000000000000",
		AssetCommitment: "090000000000000000000000000000000000000000000000000000000000000000",
		ScriptPubKey:    make([]byte, 22),
		Nonce:           make([]byte, 33),
		RangeProof:      make([]byte, 4174),
		SurjectionProof: make([]byte, 64),
		Address:         "el1qqfxwyst8u39d37k2mepqhlxhm9r00rqrnvhnqw444730a9frszjnw7ydmu8dm4j2n60asfw46ym6kum02e4pglsjdnyl68pfc",
		Spent:           false,
		Locked:          false,
		LockedBy:        nil,
		Confirmed:       true,
	}
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
