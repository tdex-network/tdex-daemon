package dbbadger

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, 8, len(u))
	assert.Equal(t, true, u[6].TxID == "6")
	assert.Equal(t, true, u[7].TxID == "7")

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
	assert.Equal(t, true, u[6].TxID == "6")
	assert.Equal(t, true, u[7].TxID == "7")
}

func TestGetBalance(t *testing.T) {
	before()
	defer after()

	balance, err := unspentRepository.GetBalance(ctx, "a", "ah")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(4), balance)
}

func TestGetAvailableUnspents(t *testing.T) {
	before()
	defer after()

	unspents, err := unspentRepository.GetAvailableUnspents(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(unspents))
}

func TestGetUnspentsForAddresses(t *testing.T) {
	before()
	defer after()

	unspents, err := unspentRepository.GetUnspentsForAddresses(
		ctx,
		[]string{"a", "adr"},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 4, len(unspents))
}

func TestGetAvailableUnspentsForAddresses(t *testing.T) {
	before()
	defer after()

	addresses := []string{"a", "adr"}
	unspents, err := unspentRepository.GetAvailableUnspentsForAddresses(ctx, addresses)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(unspents))
}

func TestGetUnspentForKey(t *testing.T) {
	before()
	defer after()

	unspent, err := unspentRepository.GetUnspentForKey(ctx, domain.UnspentKey{
		TxID: "1",
		VOut: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(2), unspent.Value)
}

func TestUpdateUnspent(t *testing.T) {
	before()
	defer after()

	unspent, err := unspentRepository.GetUnspentForKey(ctx, domain.UnspentKey{
		TxID: "1",
		VOut: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(2), unspent.Value)

	err = unspentRepository.UpdateUnspent(
		ctx,
		domain.UnspentKey{
			TxID: "1",
			VOut: 1,
		},
		func(m *domain.Unspent) (*domain.Unspent, error) {
			m.Value = 444
			return m, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	unspent, err = unspentRepository.GetUnspentForKey(ctx, domain.UnspentKey{
		TxID: "1",
		VOut: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(444), unspent.Value)
}

func TestGetUnlockedBalance(t *testing.T) {
	before()
	defer after()

	balance, err := unspentRepository.GetUnlockedBalance(ctx, "a", "ah")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(4), balance)
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

	assert.Equal(t, 2, lockedCount)

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

	assert.Equal(t, 0, lockedCount)
}
