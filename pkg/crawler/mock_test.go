package crawler_test

import (
	"github.com/stretchr/testify/mock"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

// Explorer
type mockExplorer struct {
	mock.Mock
}

func (m *mockExplorer) GetUnspents(
	addr string, blindKeys [][]byte,
) ([]explorer.Utxo, error) {
	args := m.Called(addr, blindKeys)

	var res []explorer.Utxo
	if a := args.Get(0); a != nil {
		res = a.([]explorer.Utxo)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) GetUnspentsForAddresses(
	addresses []string, blindKeys [][]byte,
) ([]explorer.Utxo, error) {
	args := m.Called(addresses, blindKeys)

	var res []explorer.Utxo
	if a := args.Get(0); a != nil {
		res = a.([]explorer.Utxo)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) GetUnspentStatus(
	hash string, index uint32,
) (explorer.UtxoStatus, error) {
	args := m.Called(hash, index)

	var res explorer.UtxoStatus
	if a := args.Get(0); a != nil {
		res = a.(explorer.UtxoStatus)
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

func (m *mockExplorer) GetTransactionStatus(txid string) (explorer.TransactionStatus, error) {
	args := m.Called(txid)

	if a := args.Get(0); a != nil {
		return a.(explorer.TransactionStatus), nil
	}
	return nil, args.Error(1)
}

func (m *mockExplorer) GetTransactionsForAddress(
	addr string, blindKey []byte,
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

func (m *mockExplorer) Faucet(addr string, amount float64, asset string) (string, error) {
	args := m.Called(addr, amount, asset)

	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	return res, args.Error(1)
}

func (m *mockExplorer) Mint(addr string, amount float64) (string, string, error) {
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

// UtxoStatus

type mockUtxoStatus struct {
	mock.Mock
}

func (m *mockUtxoStatus) Spent() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *mockUtxoStatus) Hash() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockUtxoStatus) Index() int {
	args := m.Called()
	return args.Get(0).(int)
}

// TransactionStatus
type mockTxStatus struct {
	mock.Mock
}

func (m *mockTxStatus) Confirmed() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *mockTxStatus) BlockHash() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockTxStatus) BlockHeight() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *mockTxStatus) BlockTime() int {
	args := m.Called()
	return args.Get(0).(int)
}

// Outpoint
type mockOutpoint struct {
	mock.Mock
}

func (m *mockOutpoint) Hash() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockOutpoint) Index() uint32 {
	args := m.Called()
	return uint32(args.Get(0).(int))
}
