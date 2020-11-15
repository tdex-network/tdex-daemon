package inmemory

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func TestAddUnspents(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

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
}

func TestGetBalance(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	balance, err := unspentRepository.GetBalance(ctx, []string{"a"}, "ah")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(4), balance)
}

func TestGetAvailableUnspents(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	unspents, err := unspentRepository.GetAvailableUnspents(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(unspents))
}

func TestGetAllUnspentsForAddresses(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	unspents, err := unspentRepository.GetAllUnspentsForAddresses(
		ctx,
		[]string{"a", "adr"},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 4, len(unspents))

}
func TestGetUnspentsForAddresses(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	unspents, err := unspentRepository.GetUnspentsForAddresses(
		ctx,
		[]string{"a", "adr"},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(unspents))
}

func TestGetAvailableUnspentsForAddresses(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	addresses := []string{"a", "adr"}
	unspents, err := unspentRepository.GetAvailableUnspentsForAddresses(ctx, addresses)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(unspents))
}

func TestGetUnspentForKey(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	unspent, err := unspentRepository.GetUnspentForKey(ctx, domain.UnspentKey{
		TxID: "1",
		VOut: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(2), unspent.Value)
}

func TestSpendUnspents(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	unspentKey := domain.UnspentKey{
		TxID: "1",
		VOut: 1,
	}
	unspent, err := unspentRepository.GetUnspentForKey(ctx, unspentKey)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, false, unspent.IsSpent())

	if err = unspentRepository.SpendUnspents(
		ctx,
		[]domain.UnspentKey{unspentKey},
	); err != nil {
		t.Fatal(err)
	}

	unspent, err = unspentRepository.GetUnspentForKey(ctx, unspentKey)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, unspent.IsSpent())
}

func TestConfirmUnspents(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	unspentKey := domain.UnspentKey{
		TxID: "2",
		VOut: 1,
	}
	unspent, err := unspentRepository.GetUnspentForKey(ctx, unspentKey)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, false, unspent.IsConfirmed())

	if err = unspentRepository.ConfirmUnspents(
		ctx,
		[]domain.UnspentKey{unspentKey},
	); err != nil {
		t.Fatal(err)
	}

	unspent, err = unspentRepository.GetUnspentForKey(ctx, unspentKey)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, unspent.IsConfirmed())
}

func TestGetUnlockedBalance(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	balance, err := unspentRepository.GetUnlockedBalance(ctx, []string{"a"}, "ah")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(4), balance)
}

func TestLockUnlockUnspents(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

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

func TestUnspentLockTTL(t *testing.T) {
	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	unspentsKeys := []domain.UnspentKey{
		{
			TxID: "1",
			VOut: 1,
		},
		{
			TxID: "2",
			VOut: 1,
		},
	}

	err := unspentRepository.LockUnspents(ctx, unspentsKeys, uuid.New())
	if err != nil {
		t.Fatal(err)
	}

	lockedCount := 0
	unspents := unspentRepository.GetAllUnspents(ctx)
	for _, v := range unspents {
		if v.Locked {
			lockedCount++
		}
	}
	assert.Equal(t, 2, lockedCount)

	time.Sleep(5 * time.Second)

	lockedCount = 0
	unspents = unspentRepository.GetAllUnspents(ctx)
	for _, v := range unspents {
		if v.Locked == true {
			lockedCount++
		}
	}
	assert.Equal(t, 0, lockedCount)
}

func TestConcurrentGetUnspentsAddUnspents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	dbManager := newMockDb()
	unspentRepository := NewUnspentRepositoryImpl(dbManager)

	go startWriter(t, dbManager, unspentRepository)
	time.Sleep(1 * time.Millisecond)
	startReader(t, dbManager, unspentRepository)
}

func startWriter(
	t *testing.T,
	dbManager *DbManager,
	repo domain.UnspentRepository,
) {
	var unspents []domain.Unspent
	var oldUnspentKeys []domain.UnspentKey
	for {
		if len(unspents) > 0 {
			oldUnspentKeys = make([]domain.UnspentKey, 0, len(unspents))
			for _, u := range unspents {
				oldUnspentKeys = append(oldUnspentKeys, u.Key())
			}
		}

		unspents = randUnspents()
		if err := repo.AddUnspents(ctx, unspents); err != nil {
			t.Log(err)
			continue
		}
		if err := repo.SpendUnspents(ctx, oldUnspentKeys); err != nil {
			t.Log(err)
			continue
		}

		t.Log("+++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		t.Logf("%d new unspents added", len(unspents))
		t.Logf("%d existing unspents marked as spent", len(oldUnspentKeys))
		t.Log("+++++++++++++++++++++++++++++++++++++++++++++++++++++++")

		time.Sleep(3 * time.Second)
	}
}

func startReader(
	t *testing.T,
	dbManager *DbManager,
	repo domain.UnspentRepository,
) {
	now := time.Now()

	for {
		if timeElapsed := time.Since(now); timeElapsed.Seconds() >= 20 {
			break
		}

		allUnspents := repo.GetAllUnspents(ctx)
		availableUnspents, err := repo.GetAvailableUnspents(ctx)
		if err != nil {
			panic(err)
		}
		t.Log("-------------------------------------------------------")
		t.Log("all unspents/spents:", len(allUnspents))
		t.Log("unspents:", len(availableUnspents))
		t.Log("spents:", math.Abs(float64(len(allUnspents)-len(availableUnspents))))
		t.Log("-------------------------------------------------------")

		time.Sleep(3 * time.Second)
	}
}
