package application_test

import (
	"strings"
	"sync"

	"github.com/stretchr/testify/mock"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/transaction"
)

// **** BlinderMager ****
type mockBlinderManager struct {
	mock.Mock
}

func (m *mockBlinderManager) UnblindOutput(
	txout *transaction.TxOutput,
	key []byte,
) (application.UnblindedResult, bool) {
	args := m.Called(txout, key)

	var res application.UnblindedResult
	if a := args.Get(0); a != nil {
		res = a.(application.UnblindedResult)
	}
	var res1 bool
	if a := args.Get(1); a != nil {
		res1 = a.(bool)
	}
	return res, res1
}

// **** Explorer ****

type mockExplorer struct {
	mock.Mock
}

func (m *mockExplorer) GetUnspents(
	addr string,
	blindKeys [][]byte,
) ([]explorer.Utxo, error) {
	args := m.Called(addr, blindKeys)

	var res []explorer.Utxo
	if a := args.Get(0); a != nil {
		res = a.([]explorer.Utxo)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) GetUnspentsForAddresses(
	addresses []string,
	blindKeys [][]byte,
) ([]explorer.Utxo, error) {
	args := m.Called(addresses, blindKeys)

	var res []explorer.Utxo
	if a := args.Get(0); a != nil {
		res = a.([]explorer.Utxo)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) GetTransaction(txid string) (explorer.Transaction, error) {
	args := m.Called(txid)

	var res explorer.Transaction
	if a := args.Get(0); a != nil {
		res = a.(explorer.Transaction)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) GetTransactionHex(txid string) (string, error) {
	args := m.Called(txid)

	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) IsTransactionConfirmed(txid string) (bool, error) {
	args := m.Called(txid)

	var res bool
	if a := args.Get(0); a != nil {
		res = a.(bool)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) GetTransactionStatus(txid string) (map[string]interface{}, error) {
	args := m.Called(txid)

	var res map[string]interface{}
	if a := args.Get(0); a != nil {
		res = a.(map[string]interface{})
	}
	return res, args.Error(1)
}

func (m *mockExplorer) GetTransactionsForAddress(
	addr string,
	blindKey []byte,
) ([]explorer.Transaction, error) {
	args := m.Called(addr, blindKey)

	var res []explorer.Transaction
	if a := args.Get(0); a != nil {
		res = a.([]explorer.Transaction)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) BroadcastTransaction(txhex string) (string, error) {
	args := m.Called(txhex)

	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) Faucet(addr string) (string, error) {
	args := m.Called(addr)

	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) Mint(addr string, amount int) (string, string, error) {
	args := m.Called(addr, amount)

	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	var res1 string
	if a := args.Get(1); a != nil {
		res1 = a.(string)
	}
	return res, res1, args.Error(2)
}

func (m *mockExplorer) GetBlockHeight() (int, error) {
	args := m.Called()

	var res int
	if a := args.Get(0); a != nil {
		res = a.(int)
	}
	return res, args.Error(1)
}

// **** Explorer's Transaction ****

type mockTransaction struct {
	address string
}

func (m *mockTransaction) Hash() string {
	return randomHex(32)
}

func (m *mockTransaction) Version() int {
	return 2
}

func (m *mockTransaction) Locktime() int {
	return 0
}
func (m *mockTransaction) Inputs() []*transaction.TxInput {
	inLen := randomIntInRange(1, 3)
	ins := make([]*transaction.TxInput, inLen, inLen)
	for i := 0; i < inLen; i++ {
		ins[i] = transaction.NewTxInput(randomBytes(32), 0)
	}
	return ins
}

func (m *mockTransaction) Outputs() []*transaction.TxOutput {
	outLen := randomIntInRange(2, 5)
	outs := make([]*transaction.TxOutput, outLen, outLen)
	script, _ := address.ToOutputScript(m.address)
	for i := 0; i < outLen; i++ {
		outs[i] = &transaction.TxOutput{
			Asset:           randomBytes(33),
			Value:           randomBytes(33),
			Script:          script,
			Nonce:           randomBytes(33),
			RangeProof:      randomBytes(100),
			SurjectionProof: randomBytes(100),
		}
	}
	return outs
}

func (m *mockTransaction) Size() int {
	return randomIntInRange(200, 500)
}

func (m *mockTransaction) Weight() int {
	return randomIntInRange(500, 1000)
}

func (m *mockTransaction) Confirmed() bool {
	return true
}

// **** MnemonicStore *****

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

// **** Encrypter ****

type mockCryptoHandler struct {
	encrypt func(mnemonic, passphrase string) (string, error)
	decrypt func(encryptedMnemonic, passphrase string) (string, error)
}

func (c mockCryptoHandler) Encrypt(mnemonic, passpharse string) (string, error) {
	return c.encrypt(mnemonic, passpharse)
}

func (c mockCryptoHandler) Decrypt(encryptedMnemonic, passpharse string) (string, error) {
	return c.decrypt(encryptedMnemonic, passpharse)
}
