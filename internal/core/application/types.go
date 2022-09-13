package application

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec"
	validation "github.com/go-ozzo/ozzo-validation/v4"
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

const (
	// NIL is added in proto file to recognised when predefined period is passed
	NIL PredefinedPeriod = iota
	LastHour
	LastDay
	LastWeek
	LastMonth
	LastThreeMonths
	YearToDate
	All
	LastYear

	StartYear = 2021
)

// HDWalletInfo contains info about the internal wallet of the daemon and its
// sub-accounts.
type HDWalletInfo struct {
	RootPath          string
	MasterBlindingKey string
	Accounts          []AccountInfo
	Network           string
	BuildInfo         BuildInfo
	BaseAsset         string
	QuoteAsset        string
}

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
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
	if m.BaseAsset == m.QuoteAsset {
		return fmt.Errorf("quote asset must not be equal to base asset")
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
	Password          string
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
	if r.Password == "" {
		return ErrMissingWithdrawPassword
	}
	return nil
}

type WithdrawFeeReq struct {
	Amount          uint64
	MillisatPerByte uint64
	Address         string
	Asset           string
	Password        string
	Push            bool
}

func (r WithdrawFeeReq) Validate() error {
	if r.Amount == 0 {
		return fmt.Errorf("amount must not be 0")
	}
	if r.Address == "" {
		return ErrMissingWithdrawAddress
	}
	if r.Asset != "" {
		if err := validateAssetString(r.Asset); err != nil {
			return err
		}
	}
	if r.Password == "" {
		return ErrMissingWithdrawPassword
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

type MarketReport struct {
	CollectedFees MarketCollectedFees
	Volume        MarketVolume
	GroupedVolume []MarketVolume
}

type MarketCollectedFees struct {
	BaseAmount  uint64
	QuoteAmount uint64
	StartTime   time.Time
	EndTime     time.Time
}

type MarketVolume struct {
	BaseVolume  uint64
	QuoteVolume uint64
	StartTime   time.Time
	EndTime     time.Time
}

type TimeRange struct {
	PredefinedPeriod *PredefinedPeriod
	CustomPeriod     *CustomPeriod
}

func (t *TimeRange) Validate() error {
	if t.CustomPeriod == nil && t.PredefinedPeriod == nil {
		return errors.New("both PredefinedPeriod period and CustomPeriod cant be null")
	}

	if t.CustomPeriod != nil && t.PredefinedPeriod != nil {
		return errors.New("both PredefinedPeriod period and CustomPeriod provided, please provide only one")
	}

	if t.CustomPeriod != nil {
		if err := t.CustomPeriod.validate(); err != nil {
			return err
		}
	}

	if t.PredefinedPeriod != nil {
		if err := t.PredefinedPeriod.validate(); err != nil {
			return err
		}
	}

	return nil
}

type PredefinedPeriod int

func (p *PredefinedPeriod) validate() error {
	if *p > LastYear {
		return fmt.Errorf("PredefinedPeriod cant be > %v", LastYear)
	}

	lastYear := time.Now().Year() - 1
	if lastYear < StartYear {
		return fmt.Errorf("no available data prior to year: %v", StartYear)
	}

	return nil
}

type CustomPeriod struct {
	StartDate string
	EndDate   string
}

func (c *CustomPeriod) validate() error {
	if err := validation.ValidateStruct(
		c,
		validation.Field(&c.StartDate, validation.By(validateTimeFormat)),
		validation.Field(&c.EndDate, validation.By(validateTimeFormat)),
	); err != nil {
		return err
	}

	start, _ := time.Parse(time.RFC3339, c.StartDate)
	end, _ := time.Parse(time.RFC3339, c.EndDate)

	if !start.Before(end) {
		return errors.New("startTime must be before endTime")
	}

	return nil
}

func (t *TimeRange) getStartAndEndTime(now time.Time) (startTime time.Time, endTime time.Time, err error) {
	if err = t.Validate(); err != nil {
		return
	}

	if t.CustomPeriod != nil {
		start, _ := time.Parse(time.RFC3339, t.CustomPeriod.StartDate)
		startTime = start

		endTime = now
		if t.CustomPeriod.EndDate != "" {
			end, _ := time.Parse(time.RFC3339, t.CustomPeriod.EndDate)
			endTime = end
		}
		return
	}

	if t.PredefinedPeriod != nil {
		var start time.Time
		switch *t.PredefinedPeriod {
		case LastHour:
			start = now.Add(time.Duration(-60) * time.Minute)
		case LastDay:
			start = now.AddDate(0, 0, -1)
		case LastWeek:
			start = now.AddDate(0, 0, -7)
		case LastMonth:
			start = now.AddDate(0, -1, 0)
		case LastThreeMonths:
			start = now.AddDate(0, -3, 0)
		case YearToDate:
			y, _, _ := now.Date()
			start = time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC)
		case All:
			start = time.Date(StartYear, time.January, 1, 0, 0, 0, 0, time.UTC)
		case LastYear:
			y, _, _ := now.Date()
			startTime = time.Date(y-1, time.January, 1, 0, 0, 0, 0, time.UTC)
			endTime = time.Date(y-1, time.December, 31, 23, 59, 59, 0, time.UTC)
			return
		}

		startTime = start
		endTime = now
	}

	return
}

func validateTimeFormat(t interface{}) error {
	tm, ok := t.(string)
	if !ok {
		return ErrInvalidTime
	}

	if _, err := time.Parse(time.RFC3339, tm); err != nil {
		return ErrInvalidTimeFormat
	}

	return nil
}

type utxoKey struct {
	txid  string
	index uint32
}

func (u utxoKey) Hash() string {
	return u.txid
}
func (u utxoKey) Index() uint32 {
	return u.index
}
func (u utxoKey) Value() uint64 {
	return 0
}
func (u utxoKey) Asset() string {
	return ""
}
func (u utxoKey) ValueCommitment() string {
	return ""
}
func (u utxoKey) AssetCommitment() string {
	return ""
}
func (u utxoKey) ValueBlinder() []byte {
	return nil
}
func (u utxoKey) AssetBlinder() []byte {
	return nil
}
func (u utxoKey) Script() []byte {
	return nil
}
func (u utxoKey) Nonce() []byte {
	return nil
}
func (u utxoKey) RangeProof() []byte {
	return nil
}
func (u utxoKey) SurjectionProof() []byte {
	return nil
}
func (u utxoKey) IsConfidential() bool {
	return false
}
func (u utxoKey) IsConfirmed() bool {
	return false
}
func (u utxoKey) IsRevealed() bool {
	return false
}
func (u utxoKey) Parse() (*transaction.TxInput, *transaction.TxOutput, error) {
	return nil, nil, nil
}
