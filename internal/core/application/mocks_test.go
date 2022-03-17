package application

import (
	"strings"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/transaction"
)

// **** BlinderMager ****
type mockBlinderManager struct {
	mock.Mock
}

func (m *mockBlinderManager) UnblindOutput(
	txout *transaction.TxOutput,
	key []byte,
) (UnblindedResult, bool) {
	args := m.Called(txout, key)

	var res UnblindedResult
	if a := args.Get(0); a != nil {
		res = a.(UnblindedResult)
	}
	var res1 bool
	if a := args.Get(1); a != nil {
		res1 = a.(bool)
	}
	return res, res1
}

// **** TradeManager ****

type mockTradeManager struct {
	mock.Mock
	counter int
	lock    *sync.Mutex
}

func newMockedTradeManager() *mockTradeManager {
	return &mockTradeManager{
		lock: &sync.Mutex{},
	}
}

func (m *mockTradeManager) FillProposal(
	opts FillProposalOpts,
) (*FillProposalResult, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.counter++
	args := m.Called(opts)

	var res *FillProposalResult
	if a := args.Get(0); a != nil {
		res = a.(*FillProposalResult)
	}
	return res, args.Error(1)
}

// **** TransactionManager ****

type mockTransactionManager struct {
	mock.Mock
}

func (m *mockTransactionManager) ExtractUnspents(
	txhex string,
	infoByScript map[string]domain.AddressInfo,
	net *network.Network,
) ([]domain.Unspent, []domain.UnspentKey, error) {
	args := m.Called(txhex, infoByScript, net)
	var res []domain.Unspent
	if a := args.Get(0); a != nil {
		res = a.([]domain.Unspent)
	}
	var res1 []domain.UnspentKey
	if a := args.Get(1); a != nil {
		res1 = a.([]domain.UnspentKey)
	}
	return res, res1, args.Error(2)
}

func (m *mockTransactionManager) ExtractBlindingData(
	psetBase64 string,
	inBlindingKeys, outBlindingData map[string][]byte,
) (map[int]BlindingData, map[int]BlindingData, error) {
	args := m.Called(psetBase64, inBlindingKeys, outBlindingData)
	var res map[int]BlindingData
	if a := args.Get(0); a != nil {
		res = a.(map[int]BlindingData)
	}
	var res1 map[int]BlindingData
	if a := args.Get(1); a != nil {
		res1 = a.(map[int]BlindingData)
	}
	return res, res1, args.Error(2)
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

func (m *mockExplorer) PollGetKnownTransaction(
	txid string, interval time.Duration,
) (explorer.Transaction, error) {
	args := m.Called(txid, interval)

	var res explorer.Transaction
	if a := args.Get(0); a != nil {
		res = a.(explorer.Transaction)
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

// **** PsetParser ****

type mockPsetParser struct {
	mock.Mock
}

func (m *mockPsetParser) GetTxID(psetBase64 string) (string, error) {
	args := m.Called(psetBase64)
	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	return res, args.Error(1)
}

func (m *mockPsetParser) GetTxHex(psetBase64 string) (string, error) {
	args := m.Called(psetBase64)
	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	return res, args.Error(1)
}

// **** SwapParser ****

type mockSwapParser struct {
	mock.Mock
}

func (m *mockSwapParser) SerializeRequest(req domain.SwapRequest) ([]byte, *domain.SwapError) {
	args := m.Called(req)

	var res []byte
	if a := args.Get(0); a != nil {
		res = a.([]byte)
	}

	var err *domain.SwapError
	if a := args.Get(1); a != nil {
		err = a.(*domain.SwapError)
	}
	return res, err
}

func (m *mockSwapParser) SerializeAccept(acc domain.AcceptArgs) (string, []byte, *domain.SwapError) {
	args := m.Called(acc)

	var sres string
	if a := args.Get(0); a != nil {
		sres = a.(string)
	}

	var bres []byte
	if a := args.Get(1); a != nil {
		bres = a.([]byte)
	}

	var err *domain.SwapError
	if args.Get(2) != nil {
		err = args.Get(2).(*domain.SwapError)
	}

	return sres, bres, err
}

func (m *mockSwapParser) SerializeComplete(accMsg []byte, tx string) (string, []byte, *domain.SwapError) {
	args := m.Called(accMsg, tx)

	var sres string
	if a := args.Get(0); a != nil {
		sres = a.(string)
	}

	var bres []byte
	if a := args.Get(1); a != nil {
		bres = a.([]byte)
	}

	var err *domain.SwapError
	if args.Get(2) != nil {
		err = args.Get(2).(*domain.SwapError)
	}

	return sres, bres, err
}

func (m *mockSwapParser) SerializeFail(id string, errCode int, errMsg string) (string, []byte) {
	args := m.Called(id, errCode, errMsg)

	var sres string
	if a := args.Get(0); a != nil {
		sres = a.(string)
	}

	var bres []byte
	if a := args.Get(1); a != nil {
		bres = a.([]byte)
	}

	return sres, bres
}

func (m *mockSwapParser) DeserializeRequest(msg []byte) (domain.SwapRequest, error) {
	args := m.Called(msg)
	var res domain.SwapRequest
	if a := args.Get(0); a != nil {
		res = a.(domain.SwapRequest)
	}

	return res, args.Error(1)
}

func (m *mockSwapParser) DeserializeAccept(msg []byte) (domain.SwapAccept, error) {
	args := m.Called(msg)
	var res domain.SwapAccept
	if a := args.Get(0); a != nil {
		res = a.(domain.SwapAccept)
	}
	return res, args.Error(1)
}

func (m *mockSwapParser) DeserializeComplete(msg []byte) (domain.SwapComplete, error) {
	args := m.Called(msg)
	var res domain.SwapComplete
	if a := args.Get(0); a != nil {
		res = a.(domain.SwapComplete)
	}
	return res, args.Error(1)
}

func (m *mockSwapParser) DeserializeFail(msg []byte) (domain.SwapFail, error) {
	args := m.Called(msg)
	var res domain.SwapFail
	if a := args.Get(0); a != nil {
		res = a.(domain.SwapFail)
	}
	return res, args.Error(1)
}

// trandaction status

type mockTxStatus map[string]interface{}

func (m mockTxStatus) Confirmed() bool {
	return m["confirmed"].(bool)
}
func (m mockTxStatus) BlockHash() string {
	return m["block_hash"].(string)
}
func (m mockTxStatus) BlockTime() int {
	return int(m["block_time"].(float64))
}
func (m mockTxStatus) BlockHeight() int {
	return int(m["block_height"].(float64))
}

// **** SwapRequest ****

type mockSwapRequest struct {
	id string
}

func newMockedSwapRequest() *mockSwapRequest {
	return &mockSwapRequest{randomId()}
}

func (m *mockSwapRequest) GetId() string {
	return m.id
}

func (m *mockSwapRequest) GetAssetP() string {
	return randomHex(32)
}

func (m *mockSwapRequest) GetAmountP() uint64 {
	return randomValue()
}

func (m *mockSwapRequest) GetAssetR() string {
	return randomHex(32)
}

func (m *mockSwapRequest) GetAmountR() uint64 {
	return randomValue()
}

func (m *mockSwapRequest) GetTransaction() string {
	return randomBase64()
}

func (m *mockSwapRequest) GetInputBlindingKey() map[string][]byte {
	return nil
}

func (m *mockSwapRequest) GetOutputBlindingKey() map[string][]byte {
	return nil
}

// **** SwapAccept ****

type mockSwapAccept struct {
	id string
}

func newMockedSwapAccept() *mockSwapAccept {
	return &mockSwapAccept{randomId()}
}

func (m *mockSwapAccept) GetId() string {
	return m.id
}

func (m *mockSwapAccept) GetRequestId() string {
	return randomId()
}

func (m *mockSwapAccept) GetTransaction() string {
	return randomBase64()
}

func (m *mockSwapAccept) GetInputBlindingKey() map[string][]byte {
	return nil
}

func (m *mockSwapAccept) GetOutputBlindingKey() map[string][]byte {
	return nil
}

// **** SwapComplete ****

type mockSwapComplete struct {
	id string
}

func newMockedSwapComplete() *mockSwapComplete {
	return &mockSwapComplete{randomId()}
}

func (m *mockSwapComplete) GetId() string {
	return m.id
}

func (m *mockSwapComplete) GetAcceptId() string {
	return randomId()
}

func (m *mockSwapComplete) GetTransaction() string {
	return randomBase64()
}

// **** SwapFail ****

type mockSwapFail struct {
	id string
}

func newMockedSwapFail() *mockSwapFail {
	return &mockSwapFail{randomId()}
}

func (m *mockSwapFail) GetId() string {
	return randomId()
}

func (m *mockSwapFail) GetMessageId() string {
	return randomId()
}

func (m *mockSwapFail) GetFailureCode() uint32 {
	return 1
}

func (m *mockSwapFail) GetFailureMessage() string {
	return "mocked error"
}
