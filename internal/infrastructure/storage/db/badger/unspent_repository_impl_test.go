package dbbadger

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
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

	balance, err := unspentRepository.GetBalance(ctx, []string{"a"}, "ah")
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

func TestGetAllUnspentsForAddresses(t *testing.T) {
	before()
	defer after()

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
	before()
	defer after()

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

func TestSpendUnspents(t *testing.T) {
	before()
	defer after()

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
	before()
	defer after()

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
	before()
	defer after()

	balance, err := unspentRepository.GetUnlockedBalance(ctx, []string{"a"}, "ah")
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

func TestUnspentLockTTL(t *testing.T) {
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

	time.Sleep(3 * time.Second)

	lockedCount = 0
	unspents = unspentRepository.GetAllUnspents(ctx)
	for _, v := range unspents {
		if v.IsLocked() == true {
			lockedCount++
		}
	}
	assert.Equal(t, 0, lockedCount)
}

func TestConcurrentGetUnspentsAddUnspents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	os.Mkdir(testDbDir, os.ModePerm)
	dbManager, err := NewDbManager(testDbDir, nil)
	if err != nil {
		panic(err)
	}
	unspentRepository := NewUnspentRepositoryImpl(dbManager)
	defer func() {
		rec := recover()
		os.RemoveAll(testDbDir)
		if rec != nil {
			t.Fatal(rec)
		}
	}()

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
		tx := dbManager.NewUnspentsTransaction()
		ctx := context.WithValue(context.Background(), "utx", tx)
		if len(unspents) > 0 {
			oldUnspentKeys = make([]domain.UnspentKey, 0, len(unspents))
			for _, u := range unspents {
				oldUnspentKeys = append(oldUnspentKeys, u.Key())
			}
		}

		unspents = randUnspents()
		if err := repo.AddUnspents(ctx, unspents); err != nil {
			t.Log(err)
			tx.Discard()
			continue
		}
		if err := repo.SpendUnspents(ctx, oldUnspentKeys); err != nil {
			t.Log(err)
			tx.Discard()
			continue
		}
		if err := tx.Commit(); err != nil {
			panic(err)
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
		if timeElapsed := time.Since(now); timeElapsed.Seconds() >= 30 {
			break
		}

		tx := dbManager.NewUnspentsTransaction()
		ctx := context.WithValue(context.Background(), "utx", tx)
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
