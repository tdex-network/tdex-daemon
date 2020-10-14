package dbbadger

import (
	"github.com/google/uuid"
	"github.com/magiconair/properties/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"testing"
)

func TestAddUnspents(t *testing.T) {
	before()
	defer after()
	unspents := []domain.Unspent{
		{
			TxID:            "6",
			VOut:            1,
			Value:           0,
			AssetHash:       "",
			ValueCommitment: "",
			AssetCommitment: "",
			ScriptPubKey:    nil,
			Nonce:           nil,
			RangeProof:      nil,
			SurjectionProof: nil,
			Address:         "a",
			Spent:           false,
			Locked:          false,
			LockedBy:        nil,
			Confirmed:       false,
		},
		{
			TxID:            "7",
			VOut:            2,
			Value:           0,
			AssetHash:       "",
			ValueCommitment: "",
			AssetCommitment: "",
			Nonce:           nil,
			RangeProof:      nil,
			SurjectionProof: nil,
			Address:         "b",
			Spent:           false,
			Locked:          false,
			ScriptPubKey:    nil,
			LockedBy:        nil,
			Confirmed:       false,
		},
	}
	err := unspentRepository.AddUnspents(ctx, unspents)
	if err != nil {
		t.Fatal(err)
	}

	u := unspentRepository.GetAllUnspents(ctx)

	assert.Equal(t, len(u), 8)
	assert.Equal(t, u[6].TxID == "6", true)
	assert.Equal(t, u[7].TxID == "7", true)

	//repeat insertion of same keys to check if there will be errors
	unspents = []domain.Unspent{
		{
			TxID:            "6",
			VOut:            1,
			Value:           0,
			AssetHash:       "",
			ValueCommitment: "",
			AssetCommitment: "",
			Nonce:           nil,
			RangeProof:      nil,
			SurjectionProof: nil,
			Address:         "a",
			Spent:           false,
			Locked:          false,
			ScriptPubKey:    nil,
			LockedBy:        nil,
			Confirmed:       false,
		},
		{
			TxID:            "7",
			VOut:            2,
			Value:           0,
			AssetHash:       "",
			ValueCommitment: "",
			AssetCommitment: "",
			Nonce:           nil,
			RangeProof:      nil,
			SurjectionProof: nil,
			Address:         "b",
			Spent:           false,
			Locked:          false,
			ScriptPubKey:    nil,
			LockedBy:        nil,
			Confirmed:       false,
		},
	}
	err = unspentRepository.AddUnspents(ctx, unspents)
	if err != nil {
		t.Fatal(err)
	}

	u = unspentRepository.GetAllUnspents(ctx)

	assert.Equal(t, len(u), 8)
	assert.Equal(t, u[6].TxID == "6", true)
	assert.Equal(t, u[7].TxID == "7", true)
}

func TestGetBalance(t *testing.T) {
	before()
	defer after()

	balance, err := unspentRepository.GetBalance(ctx, "a", "ah")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, balance, uint64(4))
}

func TestGetAvailableUnspents(t *testing.T) {
	before()
	defer after()

	unspents, err := unspentRepository.GetAvailableUnspents(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(unspents), 2)
}

func TestGetAvailableUnspentsForAddresses(t *testing.T) {
	before()
	defer after()

	addresses := []string{"a", "adr"}
	unspents, err := unspentRepository.GetAvailableUnspentsForAddresses(ctx, addresses)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(unspents), 2)
}

func TestGetUnlockedBalance(t *testing.T) {
	before()
	defer after()

	balance, err := unspentRepository.GetUnlockedBalance(ctx, "a", "ah")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, balance, uint64(4))
}

func TestLockUnlockUnspents(t *testing.T) {
	before()
	defer after()

	unpsentsKeys := []domain.UnspentKey{
		{
			TxID: "1",
			VOut: 1,
		},
		{
			TxID: "2",
			VOut: 1,
		},
	}

	err := unspentRepository.LockUnspents(ctx, unpsentsKeys, uuid.New())
	if err != nil {
		t.Fatal(err)
	}

	lockedCount := 0
	unspents := unspentRepository.GetAllUnspents(ctx)
	for _, v := range unspents {
		if v.IsLocked() == true {
			lockedCount++
		}
	}

	assert.Equal(t, lockedCount, 2)

	err = unspentRepository.UnlockUnspents(ctx, unpsentsKeys)
	if err != nil {
		t.Fatal(err)
	}

	lockedCount = 0
	unspents = unspentRepository.GetAllUnspents(ctx)
	for _, v := range unspents {
		if v.IsLocked() == true {
			lockedCount++
		}
	}

	assert.Equal(t, lockedCount, 0)
}
