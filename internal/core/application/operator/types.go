package operator

import (
	"sync"
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

type marketFeeInfo domain.MarketFee

func (i marketFeeInfo) GetBaseAsset() uint64 {
	return i.BaseAsset
}
func (i marketFeeInfo) GetQuoteAsset() uint64 {
	return i.QuoteAsset
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
func (i marketInfo) GetName() string {
	return i.Name
}
func (i marketInfo) GetBaseAssetPrecision() uint32 {
	return uint32(i.Market.BaseAssetPrecision)
}
func (i marketInfo) GetQuoteAssetPrecision() uint32 {
	return uint32(i.Market.QuoteAssetPrecision)
}
func (i marketInfo) GetPercentageFee() ports.MarketFee {
	return marketFeeInfo(i.PercentageFee)
}
func (i marketInfo) GetFixedFee() ports.MarketFee {
	return marketFeeInfo(i.FixedFee)
}
func (i marketInfo) IsTradable() bool {
	return i.Tradable
}
func (i marketInfo) GetStrategyType() ports.MarketStrategy {
	return marketStrategyInfo(i.StrategyType)
}
func (i marketInfo) GetMarket() ports.Market {
	return i
}
func (i marketInfo) GetPrice() ports.MarketPrice {
	return i.Price
}
func (i marketInfo) GetBalance() map[string]ports.Balance {
	return i.balance
}

type tradeTypeInfo domain.TradeType

func (i tradeTypeInfo) IsBuy() bool {
	return domain.TradeType(i) == domain.TradeBuy
}
func (i tradeTypeInfo) IsSell() bool {
	return domain.TradeType(i) == domain.TradeSell
}

type tradeStatusInfo domain.TradeStatus

func (s tradeStatusInfo) IsRequest() bool {
	return s.Code == domain.TradeStatusCodeProposal
}
func (s tradeStatusInfo) IsAccept() bool {
	return s.Code == domain.TradeStatusCodeAccepted
}
func (s tradeStatusInfo) IsComplete() bool {
	return s.Code == domain.TradeStatusCodeCompleted
}
func (s tradeStatusInfo) IsSettled() bool {
	return s.Code == domain.TradeStatusCodeSettled
}
func (s tradeStatusInfo) IsExpired() bool {
	return s.Code == domain.TradeStatusCodeExpired
}
func (s tradeStatusInfo) IsFailed() bool {
	return s.Failed
}

type tradeInfo struct {
	domain.Trade
}

func (i tradeInfo) GetId() string {
	return i.Trade.Id
}
func (i tradeInfo) GetType() ports.TradeType {
	return tradeTypeInfo(i.Trade.Type)
}
func (i tradeInfo) GetStatus() ports.TradeStatus {
	return tradeStatusInfo(i.Trade.Status)
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
func (i tradeInfo) GetMarketPercentageFee() ports.MarketFee {
	return marketFeeInfo(i.Trade.MarketPercentageFee)
}
func (i tradeInfo) GetMarketFixedFee() ports.MarketFee {
	return marketFeeInfo(i.Trade.MarketFixedFee)
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
func (i tradeInfo) GetMarketPrice() ports.MarketPrice {
	return i.MarketPrice
}
func (i tradeInfo) GetTxid() string {
	return i.Trade.TxId
}
func (i tradeInfo) GetTxHex() string {
	return i.Trade.TxHex
}
func (i tradeInfo) GetFeeAsset() string {
	return i.Trade.FeeAsset
}
func (i tradeInfo) GetFeeAmount() uint64 {
	return i.Trade.FeeAmount
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
	account string
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
	if i.account == domain.FeeAccount {
		return len(i.ownedInputs) > 0 && len(i.notOwnedInputs) <= 0 &&
			len(i.notOwnedOutputs) > 0
	}

	inAssets := make(map[string]struct{}, 0)
	for _, in := range i.ownedInputs {
		inAssets[in.asset] = struct{}{}
	}
	for _, out := range i.ownedOutputs {
		if _, ok := inAssets[out.asset]; !ok {
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
	marketPrice string
}

func (i tradeFeeInfo) GetTradeId() string {
	return i.Trade.Id
}
func (i tradeFeeInfo) GetPercentageFee() uint64 {
	if i.Trade.FeeAsset == i.Trade.MarketBaseAsset {
		return i.Trade.MarketPercentageFee.BaseAsset
	}
	return i.Trade.MarketPercentageFee.QuoteAsset
}
func (i tradeFeeInfo) GetFixedFee() uint64 {
	if i.Trade.FeeAsset == i.Trade.MarketBaseAsset {
		return i.Trade.MarketFixedFee.BaseAsset
	}
	return i.Trade.MarketFixedFee.QuoteAsset
}
func (i tradeFeeInfo) GetFeeAsset() string {
	return i.Trade.FeeAsset
}
func (i tradeFeeInfo) GetFeeAmount() uint64 {
	return i.Trade.FeeAmount
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

type accountMap struct {
	lock              *sync.RWMutex
	labelsByNamespace map[string]string
}

func (m *accountMap) add(namespace, label string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.labelsByNamespace[namespace]; ok {
		return
	}

	m.labelsByNamespace[namespace] = label
}

func (m *accountMap) getLabel(namespace string) string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.labelsByNamespace[namespace]
}
