package application

import (
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/transaction"
)

// SwapInfo is the data struct returned by ListSwap RPC.
type SwapInfo struct {
	Status           int32
	AmountP          uint64
	AssetP           string
	AmountR          uint64
	AssetR           string
	MarketFee        Fee
	RequestTimeUnix  uint64
	AcceptTimeUnix   uint64
	CompleteTimeUnix uint64
	ExpiryTimeUnix   uint64
}

// MarketInfo is the data struct returned by ListMarket RPC.
type MarketInfo struct {
	AccountIndex uint64
	Market       Market
	Fee          Fee
	Tradable     bool
	StrategyType int
	Price        domain.Prices
}

type Market struct {
	BaseAsset  string
	QuoteAsset string
}

type MarketWithFee struct {
	Market
	Fee
}

// Fee is the market fee percentage in basis point:
// 	- 0,01% -> 1 bp
//	- 1,00% -> 100 bp
//	- 99,99% -> 9999 bp
type Fee struct {
	BasisPoint int64
}

type MarketWithPrice struct {
	Market
	Price
}

type Price struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}

type PriceWithFee struct {
	Price  Price
	Fee    Fee
	Amount uint64
	Asset  string
}

type MarketStrategy struct {
	Market
	Strategy domain.StrategyType
}

type Balance struct {
	BaseAmount  int64
	QuoteAmount int64
}

type BalanceWithFee struct {
	Balance Balance
	Fee     Fee
}

type BalanceInfo struct {
	TotalBalance       uint64
	ConfirmedBalance   uint64
	UnconfirmedBalance uint64
}

type WithdrawMarketReq struct {
	Market
	BalanceToWithdraw Balance
	MillisatPerByte   int64
	Address           string
	Push              bool
}

type ReportMarketFee struct {
	CollectedFees              []FeeInfo
	TotalCollectedFeesPerAsset map[string]int64
}

type AddressAndBlindingKey struct {
	Address     string
	BlindingKey string
}

type FeeInfo struct {
	TradeID     string
	BasisPoint  int64
	Asset       string
	Amount      uint64
	MarketPrice decimal.Decimal
}

type TxOutpoint struct {
	Hash  string
	Index int
}

type UtxoInfoList struct {
	Unspents []UtxoInfo
	Spents   []UtxoInfo
	Locks    []UtxoInfo
}

type UtxoInfo struct {
	Outpoint *TxOutpoint
	Value    uint64
	Asset    string
}

type Unspents []domain.Unspent

func (u Unspents) ToUtxos() []explorer.Utxo {
	l := make([]explorer.Utxo, len(u), len(u))
	for i := range u {
		l[i] = u[i].ToUtxo()
	}
	return l
}

type UnblindedResult *transactionutil.UnblindedResult
type BlindingData wallet.BlindingData

type Blinder interface {
	UnblindOutput(txout *transaction.TxOutput, key []byte) (UnblindedResult, bool)
}

type FillProposalOpts struct {
	Mnemonic      []string
	SwapRequest   domain.SwapRequest
	MarketUtxos   []explorer.Utxo
	FeeUtxos      []explorer.Utxo
	MarketInfo    domain.AddressesInfo
	FeeInfo       domain.AddressesInfo
	OutputInfo    domain.AddressInfo
	ChangeInfo    domain.AddressInfo
	FeeChangeInfo domain.AddressInfo
	Network       *network.Network
}

type FillProposalResult struct {
	PsetBase64         string
	SelectedUnspents   []explorer.Utxo
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
}
type TradeHandler interface {
	FillProposal(FillProposalOpts) (*FillProposalResult, error)
}

var (
	BlinderManager Blinder
	TradeManager   TradeHandler
)

type blinderManager struct{}

func (b blinderManager) UnblindOutput(
	txout *transaction.TxOutput,
	key []byte,
) (UnblindedResult, bool) {
	return transactionutil.UnblindOutput(txout, key)
}

type tradeManager struct{}

func (t tradeManager) FillProposal(opts FillProposalOpts) (*FillProposalResult, error) {
	return fillProposal(opts)
}

func init() {
	BlinderManager = blinderManager{}
	TradeManager = tradeManager{}
}
