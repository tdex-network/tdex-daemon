package domain_test

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/vulpemventures/go-elements/network"
)

func TestIsZero(t *testing.T) {
	v := newTestVaultEmpty()
	require.True(t, v.IsZero())

	v = newTestVaultLocked()
	require.False(t, v.IsZero())
}

func TestLockUnlock(t *testing.T) {
	v := newTestVaultLocked()
	domain.EncrypterManager = mockedCryptoHandler{
		encrypt: func(_, _ string) (string, error) {
			return v.EncryptedMnemonic, nil
		},
		decrypt: func(_, _ string) (string, error) {
			return "leave dice fine decrease dune ribbon ocean earn " +
				"lunar account silver admit cheap fringe disorder trade " +
				"because trade steak clock grace video jacket equal", nil
		},
	}
	domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)

	require.True(t, v.IsLocked())

	mnemonic, err := v.GetMnemonicSafe()
	require.Error(t, err)
	require.Nil(t, mnemonic)

	err = v.Unlock("pass")
	require.NoError(t, err)
	require.False(t, v.IsLocked())

	mnemonic, err = v.GetMnemonicSafe()
	require.NoError(t, err)
	require.Len(t, mnemonic, 24)

	v.Lock()
	require.NoError(t, err)
	require.True(t, v.IsLocked())
}

func TestFailingUnlock(t *testing.T) {
	v := newTestVaultLocked()

	expectedErr := errors.New("something went wrong")
	domain.EncrypterManager = mockedCryptoHandler{
		decrypt: func(_, _ string) (string, error) {
			return "", expectedErr
		},
	}
	domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)

	require.True(t, v.IsLocked())
	err := v.Unlock("")
	require.EqualError(t, err, expectedErr.Error())
	require.True(t, v.IsLocked())
}

func TestChangePasshprase(t *testing.T) {
	v := newTestVaultLocked()
	domain.EncrypterManager = mockedCryptoHandler{
		encrypt: func(_, _ string) (string, error) {
			return v.EncryptedMnemonic, nil
		},
		decrypt: func(_, _ string) (string, error) {
			return "leave dice fine decrease dune ribbon ocean earn " +
				"lunar account silver admit cheap fringe disorder trade " +
				"because trade steak clock grace video jacket equal", nil
		},
	}
	domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)

	passphrase := "pass"
	newPassphrase := "newpass"
	expectedPassphraseHash := btcutil.Hash160([]byte(newPassphrase))

	err := v.ChangePassphrase(passphrase, newPassphrase)
	require.NoError(t, err)
	require.Equal(t, expectedPassphraseHash, v.PassphraseHash)
}

func TestFailingChangePassphrase(t *testing.T) {
	passphrase := "pass"
	newPassphrase := "newpass"
	expectedErr := errors.New("something went wrong")

	t.Run("failing_decrypt", func(t *testing.T) {
		v := newTestVaultLocked()
		domain.EncrypterManager = mockedCryptoHandler{
			decrypt: func(_, _ string) (string, error) {
				return "", expectedErr
			},
		}
		domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)

		err := v.ChangePassphrase(passphrase, newPassphrase)
		require.EqualError(t, err, expectedErr.Error())
	})

	t.Run("failing_encrypt", func(t *testing.T) {
		v := newTestVaultLocked()
		domain.EncrypterManager = mockedCryptoHandler{
			encrypt: func(_, _ string) (string, error) {
				return "", expectedErr
			},
			decrypt: func(_, _ string) (string, error) {
				return "leave dice fine decrease dune ribbon ocean earn " +
					"lunar account silver admit cheap fringe disorder trade " +
					"because trade steak clock grace video jacket equal", nil
			},
		}

		err := v.ChangePassphrase(passphrase, newPassphrase)
		require.EqualError(t, err, expectedErr.Error())
	})
}

func TestAccountByIndex(t *testing.T) {
	v := newTestVaultLocked()
	domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)
	accountIndex := 0

	v.InitAccount(accountIndex)

	account, err := v.AccountByIndex(accountIndex)
	require.NoError(t, err)
	require.NotNil(t, account)
}

func TestFailingAccountByIndex(t *testing.T) {
	v := newTestVaultLocked()
	domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)
	accountIndex := 1

	account, err := v.AccountByIndex(accountIndex)
	require.EqualError(t, err, domain.ErrVaultAccountNotFound.Error())
	require.Nil(t, account)
}

func TestDeriveAddresses(t *testing.T) {
	v := newTestVaultLocked()
	domain.MnemonicStoreManager = newSimpleMnemonicStore([]string{
		"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
		"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
		"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
	})
	accountIndex := 2
	expectedExternalAddresses := 2
	expectedInternalAddresses := 3

	for i := 0; i < expectedExternalAddresses; i++ {
		info, err := v.DeriveNextExternalAddressForAccount(accountIndex)
		require.NoError(t, err)
		require.NotNil(t, info)
		require.GreaterOrEqual(t, info.AccountIndex, 0)
		require.NotEmpty(t, info.Address)
		require.Len(t, info.BlindingKey, 32)
		require.NotEmpty(t, info.DerivationPath)
		require.NotEmpty(t, info.Script)
	}

	for i := 0; i < expectedInternalAddresses; i++ {
		info, err := v.DeriveNextInternalAddressForAccount(accountIndex)
		require.NoError(t, err)
		require.NotNil(t, info)
		require.GreaterOrEqual(t, info.AccountIndex, 0)
		require.NotEmpty(t, info.Address)
		require.Len(t, info.BlindingKey, 32)
		require.NotEmpty(t, info.DerivationPath)
		require.NotEmpty(t, info.Script)
	}

	account, err := v.AccountByIndex(accountIndex)
	require.NoError(t, err)
	require.Equal(t, expectedExternalAddresses, account.LastExternalIndex)
	require.Equal(t, expectedInternalAddresses, account.LastInternalIndex)
}

func TestFailingDeriveAddresses(t *testing.T) {
	accountIndex := 3

	t.Run("failing_derive_external_address", func(t *testing.T) {
		v := newTestVaultLocked()
		domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)

		info, err := v.DeriveNextExternalAddressForAccount(accountIndex)
		require.EqualError(t, err, domain.ErrVaultMustBeUnlocked.Error())
		require.Nil(t, info)
	})

	t.Run("failing_derive_internal_address", func(t *testing.T) {
		v := newTestVaultLocked()
		domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)

		info, err := v.DeriveNextInternalAddressForAccount(accountIndex)
		require.EqualError(t, err, domain.ErrVaultMustBeUnlocked.Error())
		require.Nil(t, info)
	})
}

func TestAccountByAddress(t *testing.T) {
	v := newTestVaultLocked()
	domain.MnemonicStoreManager = newSimpleMnemonicStore([]string{
		"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
		"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
		"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
	})
	accountIndex := 4

	extInfo, err := v.DeriveNextExternalAddressForAccount(accountIndex)
	require.NoError(t, err)

	account, index, err := v.AccountByAddress(extInfo.Address)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, accountIndex, index)

	inInfo, err := v.DeriveNextInternalAddressForAccount(accountIndex)
	require.NoError(t, err)
	require.NoError(t, err)
	require.NotNil(t, account)

	account, index, err = v.AccountByAddress(inInfo.Address)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, accountIndex, index)
}

func TestFailingAccountByAddress(t *testing.T) {
	v := newTestVaultLocked()
	addr := "el1qq22f83p6asdy7jsp4tuke0d9emvxhcenqee5umsn88fsn8gggzlrx0md4hp38rnwcnu9lusmzhmktlt3h5q0gecfpfvx6uac2"

	account, index, err := v.AccountByAddress(addr)
	require.EqualError(t, err, domain.ErrVaultAccountNotFound.Error())
	require.Nil(t, account)
	require.Equal(t, -1, index)
}

func TestAllDerivedAddressesInfoForAccount(t *testing.T) {
	v := newTestVaultLocked()
	domain.MnemonicStoreManager = newSimpleMnemonicStore([]string{
		"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
		"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
		"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
	})
	accountIndex := 5

	extInfo, err := v.DeriveNextExternalAddressForAccount(accountIndex)
	require.NoError(t, err)
	inInfo, err := v.DeriveNextInternalAddressForAccount(accountIndex)
	require.NoError(t, err)

	allInfo, err := v.AllDerivedAddressesInfoForAccount(accountIndex)
	require.NoError(t, err)
	require.Len(t, allInfo, 2)

	addresses, blindkeys := allInfo.AddressesAndKeys()
	require.Len(t, addresses, 2)
	require.Len(t, blindkeys, 2)
	require.Equal(t, extInfo.Address, addresses[0])
	require.Equal(t, extInfo.BlindingKey, blindkeys[0])
	require.Equal(t, inInfo.Address, addresses[1])
	require.Equal(t, inInfo.BlindingKey, blindkeys[1])
}

func TestFailingAllDerivedAddressesInfoForAccount(t *testing.T) {
	accountIndex := 6

	t.Run("failing_because_locked", func(t *testing.T) {
		v := newTestVaultLocked()
		domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)

		info, err := v.AllDerivedAddressesInfoForAccount(accountIndex)
		require.EqualError(t, err, domain.ErrVaultMustBeUnlocked.Error())
		require.Nil(t, info)
	})

	t.Run("failing_because_account_not_found", func(t *testing.T) {
		v := newTestVaultLocked()
		domain.MnemonicStoreManager = newSimpleMnemonicStore([]string{
			"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
			"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
			"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
		})

		info, err := v.AllDerivedAddressesInfoForAccount(accountIndex)
		require.EqualError(t, err, domain.ErrVaultAccountNotFound.Error())
		require.Nil(t, info)
	})
}

func TestAllDerivedExternalAddressesInfoForAccount(t *testing.T) {
	v := newTestVaultLocked()
	domain.MnemonicStoreManager = newSimpleMnemonicStore([]string{
		"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
		"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
		"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
	})
	accountIndex := 7

	info, err := v.DeriveNextExternalAddressForAccount(accountIndex)
	require.NoError(t, err)
	_, err = v.DeriveNextInternalAddressForAccount(accountIndex)
	require.NoError(t, err)

	allInfo, err := v.AllDerivedExternalAddressesInfoForAccount(accountIndex)
	require.NoError(t, err)
	require.Len(t, allInfo, 1)
	addresses := allInfo.Addresses()
	require.Len(t, addresses, 1)
	require.Equal(t, info.Address, addresses[0])
}

func TestFailingAllDerivedExternalAddressesInfoForAccount(t *testing.T) {
	accountIndex := 8

	t.Run("failing_because_locked", func(t *testing.T) {
		v := newTestVaultLocked()
		domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)

		info, err := v.AllDerivedExternalAddressesInfoForAccount(accountIndex)
		require.EqualError(t, err, domain.ErrVaultMustBeUnlocked.Error())
		require.Nil(t, info)
	})

	t.Run("failing_because_account_not_found", func(t *testing.T) {
		v := newTestVaultLocked()
		domain.MnemonicStoreManager = newSimpleMnemonicStore([]string{
			"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
			"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
			"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
		})

		info, err := v.AllDerivedExternalAddressesInfoForAccount(accountIndex)
		require.EqualError(t, err, domain.ErrVaultAccountNotFound.Error())
		require.Nil(t, info)
	})
}

func newTestVaultEmpty() *domain.Vault {
	return &domain.Vault{}
}

func newTestVaultLocked() *domain.Vault {
	return &domain.Vault{
		Accounts:               make(map[int]*domain.Account),
		AccountAndKeyByAddress: make(map[string]domain.AccountAndKey),
		PassphraseHash:         btcutil.Hash160([]byte("pass")),
		EncryptedMnemonic:      "dVoBFte1oeRkPl8Vf8DzBP3PRnzPA3fxtyvDHXFGYAS9MP8V2Sc9nHcQW4PrMkQNnf2uGrDg81dFgBrwqv1n3frXxRBKhp83fSsTm4xqj8+jdwTI3nouFmi1W/O4UqpHdQ62EYoabJQtKpptWO11TFJzw8WF02pfS6git8YjLR4xrnfp2LkOEjSU9CI82ZasF46WZFKcpeUJTAsxU/03ONpAdwwEsC96f1KAvh8tqaO0yLDOcmPf8a5B82jefgncCRrt32kCpbpIE4YiCFrqqdUHXKH+",
		Network:                &network.Regtest,
	}
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

type mockedCryptoHandler struct {
	encrypt func(mnemonic, passphrase string) (string, error)
	decrypt func(encryptedMnemonic, passphrase string) (string, error)
}

func (c mockedCryptoHandler) Encrypt(mnemonic, passpharse string) (string, error) {
	return c.encrypt(mnemonic, passpharse)
}

func (c mockedCryptoHandler) Decrypt(encryptedMnemonic, passpharse string) (string, error) {
	return c.decrypt(encryptedMnemonic, passpharse)
}
