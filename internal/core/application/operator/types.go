package operator

import (
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
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
	LastYear
	All

	StartYear = 2021
)

// Internal types used to
type depositInfo domain.Deposit

func (i depositInfo) GetTxid() string {
	return i.TxID
}
func (i depositInfo) GetTotalAmountPerAsset() map[string]uint64 {
	return i.TotAmountPerAsset
}
func (i depositInfo) GetAccountName() string {
	return i.AccountName
}
func (i depositInfo) GetTimestamp() int64 {
	return i.Timestamp
}

type depositList []domain.Deposit

func (l depositList) toPortableList() []ports.Deposit {
	list := make([]ports.Deposit, 0, len(l))
	for _, d := range l {
		list = append(list, depositInfo(d))
	}
	return list
}

type withdrawalInfo domain.Withdrawal

func (i withdrawalInfo) GetTxid() string {
	return i.TxID
}
func (i withdrawalInfo) GetAccountName() string {
	return i.AccountName
}
func (i withdrawalInfo) GetTotalAmountPerAsset() map[string]uint64 {
	return i.TotAmountPerAsset
}
func (i withdrawalInfo) GetTimestamp() int64 {
	return i.Timestamp
}

type withdrawalList []domain.Withdrawal

func (l withdrawalList) toPortableList() []ports.Withdrawal {
	list := make([]ports.Withdrawal, 0, len(l))
	for _, w := range l {
		list = append(list, withdrawalInfo(w))
	}
	return list
}

type marketInfo struct {
	domain.Market
	balance map[string]ports.Balance
}

func (i marketInfo) GetBaseAsset() string {
	return i.BaseAsset
}
func (i marketInfo) GetQuoteAsset() string {
	return i.QuoteAsset
}
func (i marketInfo) GetAccountName() string {
	return i.Name
}
func (i marketInfo) GetBasePrice() decimal.Decimal {
	p, _ := decimal.NewFromString(i.Price.BasePrice)
	return p
}
func (i marketInfo) GetQuotePrice() decimal.Decimal {
	p, _ := decimal.NewFromString(i.Price.QuotePrice)
	return p
}
func (i marketInfo) GetPercentageFee() uint32 {
	return i.PercentageFee
}
func (i marketInfo) GetFixedBaseFee() uint64 {
	return i.FixedFee.BaseFee
}
func (i marketInfo) GetFixedQuoteFee() uint64 {
	return i.FixedFee.QuoteFee
}
func (i marketInfo) IsTradable() bool {
	return i.Tradable
}
func (i marketInfo) GetStrategyType() ports.MarketStartegy {
	return marketStrategyInfo(i.StrategyType)
}
func (i marketInfo) GetMarket() ports.Market {
	return i
}
func (i marketInfo) GetFee() ports.MarketFee {
	return i
}
func (i marketInfo) GetPrice() ports.MarketPrice {
	return i
}
func (i marketInfo) GetBalance() map[string]ports.Balance {
	return i.balance
}

type tradeStatus struct {
	domain.TradeStatus
}

func (s tradeStatus) IsRequest() bool {
	return s.TradeStatus.Code == domain.TradeStatusCodeProposal
}
func (s tradeStatus) IsAccept() bool {
	return s.TradeStatus.Code == domain.TradeStatusCodeAccepted
}
func (s tradeStatus) IsComplete() bool {
	return s.TradeStatus.Code == domain.TradeStatusCodeCompleted
}
func (s tradeStatus) IsSettled() bool {
	return s.TradeStatus.Code == domain.TradeStatusCodeSettled
}
func (s tradeStatus) IsExpired() bool {
	return s.TradeStatus.Code == domain.TradeStatusCodeExpired
}
func (s tradeStatus) IsFailed() bool {
	return s.TradeStatus.Failed
}

type tradeInfo struct {
	domain.Trade
}

func (i tradeInfo) GetId() string {
	return i.Trade.Id
}
func (i tradeInfo) GetStatus() ports.TradeStatus {
	return tradeStatus{i.Trade.Status}
}
func (i tradeInfo) GetSwapInfo() ports.SwapRequest {
	info := i.Trade
	if info.SwapRequestMessage() == nil {
		return nil
	}
	return swapRequestInfo{info.SwapRequestMessage()}
}
func (i tradeInfo) GetSwapFailInfo() ports.SwapFail {
	info := i.Trade
	if info.SwapFailMessage() == nil {
		return nil
	}
	return info.SwapFailMessage()
}
func (i tradeInfo) GetBaseAsset() string {
	return i.Trade.MarketBaseAsset
}
func (i tradeInfo) GetQuoteAsset() string {
	return i.Trade.MarketQuoteAsset
}
func (i tradeInfo) GetBasePrice() decimal.Decimal {
	p, _ := decimal.NewFromString(i.Trade.MarketPrice.BasePrice)
	return p
}
func (i tradeInfo) GetQuotePrice() decimal.Decimal {
	p, _ := decimal.NewFromString(i.Trade.MarketPrice.QuotePrice)
	return p
}
func (i tradeInfo) GetPercentageFee() uint32 {
	return i.Trade.MarketPercentageFee
}
func (i tradeInfo) GetFixedBaseFee() uint64 {
	return i.Trade.MarketFixedBaseFee
}
func (i tradeInfo) GetFixedQuoteFee() uint64 {
	return i.Trade.MarketFixedQuoteFee
}
func (i tradeInfo) GetRequestTimestamp() int64 {
	info := i.Trade
	if info.SwapRequest == nil {
		return 0
	}
	return info.SwapRequest.Timestamp
}
func (i tradeInfo) GetAcceptTimestamp() int64 {
	info := i.Trade
	if info.SwapAccept == nil {
		return 0
	}
	return info.SwapAccept.Timestamp
}
func (i tradeInfo) GetCompleteTimestamp() int64 {
	info := i.Trade
	if info.SwapComplete == nil {
		return 0
	}
	return info.SwapComplete.Timestamp
}
func (i tradeInfo) GetSettleTimestamp() int64 {
	return i.Trade.SettlementTime
}
func (i tradeInfo) GetExpiryTimestamp() int64 {
	return i.Trade.ExpiryTime
}
func (i tradeInfo) GetMarket() ports.Market {
	return i
}
func (i tradeInfo) GetMarketFee() ports.MarketFee {
	return i
}
func (i tradeInfo) GetMarketPrice() ports.MarketPrice {
	return i
}

type marketStrategyInfo int

func (i marketStrategyInfo) IsBalanced() bool {
	return i == domain.StrategyTypeBalanced
}
func (i marketStrategyInfo) IsPluggable() bool {
	return i == domain.StrategyTypePluggable
}

type tradeList []domain.Trade

func (l tradeList) toPortableList() []ports.Trade {
	list := make([]ports.Trade, 0)
	for _, t := range l {
		list = append(list, tradeInfo{t})
	}
	return list
}

type swapRequestInfo struct {
	*domain.SwapRequest
}

func (i swapRequestInfo) GetUnblindedInputs() []ports.UnblindedInput {
	info := i.SwapRequest
	list := make([]ports.UnblindedInput, 0, len(info.UnblindedInputs))
	for _, in := range info.UnblindedInputs {
		list = append(list, in)
	}
	return list
}

type txInfo struct {
	transaction.Transaction
	ownedInputs     []txOutputInfo
	notOwnedInputs  []txOutputInfo
	ownedOutputs    []txOutputInfo
	notOwnedOutputs []txOutputInfo
	fee             uint64
}

func (i txInfo) isDeposit() bool {
	return len(i.ownedInputs) == 0 && len(i.ownedOutputs) > 0
}

func (i txInfo) isWithdrawal() bool {
	if len(i.ownedInputs) <= 0 || len(i.notOwnedInputs) > 0 {
		return false
	}

	inAssets, outAssets := make(map[string]struct{}), make(map[string]struct{})
	for _, in := range i.ownedInputs {
		inAssets[in.asset] = struct{}{}
	}
	for _, out := range i.ownedOutputs {
		outAssets[out.asset] = struct{}{}
	}

	for inAsset := range inAssets {
		if _, ok := outAssets[inAsset]; !ok {
			return false
		}
	}
	return true
}

func (i txInfo) depositAmountPerAsset() map[string]uint64 {
	tot := make(map[string]uint64)
	for _, out := range i.ownedOutputs {
		tot[out.asset] += out.amount
	}
	return tot
}

func (i txInfo) withdrawalAmountPerAsset() map[string]uint64 {
	inTotAmountPerAsset := make(map[string]uint64)
	for _, in := range i.ownedInputs {
		inTotAmountPerAsset[in.asset] += in.amount
	}
	outTotAmountPerAsset := make(map[string]uint64)
	for _, out := range i.ownedOutputs {
		outTotAmountPerAsset[out.asset] += out.amount
	}

	totAmountPerAsset := make(map[string]uint64)
	for asset, amount := range inTotAmountPerAsset {
		totAmountPerAsset[asset] = amount - outTotAmountPerAsset[asset]
	}

	return totAmountPerAsset
}

type txOutputInfo struct {
	asset  string
	amount uint64
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
	if *p > All {
		return fmt.Errorf("PredefinedPeriod cant be > %v", All)
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
		case LastYear:
			y, _, _ := now.Date()
			startTime = time.Date(y-1, time.January, 1, 0, 0, 0, 0, time.UTC)
			endTime = time.Date(y-1, time.December, 31, 23, 59, 59, 0, time.UTC)
			return
		case All:
			start = time.Date(StartYear, time.January, 1, 0, 0, 0, 0, time.UTC)
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

type Market struct {
	BaseAsset  string
	QuoteAsset string
}

type MarketReport struct {
	Market          Market
	CollectedFees   MarketCollectedFees
	TotalVolume     MarketVolume
	VolumesPerFrame []MarketVolume
}

type MarketCollectedFees struct {
	BaseAmount   uint64
	QuoteAmount  uint64
	TradeFeeInfo []TradeFeeInfo
	StartTime    time.Time
	EndTime      time.Time
}

type MarketVolume struct {
	BaseVolume  uint64
	QuoteVolume uint64
	StartTime   time.Time
	EndTime     time.Time
}

type TradeFeeInfo struct {
	TradeID             string
	PercentageFee       uint64
	FeeAsset            string
	PercentageFeeAmount uint64
	FixedFeeAmount      uint64
	MarketPrice         decimal.Decimal
}
