package db_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
	"github.com/vulpemventures/go-elements/network"
)

var (
	mnemonic = []string{
		"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
		"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
		"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
	}
	encryptedMnemonic = "KfxUQ4NhpbeGb/rdsEkgewYZrAtMODrjvgXrOjBU9vRTNo5zk7F37GCKNskdYEJovjUG/STae7gutnUBBOIgXRnQIiDl7cn/8t9CYnfHVta+Pu2E2U39LTRcAsosKBam8F4GcsJ7miId62HkhhVnezl8MOeA1y7g6kwYadzjzOPXnToYyGyDPU+lO4HngKfGIJyNyhAim2yuIy4d54PrJ3fQYZCMfc6oHEVyCGxVePJ6ZMvO+gNhhO5w9eT1bv4DsnCbdb+heseZt7K8b0cHoj1rR2ky"
	passphrase        = "passphrase"
	net               = &network.Regtest
)

func TestVaultRepositoryImplementations(t *testing.T) {
	repositories, cancel := createVaultRepositories(t)
	t.Cleanup(cancel)

	domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)
	mockedEncrypter := newMockedEncrypter(mnemonic, encryptedMnemonic)
	domain.EncrypterManager = mockedEncrypter

	for i := range repositories {
		repo := repositories[i]

		t.Run(repo.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("testGetOrCreateVault", func(t *testing.T) {
				t.Parallel()

				testGetOrCreateVault(t, repo)
			})

			t.Run("testUpdateVault", func(t *testing.T) {
				t.Parallel()

				testUpdateVault(t, repo)
			})

			// TODO: uncomment - the following test demonstrate that in case of error,
			// any change to the db is rolled back. Currently, the transactional
			// component is not properly implemented for inmemory DbManager and the
			// test below would fail for te inmemory implementation.

			// t.Run("testUpdateVault_rollback", func(t *testing.T) {
			// 	testUpdateVaultRollback(t, repo)
			// })
		})
	}
}

func testGetOrCreateVault(t *testing.T, repo vaultRepository) {
	iNewVault, err := repo.write(func(ctx context.Context) (interface{}, error) {
		return repo.Repository.GetOrCreateVault(ctx, mnemonic, passphrase, net)
	})
	require.NoError(t, err)

	vault, ok := iNewVault.(*domain.Vault)
	require.True(t, ok)
	require.NotNil(t, vault)
	require.True(t, vault.IsInitialized())
}

func testUpdateVault(t *testing.T, repo vaultRepository) {
	iVault, err := repo.write(func(ctx context.Context) (interface{}, error) {
		_, err := repo.Repository.GetOrCreateVault(ctx, mnemonic, passphrase, net)
		if err != nil {
			return nil, err
		}
		if err := repo.Repository.UpdateVault(
			ctx,
			func(v *domain.Vault) (*domain.Vault, error) {
				return v, nil
			},
		); err != nil {
			return nil, err
		}
		return repo.Repository.GetOrCreateVault(ctx, nil, "", nil)
	})
	require.NoError(t, err)

	vault, ok := iVault.(*domain.Vault)
	require.True(t, ok)
	require.NotNil(t, vault)
}

func testUpdateVaultRollback(t *testing.T, repo vaultRepository) {
	expectedErr := errors.New("something went wrong")

	vault, err := repo.write(func(ctx context.Context) (interface{}, error) {
		_, err := repo.Repository.GetOrCreateVault(ctx, mnemonic, passphrase, net)
		if err != nil {
			return nil, err
		}
		if err := repo.Repository.UpdateVault(
			ctx,
			func(v *domain.Vault) (*domain.Vault, error) {
				return nil, expectedErr
			},
		); err != nil {
			return nil, err
		}
		return repo.Repository.GetOrCreateVault(ctx, nil, "", nil)
	})
	require.EqualError(t, err, expectedErr.Error())
	require.Nil(t, vault)

	vault, err = repo.read(func(ctx context.Context) (interface{}, error) {
		return repo.Repository.GetOrCreateVault(ctx, nil, "", nil)
	})
	require.Error(t, err)
	require.Nil(t, vault)
}

func createVaultRepositories(t *testing.T) ([]vaultRepository, func()) {
	datadir := "vaultdb"
	err := os.Mkdir(datadir, os.ModePerm)
	require.NoError(t, err)

	inmemoryDBManager := inmemory.NewDbManager()
	badgerDBManager, err := dbbadger.NewDbManager(datadir, nil)
	require.NoError(t, err)

	return []vaultRepository{
			{
				Name:       "badger",
				DBManager:  badgerDBManager,
				Repository: newBadgerVaultRepository(badgerDBManager),
			},
			{
				Name:       "inmemory",
				DBManager:  inmemoryDBManager,
				Repository: newInMemoryVaultRepository(inmemoryDBManager),
			},
		}, func() {
			os.RemoveAll(datadir)
		}
}

func newBadgerVaultRepository(dbmanager *dbbadger.DbManager) domain.VaultRepository {
	return dbbadger.NewVaultRepositoryImpl(dbmanager)
}

func newInMemoryVaultRepository(dbmanager *inmemory.DbManager) domain.VaultRepository {
	return inmemory.NewVaultRepositoryImpl(dbmanager)
}

type vaultRepository struct {
	Name       string
	DBManager  ports.DbManager
	Repository domain.VaultRepository
}

func (r vaultRepository) read(query func(context.Context) (interface{}, error)) (interface{}, error) {
	return r.DBManager.RunTransaction(context.Background(), true, query)
}

func (r vaultRepository) write(query func(context.Context) (interface{}, error)) (interface{}, error) {
	return r.DBManager.RunTransaction(context.Background(), false, query)
}

type simpleMnemonicStore struct {
	mnemonic []string
	lock     *sync.RWMutex
}

func newSimpleMnemonicStore(m []string) domain.MnemonicStore {
	return &simpleMnemonicStore{
		mnemonic: m,
		lock:     &sync.RWMutex{},
	}
}

func (s *simpleMnemonicStore) Set(mnemonic string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.mnemonic = strings.Split(mnemonic, " ")
}

func (s *simpleMnemonicStore) Unset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.mnemonic = nil
}

func (s *simpleMnemonicStore) IsSet() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.mnemonic) > 0
}

func (s *simpleMnemonicStore) Get() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.mnemonic
}

type mockEncrypter struct {
	mnemonic          []string
	encryptedMnemonic string
}

func newMockedEncrypter(mnemonic []string, encryptedMnemonic string) domain.Encrypter {
	return mockEncrypter{mnemonic, encryptedMnemonic}
}

func (m mockEncrypter) Encrypt(mnemonic, passphrase string) (string, error) {
	return m.encryptedMnemonic, nil
}

func (m mockEncrypter) Decrypt(encryptedMnemonic, passphrase string) (string, error) {
	return strings.Join(m.mnemonic, " "), nil
}
