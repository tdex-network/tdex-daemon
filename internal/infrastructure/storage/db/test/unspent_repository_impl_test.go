package db_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
)

func TestUnspentRepositoryImplementations(t *testing.T) {
	repositories := createUnspentRepositories(t)

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("testAddUnspents", func(t *testing.T) {
				t.Parallel()
				testAddUnspents(t, repo)
			})

			t.Run("testGetAvailableUnspents", func(t *testing.T) {
				t.Parallel()
				testGetAvailableUnspents(t, repo)
			})

			t.Run("testGetAllUnspentsForAddresses", func(t *testing.T) {
				t.Parallel()
				testGetAllUnspentsForAddresses(t, repo)
			})

			t.Run("testGetUnspentsForAddresses", func(t *testing.T) {
				t.Parallel()
				testGetUnspentsForAddresses(t, repo)
			})

			t.Run("testGetAvailableUnspentsForAddresses", func(t *testing.T) {
				t.Parallel()
				testGetAvailableUnspentsForAddresses(t, repo)
			})

			t.Run("testGetBalance", func(t *testing.T) {
				t.Parallel()
				testGetBalance(t, repo)
			})

			t.Run("testGetUnlockedBalance", func(t *testing.T) {
				t.Parallel()
				testGetUnlockedBalance(t, repo)
			})

			t.Run("testSpendUnspents", func(t *testing.T) {
				t.Parallel()
				testSpendUnspents(t, repo)
			})

			t.Run("testConfirmUnspents", func(t *testing.T) {
				t.Parallel()
				testConfirmUnspents(t, repo)
			})

			t.Run("testLockUnlockUnspents", func(t *testing.T) {
				t.Parallel()
				testLockUnlockUnspents(t, repo)
			})
		})
	}
}

func testAddUnspents(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfUnspents: 1, numOfSpents: 1})
	mockedUnspents := mockedData.unspents

	iNewUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		return repo.Repository.GetAllUnspents(ctx), nil
	})
	require.NoError(t, err)

	unspents, ok := iNewUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(unspents), len(mockedUnspents))
}

func testGetAvailableUnspents(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfUnspents: 2, numOfSpents: 1})
	mockedUnspents := mockedData.unspents

	iAvailableUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		return repo.Repository.GetAvailableUnspents(ctx)
	})
	require.NoError(t, err)

	unspents, ok := iAvailableUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(unspents), mockedData.expectedUnspents)
}

func testGetAllUnspentsForAddresses(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfUnspents: 1, numOfSpents: 1, numOfAddresses: 2})
	addresses := mockedData.addresses
	mockedUnspents := mockedData.unspents

	iAllUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		return repo.Repository.GetAllUnspentsForAddresses(ctx, addresses)
	})
	require.NoError(t, err)

	unspents, ok := iAllUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(unspents), mockedData.expectedUnspents)
}

func testGetUnspentsForAddresses(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfAddresses: 2, numOfUnspents: 1, numOfSpents: 1})
	addresses := mockedData.addresses
	mockedUnspents := mockedData.unspents
	expectedUnspents := mockedData.expectedUnspents

	iUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		return repo.Repository.GetUnspentsForAddresses(ctx, addresses)
	})
	require.NoError(t, err)

	unspents, ok := iUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(unspents), expectedUnspents)
}

func testGetAvailableUnspentsForAddresses(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfAddresses: 2, numOfUnspents: 2, numOfSpents: 1})
	addresses := mockedData.addresses
	mockedUnspents := mockedData.unspents

	iAvailableUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		return repo.Repository.GetAvailableUnspentsForAddresses(ctx, addresses)
	})
	require.NoError(t, err)

	unspents, ok := iAvailableUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(unspents), mockedData.expectedUnspents)
}

func testGetBalance(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfAddresses: 1, numOfUnspents: 1, numOfSpents: 1, numOfUnspentsWithAsset: 1, numOfSpentsWithAsset: 1})
	mockedUnspents := mockedData.unspents
	addresses := mockedData.addresses
	asset := mockedData.asset

	iBalance, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		return repo.Repository.GetBalance(ctx, addresses, asset)
	})
	require.NoError(t, err)

	balance, ok := iBalance.(uint64)
	require.True(t, ok)
	require.Greater(t, int(balance), 0)
}

func testGetUnlockedBalance(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfAddresses: 1, numOfUnspents: 1, numOfSpents: 1, numOfUnspentsWithAsset: 1, numOfSpentsWithAsset: 1})
	mockedUnspents := mockedData.unspents
	addresses := mockedData.addresses
	asset := mockedData.asset

	iBalance, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		return repo.Repository.GetUnlockedBalance(ctx, addresses, asset)
	})
	require.NoError(t, err)

	balance, ok := iBalance.(uint64)
	require.True(t, ok)
	require.Greater(t, int(balance), 0)
}

func testSpendUnspents(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfUnspents: 1})
	mockedUnspents := mockedData.unspents
	unspentKeys := mockedData.unspentKeys

	iUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		return repo.Repository.GetAllUnspents(ctx), nil
	})
	require.NoError(t, err)

	unspents, ok := iUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.NotNil(t, unspents)
	for _, key := range unspentKeys {
		for _, u := range unspents {
			if u.IsKeyEqual(key) {
				require.False(t, u.IsSpent())
				break
			}
		}
	}

	iSpentUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.SpendUnspents(ctx, unspentKeys); err != nil {
			return nil, err
		}

		return repo.Repository.GetAllUnspents(ctx), nil
	})
	require.NoError(t, err)

	unspents, ok = iSpentUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.NotNil(t, unspents)
	for _, key := range unspentKeys {
		for _, u := range unspents {
			if u.IsKeyEqual(key) {
				require.True(t, u.IsSpent())
				break
			}
		}
	}
}

func testConfirmUnspents(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfUnconfirmedUnspents: 1})
	mockedUnspents := mockedData.unspents
	unspentKeys := mockedData.unspentKeys

	iUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		return repo.Repository.GetAllUnspents(ctx), nil
	})
	require.NoError(t, err)

	unspents, ok := iUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.NotNil(t, unspents)
	for _, key := range unspentKeys {
		for _, u := range unspents {
			if u.IsKeyEqual(key) {
				require.False(t, u.IsConfirmed())
				break
			}
		}
	}

	iConfirmedUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.ConfirmUnspents(ctx, unspentKeys); err != nil {
			return nil, err
		}

		return repo.Repository.GetAllUnspents(ctx), nil
	})
	require.NoError(t, err)

	unspents, ok = iConfirmedUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.NotNil(t, unspents)
	for _, key := range unspentKeys {
		for _, u := range unspents {
			if u.IsKeyEqual(key) {
				require.True(t, u.IsConfirmed())
				break
			}
		}
	}
}

func testLockUnlockUnspents(t *testing.T, repo unspentRepository) {
	mockedData := mockUnspentTestData(opts{numOfUnspents: 1})
	mockedUnspents := mockedData.unspents
	unspentKeys := mockedData.unspentKeys
	mockedTradeID := uuid.New()

	iLockedUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.AddUnspents(ctx, mockedUnspents); err != nil {
			return nil, err
		}

		if _, err := repo.Repository.LockUnspents(ctx, unspentKeys, mockedTradeID); err != nil {
			return nil, err
		}

		return repo.Repository.GetAllUnspents(ctx), nil
	})
	require.NoError(t, err)

	unspents, ok := iLockedUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.NotNil(t, unspents)
	for _, key := range unspentKeys {
		for _, u := range unspents {
			if u.IsKeyEqual(key) {
				require.True(t, u.IsLocked())
				break
			}
		}
	}

	iUnlockedUnspents, err := repo.write(func(ctx context.Context) (interface{}, error) {
		if _, err := repo.Repository.UnlockUnspents(ctx, unspentKeys); err != nil {
			return nil, err
		}

		return repo.Repository.GetAllUnspents(ctx), nil
	})
	require.NoError(t, err)

	unspents, ok = iUnlockedUnspents.([]domain.Unspent)
	require.True(t, ok)
	require.NotNil(t, unspents)
	for _, key := range unspentKeys {
		for _, u := range unspents {
			if u.IsKeyEqual(key) {
				require.False(t, u.IsLocked())
				break
			}
		}
	}
}

func createUnspentRepositories(t *testing.T) []unspentRepository {
	inmemoryDBManager := inmemory.NewRepoManager()
	badgerDBManager, err := dbbadger.NewRepoManager("", nil)
	require.NoError(t, err)

	return []unspentRepository{
		{
			Name:       "badger",
			DBManager:  badgerDBManager,
			Repository: badgerDBManager.UnspentRepository(),
		},
		{
			Name:       "inmemory",
			DBManager:  inmemoryDBManager,
			Repository: inmemoryDBManager.UnspentRepository(),
		},
	}
}

type unspentRepository struct {
	Name       string
	DBManager  ports.RepoManager
	Repository domain.UnspentRepository
}

func (r unspentRepository) read(query func(context.Context) (interface{}, error)) (interface{}, error) {
	return r.DBManager.RunUnspentsTransaction(context.Background(), true, query)
}

func (r unspentRepository) write(query func(context.Context) (interface{}, error)) (interface{}, error) {
	return r.DBManager.RunUnspentsTransaction(context.Background(), false, query)
}

type opts struct {
	numOfAddresses           int
	numOfUnspents            int
	numOfSpents              int
	numOfSpentsWithAsset     int
	numOfUnspentsWithAsset   int
	numOfUnconfirmedUnspents int
}

type mock struct {
	addresses        []string
	unspents         []domain.Unspent
	unspentKeys      []domain.UnspentKey
	asset            string
	expectedUnspents int
	expectedSpents   int
}

func mockUnspentTestData(o opts) mock {
	var asset string
	var assetPtr *string
	if o.numOfSpentsWithAsset > 0 || o.numOfUnspentsWithAsset > 0 {
		assetPtr = mockAsset()
		asset = *assetPtr
	}

	if o.numOfAddresses > 0 {
		addresses := newTestAddresses(o.numOfAddresses)
		unspents := make([]domain.Unspent, 0)
		unspentKeys := make([]domain.UnspentKey, 0)

		for _, addr := range addresses {
			u, uk := mockUnspents(
				&addr,
				assetPtr,
				o.numOfUnspents, o.numOfSpents,
				o.numOfUnspentsWithAsset, o.numOfSpentsWithAsset,
				o.numOfUnconfirmedUnspents,
			)
			unspents = append(unspents, u...)
			unspentKeys = append(unspentKeys, uk...)
		}

		expectedUnspents := o.numOfAddresses * (o.numOfUnspents + o.numOfUnspentsWithAsset + o.numOfUnconfirmedUnspents)
		expectedSpents := o.numOfAddresses * (o.numOfSpents + o.numOfSpentsWithAsset + o.numOfUnconfirmedUnspents)

		return mock{addresses, unspents, unspentKeys, asset, expectedUnspents, expectedSpents}
	}

	unspents, unspentKeys := mockUnspents(
		nil,
		assetPtr,
		o.numOfUnspents, o.numOfSpents,
		o.numOfUnspentsWithAsset, o.numOfSpentsWithAsset,
		o.numOfUnconfirmedUnspents,
	)
	expectedUnspents := o.numOfUnspents + o.numOfUnspentsWithAsset + o.numOfUnconfirmedUnspents
	expectedSpents := o.numOfSpents + o.numOfSpentsWithAsset + o.numOfUnconfirmedUnspents
	return mock{nil, unspents, unspentKeys, asset, expectedUnspents, expectedSpents}
}

func mockUnspents(
	addr *string,
	asset *string,
	numOfUnspents, numOfSpents,
	numOfUnspentsWithAsset, numOfSpentsWithAsset,
	numOfUnconfirmedUnspents int,
) ([]domain.Unspent, []domain.UnspentKey) {
	unspents := make([]domain.Unspent, numOfUnspents, numOfUnspents)
	spents := make([]domain.Unspent, numOfSpents, numOfSpents)
	unspentsWithAsset := make([]domain.Unspent, numOfUnspentsWithAsset, numOfUnspentsWithAsset)
	spentsWithAsset := make([]domain.Unspent, numOfSpentsWithAsset, numOfSpentsWithAsset)
	unconfirmedUnspents := make([]domain.Unspent, numOfUnconfirmedUnspents, numOfUnconfirmedUnspents)

	spent := true
	confirmed := true
	for i := 0; i < numOfUnspents; i++ {
		unspents[i] = mockUnspent(addr, nil, !spent, confirmed)
	}
	for i := 0; i < numOfSpents; i++ {
		spents[i] = mockUnspent(addr, nil, spent, confirmed)
	}
	for i := 0; i < numOfUnspentsWithAsset; i++ {
		unspentsWithAsset[i] = mockUnspent(addr, asset, !spent, confirmed)
	}
	for i := 0; i < numOfSpentsWithAsset; i++ {
		spentsWithAsset[i] = mockUnspent(addr, asset, spent, confirmed)
	}
	for i := 0; i < numOfUnconfirmedUnspents; i++ {
		unconfirmedUnspents[i] = mockUnspent(addr, asset, !spent, !confirmed)
	}

	mockedUnspents := append(unspents, spents...)
	mockedUnspents = append(mockedUnspents, unspentsWithAsset...)
	mockedUnspents = append(mockedUnspents, spentsWithAsset...)
	mockedUnspents = append(mockedUnspents, unconfirmedUnspents...)

	unspentKeys := make([]domain.UnspentKey, len(mockedUnspents), len(mockedUnspents))
	for i, u := range mockedUnspents {
		unspentKeys[i] = u.Key()
	}

	return mockedUnspents, unspentKeys
}

func newTestAddresses(len int) []string {
	addresses := make([]string, len, len)
	for i := range addresses {
		addresses[i] = randomAddress()
	}
	return addresses
}

func mockAsset() *string {
	asset := randomString(32)
	return &asset
}

func mockUnspent(addr, asset *string, spent, confirmed bool) domain.Unspent {
	mockedAsset := randomString(32)
	if asset != nil {
		mockedAsset = *asset
	}
	var mockedAddress string
	if addr != nil {
		mockedAddress = *addr
	}

	return domain.Unspent{
		TxID:            randomString(32),
		VOut:            uint32(randomIntInRange(0, 15)),
		Value:           uint64(randomIntInRange(1, 100000000)),
		AssetHash:       mockedAsset,
		ValueCommitment: randomString(33),
		AssetCommitment: randomString(33),
		ValueBlinder:    randomBytes(32),
		AssetBlinder:    randomBytes(32),
		ScriptPubKey:    make([]byte, 20),
		Nonce:           make([]byte, 33),
		RangeProof:      make([]byte, 1),
		SurjectionProof: make([]byte, 1),
		Address:         mockedAddress,
		Spent:           spent,
		Confirmed:       confirmed,
	}
}

func randomAddress() string {
	key, _ := btcec.NewPrivateKey(btcec.S256())
	blindkey, _ := btcec.NewPrivateKey(btcec.S256())
	p := payment.FromPublicKey(key.PubKey(), &network.Regtest, blindkey.PubKey())
	addr, _ := p.WitnessPubKeyHash()
	return addr
}

func randomIntInRange(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(int(n.Int64())) + min
}

func randomString(len int) string {
	return hex.EncodeToString(randomBytes(32))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	rand.Read(b)
	return b
}
