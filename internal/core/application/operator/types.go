package operator

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/vulpemventures/go-elements/transaction"
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
func (i marketInfo) GetBaseAssetPrecision() uint32 {
	return uint32(i.Market.BaseAssetPrecision)
}
func (i marketInfo) GetQuoteAssetPrecision() uint32 {
	return uint32(i.Market.QuoteAssetPrecision)
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

type marketVolumeInfo struct {
	start       time.Time
	end         time.Time
	baseVolume  uint64
	quoteVolume uint64
}

func (i marketVolumeInfo) GetBaseVolume() uint64 {
	return i.baseVolume
}
func (i marketVolumeInfo) GetQuoteVolume() uint64 {
	return i.quoteVolume
}
func (i marketVolumeInfo) GetStartDate() string {
	return i.start.Format(time.RFC3339)
}
func (i marketVolumeInfo) GetEndDate() string {
	return i.end.Format(time.RFC3339)
}

type marketVolumeInfoList []*marketVolumeInfo

func (l marketVolumeInfoList) toPortableList() []ports.MarketVolume {
	list := make([]ports.MarketVolume, 0, len(l))
	for _, v := range l {
		list = append(list, *v)
	}
	return list
}

type tradeFeeInfo struct {
	domain.Trade
	feeAsset            string
	percentageFeeAmount uint64
	fixedFeeAmount      uint64
	marketPrice         string
}

func (i tradeFeeInfo) GetTradeId() string {
	return i.Trade.Id
}
func (i tradeFeeInfo) GetPercentageFee() uint64 {
	return uint64(i.Trade.MarketPercentageFee)
}
func (i tradeFeeInfo) GetFeeAsset() string {
	return i.feeAsset
}
func (i tradeFeeInfo) GetPercentageFeeAmount() uint64 {
	return i.percentageFeeAmount
}
func (i tradeFeeInfo) GetFixedFeeAmount() uint64 {
	return i.fixedFeeAmount
}
func (i tradeFeeInfo) GetMarketPrice() decimal.Decimal {
	p, _ := decimal.NewFromString(i.marketPrice)
	return p
}
func (i tradeFeeInfo) GetTimestamp() int64 {
	return i.Trade.SwapRequest.Timestamp
}

type tradeFeeInfoList []tradeFeeInfo

func (l tradeFeeInfoList) toPortableList() []ports.TradeFeeInfo {
	list := make([]ports.TradeFeeInfo, 0, len(l))
	for _, v := range l {
		list = append(list, v)
	}
	return list
}

type marketFeeReportInfo struct {
	start       time.Time
	end         time.Time
	totBaseFee  uint64
	totQuoteFee uint64
	tradeFees   tradeFeeInfoList
}

func (i marketFeeReportInfo) GetBaseAmount() uint64 {
	return i.totBaseFee
}
func (i marketFeeReportInfo) GetQuoteAmount() uint64 {
	return i.totQuoteFee
}
func (i marketFeeReportInfo) GetTradeFeeInfo() []ports.TradeFeeInfo {
	return i.tradeFees.toPortableList()
}
func (i marketFeeReportInfo) GetStartDate() string {
	return i.start.Format(time.RFC3339)
}
func (i marketFeeReportInfo) GetEndDate() string {
	return i.end.Format(time.RFC3339)
}

type marketReportInfo struct {
	domain.Market
	marketFeeReportInfo
	totVolume  marketVolumeInfo
	subVolumes marketVolumeInfoList
}

func (i marketReportInfo) GetMarket() ports.Market {
	return i
}
func (i marketReportInfo) GetBaseAsset() string {
	return i.Market.BaseAsset
}
func (i marketReportInfo) GetQuoteAsset() string {
	return i.Market.QuoteAsset
}
func (i marketReportInfo) GetCollectedFees() ports.MarketCollectedFees {
	return i.marketFeeReportInfo
}
func (i marketReportInfo) GetTotalVolume() ports.MarketVolume {
	return i.totVolume
}
func (i marketReportInfo) GetVolumesPerFrame() []ports.MarketVolume {
	return i.subVolumes.toPortableList()
}
