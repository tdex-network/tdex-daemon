package application

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/transaction"
)

// HDWalletInfo contains info about the internal wallet of the daemon and its
// sub-accounts.
type HDWalletInfo struct {
	RootPath          string
	MasterBlindingKey string
	Accounts          []AccountInfo
}

type HDWalletStatus struct {
	Initialized bool
	Unlocked    bool
	Synced      bool
}

type PassphraseMsg struct {
	Method     int
	CurrentPwd string
	NewPwd     string
}

type InitWalletReply struct {
	AccountIndex int32
	Status       int
	Data         string
	Err          error
}

// AccountInfo contains info about a wallet account.
type AccountInfo struct {
	Index               uint32
	DerivationPath      string
	Xpub                string
	LastExternalDerived uint32
	LastInternalDerived uint32
}

// SwapInfo contains info about a swap
type SwapInfo struct {
	AmountP uint64
	AssetP  string
	AmountR uint64
	AssetR  string
}

type SwapFailInfo struct {
	Code    int
	Message string
}

// TradeInfo contains info about a trade.
type TradeInfo struct {
	ID               string
	Status           domain.Status
	SwapInfo         SwapInfo
	SwapFailInfo     SwapFailInfo
	MarketWithFee    MarketWithFee
	Price            Price
	TxURL            string
	RequestTimeUnix  uint64
	AcceptTimeUnix   uint64
	CompleteTimeUnix uint64
	SettleTimeUnix   uint64
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
	Balance      Balance
}

type Market struct {
	BaseAsset  string
	QuoteAsset string
}

func (m Market) Validate() error {
	if err := validateAssetString(m.BaseAsset); err != nil {
		return ErrMarketInvalidBaseAsset
	}
	if err := validateAssetString(m.QuoteAsset); err != nil {
		return ErrMarketInvalidQuoteAsset
	}
	return nil
}

type Fee struct {
	BasisPoint    int64
	FixedBaseFee  int64
	FixedQuoteFee int64
}

type MarketWithFee struct {
	Market
	Fee
}

type MarketWithPrice struct {
	Market
	Price
}

type Price struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}

func (p Price) Validate() error {
	zero := decimal.NewFromInt(0)
	if p.BasePrice.LessThanOrEqual(zero) {
		return domain.ErrMarketInvalidBasePrice
	}
	if p.QuotePrice.LessThanOrEqual(zero) {
		return domain.ErrMarketInvalidQuotePrice
	}
	return nil
}

type PriceWithFee struct {
	Price   Price
	Fee     Fee
	Amount  uint64
	Asset   string
	Balance Balance
}

type MarketStrategy struct {
	Market
	Strategy domain.StrategyType
}

type Balance struct {
	BaseAmount  uint64
	QuoteAmount uint64
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

type FragmenterSplitFundsReply struct {
	Msg string
	Err error
}

type WithdrawMarketReq struct {
	Market
	BalanceToWithdraw Balance
	MillisatPerByte   int64
	Address           string
	Push              bool
}

func (r WithdrawMarketReq) Validate() error {
	if err := validateAssetString(r.BaseAsset); err != nil {
		return ErrMarketInvalidBaseAsset
	}
	if err := validateAssetString(r.QuoteAsset); err != nil {
		return ErrMarketInvalidQuoteAsset
	}
	if r.BalanceToWithdraw.BaseAmount == 0 && r.BalanceToWithdraw.QuoteAmount == 0 {
		return ErrMissingMarketBalanceToWithdraw
	}
	if r.Address == "" {
		return ErrMissingWithdrawAddress
	}
	return nil
}

type WithdrawFeeReq struct {
	Amount          uint64
	MillisatPerByte uint64
	Address         string
	Asset           string
	Push            bool
}

func (r WithdrawFeeReq) Validate() error {
	if r.Amount == 0 {
		return fmt.Errorf("amount must not be 0")
	}
	if r.Address == "" {
		return fmt.Errorf("address must not be null")
	}
	if r.Asset != "" {
		if err := validateAssetString(r.Asset); err != nil {
			return err
		}
	}

	return nil
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
	TradeID             string
	BasisPoint          int64
	Asset               string
	PercentageFeeAmount uint64
	FixedFeeAmount      uint64
	MarketPrice         decimal.Decimal
}

type TxOutpoint struct {
	Hash  string
	Index int
}

type TxOut struct {
	Asset   string
	Value   int64
	Address string
}

func NewTxOut(address, asset string, value int64) TxOut {
	return TxOut{asset, value, address}
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

type Webhook struct {
	ActionType int
	Endpoint   string
	Secret     string
}
type WebhookInfo struct {
	Id         string
	ActionType int
	Endpoint   string
	IsSecured  bool
}

type Unspents []domain.Unspent

func (u Unspents) ToUtxos() []explorer.Utxo {
	l := make([]explorer.Utxo, 0, len(u))
	for i := range u {
		l = append(l, u[i].ToUtxo())
	}
	return l
}

type Page domain.Page

func (p *Page) ToDomain() domain.Page {
	return domain.NewPage(p.Number, p.Size)
}

type Deposits []domain.Deposit
type Withdrawals []domain.Withdrawal

type UnblindedResult *transactionutil.UnblindedResult
type BlindingData wallet.BlindingData

type Blinder interface {
	UnblindOutput(txout *transaction.TxOutput, key []byte) (UnblindedResult, bool)
}

type FillProposalOpts struct {
	Mnemonic         []string
	SwapRequest      domain.SwapRequest
	MarketUtxos      []explorer.Utxo
	FeeUtxos         []explorer.Utxo
	MarketInfo       domain.AddressesInfo
	FeeInfo          domain.AddressesInfo
	OutputInfo       domain.AddressInfo
	ChangeInfo       domain.AddressInfo
	FeeChangeInfo    domain.AddressInfo
	Network          *network.Network
	MilliSatsPerByte int
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

type TransactionHandler interface {
	ExtractUnspents(
		txhex string,
		infoByScript map[string]domain.AddressInfo,
		network *network.Network,
	) ([]domain.Unspent, []domain.UnspentKey, error)
	ExtractBlindingData(
		psetBase64 string,
		inBlindingKeys, outBlidningKeys map[string][]byte,
	) (map[int]BlindingData, map[int]BlindingData, error)
}

var (
	BlinderManager     Blinder
	TradeManager       TradeHandler
	TransactionManager TransactionHandler
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

type transactionManager struct{}

func (t transactionManager) ExtractUnspents(
	txHex string,
	infoByScript map[string]domain.AddressInfo,
	network *network.Network,
) ([]domain.Unspent, []domain.UnspentKey, error) {
	tx, err := transaction.NewTxFromHex(txHex)
	if err != nil {
		return nil, nil, err
	}

	unspentsToAdd := make([]domain.Unspent, 0)
	unspentsToSpend := make([]domain.UnspentKey, 0)

	for _, in := range tx.Inputs {
		// our unspents are native-segiwt only
		if len(in.Witness) > 0 {
			pubkey, _ := btcec.ParsePubKey(in.Witness[1], btcec.S256())
			p := payment.FromPublicKey(pubkey, network, nil)

			script := hex.EncodeToString(p.WitnessScript)
			if _, ok := infoByScript[script]; ok {
				unspentsToSpend = append(unspentsToSpend, domain.UnspentKey{
					TxID: bufferutil.TxIDFromBytes(in.Hash),
					VOut: in.Index,
				})
			}
		}
	}

	for i, out := range tx.Outputs {
		script := hex.EncodeToString(out.Script)
		if info, ok := infoByScript[script]; ok {
			unconfidential, ok := transactionutil.UnblindOutput(out, info.BlindingKey)
			if !ok {
				return nil, nil, fmt.Errorf("unable to unblind output")
			}
			unspentsToAdd = append(unspentsToAdd, domain.Unspent{
				TxID:            tx.TxHash().String(),
				VOut:            uint32(i),
				Value:           unconfidential.Value,
				AssetHash:       unconfidential.AssetHash,
				ValueCommitment: bufferutil.CommitmentFromBytes(out.Value),
				AssetCommitment: bufferutil.CommitmentFromBytes(out.Asset),
				ValueBlinder:    unconfidential.ValueBlinder,
				AssetBlinder:    unconfidential.AssetBlinder,
				ScriptPubKey:    out.Script,
				Nonce:           out.Nonce,
				RangeProof:      make([]byte, 1),
				SurjectionProof: make([]byte, 1),
				Address:         info.Address,
				Confirmed:       false,
			})
		}
	}
	return unspentsToAdd, unspentsToSpend, nil
}

func (t transactionManager) ExtractBlindingData(
	psetBase64 string,
	inBlindingKeys, outBlindingKeys map[string][]byte,
) (inBlindingData, outBlindingData map[int]BlindingData, err error) {
	in, out, err := wallet.ExtractBlindingDataFromTx(psetBase64, inBlindingKeys, outBlindingKeys)
	if err != nil {
		return
	}

	if in != nil {
		inBlindingData = make(map[int]BlindingData)
		for i, d := range in {
			inBlindingData[i] = BlindingData(d)
		}
	}
	if out != nil {
		outBlindingData = make(map[int]BlindingData)
		for i, d := range out {
			outBlindingData[i] = BlindingData(d)
		}
	}
	return
}

func init() {
	BlinderManager = blinderManager{}
	TradeManager = tradeManager{}
	TransactionManager = transactionManager{}
}
