package grpchandler

import (
	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type market struct {
	ports.Market
}

func (m market) toProto() *tdexv1.Market {
	return &tdexv1.Market{
		BaseAsset:  m.GetBaseAsset(),
		QuoteAsset: m.GetQuoteAsset(),
	}
}

type marketInfo struct {
	ports.MarketInfo
}

func (i marketInfo) toProto() *daemonv2.MarketInfo {
	info := i.MarketInfo
	balance := make(map[string]*daemonv2.Balance)
	for asset, bal := range info.GetBalance() {
		balance[asset] = marketBalanceInfo{bal}.toProto()
	}

	return &daemonv2.MarketInfo{
		Market:       market{info.GetMarket()}.toProto(),
		Fee:          marketFeeInfo{info.GetFee()}.toProto(),
		Tradable:     info.IsTradable(),
		StrategyType: marketStrategyInfo{info.GetStrategyType()}.toProto(),
		AccountName:  info.GetAccountName(),
		Price:        marketPriceInfo{info.GetPrice()}.toProto(),
		Balance:      balance,
	}
}

type marketFeeInfo struct {
	ports.MarketFee
}

func (i marketFeeInfo) toProto() *tdexv1.Fee {
	info := i.MarketFee
	return &tdexv1.Fee{
		BasisPoint: int64(info.GetPercentageFee()),
		Fixed: &tdexv1.Fixed{
			BaseFee:  int64(info.GetFixedBaseFee()),
			QuoteFee: int64(info.GetFixedQuoteFee()),
		},
	}
}

type marketPriceInfo struct {
	ports.MarketPrice
}

func (i marketPriceInfo) toProto() *tdexv1.Price {
	info := i.MarketPrice
	basePrice, _ := info.GetBasePrice().Float64()
	quotePrice, _ := info.GetQuotePrice().Float64()
	return &tdexv1.Price{
		BasePrice:  basePrice,
		QuotePrice: quotePrice,
	}
}

type marketBalanceInfo struct {
	ports.Balance
}

func (i marketBalanceInfo) toProto() *daemonv2.Balance {
	info := i.Balance
	if info == nil {
		return &daemonv2.Balance{}
	}
	return &daemonv2.Balance{
		ConfirmedBalance:   info.GetConfirmedBalance(),
		UnconfirmedBalance: info.GetUnconfirmedBalance(),
		LockedBalance:      info.GetLockedBalance(),
		TotalBalance:       info.GetTotalBalance(),
	}
}

type marketStrategyInfo struct {
	ports.MarketStartegy
}

func (i marketStrategyInfo) toProto() daemonv2.StrategyType {
	if i.MarketStartegy.IsBalanced() {
		return daemonv2.StrategyType_STRATEGY_TYPE_BALANCED
	}
	if i.MarketStartegy.IsPluggable() {
		return daemonv2.StrategyType_STRATEGY_TYPE_PLUGGABLE
	}
	return daemonv2.StrategyType_STRATEGY_TYPE_UNSPECIFIED
}

type marketReportInfo struct {
	ports.MarketReport
}

func (i marketReportInfo) toProto() *daemonv2.MarketReport {
	report := i.MarketReport
	volumesPerFrame := make(
		[]*daemonv2.MarketVolume, 0, len(report.GetVolumesPerFrame()),
	)
	for _, info := range report.GetVolumesPerFrame() {
		volumesPerFrame = append(volumesPerFrame, volumeInfo{info}.toProto())
	}

	return &daemonv2.MarketReport{
		TotalCollectedFees: collectedFeesInfo{report.GetCollectedFees()}.toProto(),
		TotalVolume:        volumeInfo{report.GetTotalVolume()}.toProto(),
		VolumesPerFrame:    volumesPerFrame,
	}
}

type collectedFeesInfo struct {
	ports.MarketCollectedFees
}

func (i collectedFeesInfo) toProto() *daemonv2.MarketCollectedFees {
	info := i.MarketCollectedFees
	feesPerTrade := make([]*daemonv2.FeeInfo, 0, len(info.GetTradeFeeInfo()))
	for _, i := range info.GetTradeFeeInfo() {
		price, _ := i.GetMarketPrice().Float64()
		feesPerTrade = append(feesPerTrade, &daemonv2.FeeInfo{
			TradeId:             i.GetTradeId(),
			BasisPoint:          int64(i.GetPercentageFee()),
			Asset:               i.GetFeeAsset(),
			PercentageFeeAmount: i.GetPercentageFeeAmount(),
			FixedFeeAmount:      i.GetFixedFeeAmount(),
			MarketPrice:         price,
		})
	}
	return &daemonv2.MarketCollectedFees{
		BaseAmount:   info.GetBaseAmount(),
		QuoteAmount:  info.GetQuoteAmount(),
		StartDate:    info.GetStartDate(),
		EndDate:      info.GetEndDate(),
		FeesPerTrade: feesPerTrade,
	}
}

type volumeInfo struct {
	ports.MarketVolume
}

func (i volumeInfo) toProto() *daemonv2.MarketVolume {
	info := i.MarketVolume
	return &daemonv2.MarketVolume{
		BaseVolume:  info.GetBaseVolume(),
		QuoteVolume: info.GetQuoteVolume(),
		StartDate:   info.GetStartDate(),
		EndDate:     info.GetEndDate(),
	}
}

type tradeTypeInfo tdexv1.TradeType

func (i tradeTypeInfo) IsBuy() bool {
	return tdexv1.TradeType(i) == tdexv1.TradeType_TRADE_TYPE_BUY
}
func (i tradeTypeInfo) IsSell() bool {
	return tdexv1.TradeType(i) == tdexv1.TradeType_TRADE_TYPE_SELL
}

type tradesInfo []ports.Trade

func (i tradesInfo) toProto() []*daemonv2.TradeInfo {
	list := make([]*daemonv2.TradeInfo, 0, len(i))
	for _, info := range i {
		list = append(list, &daemonv2.TradeInfo{
			TradeId:  info.GetId(),
			Status:   tradeStatusInfo{info.GetStatus()}.toProto(),
			SwapInfo: swapInfo{info.GetSwapInfo()}.toProto(),
			FailInfo: failInfo{info.GetSwapFailInfo()}.toProto(),
			MarketWithFee: &tdexv1.MarketWithFee{
				Market: market{info.GetMarket()}.toProto(),
				Fee:    marketFeeInfo{info.GetMarketFee()}.toProto(),
			},
			Price:             marketPriceInfo{info.GetMarketPrice()}.toProto(),
			RequestTimestamp:  info.GetRequestTimestamp(),
			AcceptTimestamp:   info.GetAcceptTimestamp(),
			CompleteTimestamp: info.GetCompleteTimestamp(),
			SettleTimestamp:   info.GetSettleTimestamp(),
			ExpiryTimestamp:   info.GetExpiryTimestamp(),
			RequestDate:       timestampToString(info.GetRequestTimestamp()),
			AcceptDate:        timestampToString(info.GetAcceptTimestamp()),
			CompleteDate:      timestampToString(info.GetCompleteTimestamp()),
			SettleDate:        timestampToString(info.GetSettleTimestamp()),
			ExpiryDate:        timestampToString(info.GetExpiryTimestamp()),
		})
	}
	return list
}

type tradeStatusInfo struct {
	ports.TradeStatus
}

func (i tradeStatusInfo) toProto() *daemonv2.TradeStatusInfo {
	status := daemonv2.TradeStatus_TRADE_STATUS_UNSPECIFIED
	if i.TradeStatus.IsRequest() {
		status = daemonv2.TradeStatus_TRADE_STATUS_REQUEST
	}
	if i.TradeStatus.IsAccept() {
		status = daemonv2.TradeStatus_TRADE_STATUS_ACCEPT
	}
	if i.TradeStatus.IsComplete() {
		status = daemonv2.TradeStatus_TRADE_STATUS_COMPLETE
	}
	if i.TradeStatus.IsSettled() {
		status = daemonv2.TradeStatus_TRADE_STATUS_SETTLED
	}
	if i.TradeStatus.IsExpired() {
		status = daemonv2.TradeStatus_TRADE_STATUS_EXPIRED
	}
	return &daemonv2.TradeStatusInfo{
		Status: status,
		Failed: i.TradeStatus.IsFailed(),
	}
}

type swapRequestInfo struct {
	*tdexv1.SwapRequest
}

func (i swapRequestInfo) GetUnblindedInputs() []ports.UnblindedInput {
	info := i.SwapRequest
	list := make([]ports.UnblindedInput, 0, len(info.GetUnblindedInputs()))
	for _, in := range info.GetUnblindedInputs() {
		list = append(list, in)
	}
	return list
}

type swapAcceptInfo struct {
	ports.SwapAccept
}

func (i swapAcceptInfo) toProto() *tdexv1.SwapAccept {
	info := i.SwapAccept
	if info == nil {
		return nil
	}
	return &tdexv1.SwapAccept{
		Id:          info.GetId(),
		RequestId:   info.GetRequestId(),
		Transaction: info.GetTransaction(),
	}
}

type swapFailInfo struct {
	ports.SwapFail
}

func (i swapFailInfo) toProto() *tdexv1.SwapFail {
	info := i.SwapFail
	if info == nil {
		return nil
	}
	return &tdexv1.SwapFail{
		Id:             info.GetId(),
		MessageId:      info.GetMessageId(),
		FailureCode:    info.GetFailureCode(),
		FailureMessage: info.GetFailureMessage(),
	}
}

type swapInfo struct {
	ports.SwapRequest
}

func (i swapInfo) toProto() *daemonv2.SwapInfo {
	info := i.SwapRequest
	if info == nil {
		return nil
	}

	return &daemonv2.SwapInfo{
		AssetP:  info.GetAssetP(),
		AmountP: info.GetAmountP(),
		AssetR:  info.GetAssetR(),
		AmountR: info.GetAmountR(),
	}
}

type failInfo struct {
	ports.SwapFail
}

func (i failInfo) toProto() *daemonv2.SwapFailInfo {
	if i.SwapFail == nil {
		return nil
	}

	return &daemonv2.SwapFailInfo{
		FailureCode:    i.SwapFail.GetFailureCode(),
		FailureMessage: i.SwapFail.GetFailureMessage(),
	}
}

type utxosInfo []ports.Utxo

func (i utxosInfo) toProto() []*daemonv2.UtxoInfo {
	list := make([]*daemonv2.UtxoInfo, 0, len(i))
	for _, info := range i {
		list = append(list, utxoInfo{info}.toProto())
	}
	return list
}

type utxoInfo struct {
	ports.Utxo
}

func (i utxoInfo) toProto() *daemonv2.UtxoInfo {
	return &daemonv2.UtxoInfo{
		Outpoint: &daemonv2.Outpoint{
			Hash:  i.Utxo.GetTxid(),
			Index: i.Utxo.GetIndex(),
		},
		Asset: i.Utxo.GetAsset(),
		Value: i.Utxo.GetValue(),
	}
}

type hooksInfo []ports.WebhookInfo

func (i hooksInfo) toProto() []*daemonv2.WebhookInfo {
	list := make([]*daemonv2.WebhookInfo, 0, len(i))
	for _, info := range i {
		list = append(list, &daemonv2.WebhookInfo{
			Id:         info.GetId(),
			Endpoint:   info.GetEndpoint(),
			IsSecured:  info.IsSecured(),
			ActionType: daemonv2.ActionType(info.GetActionType()),
		})
	}
	return list
}

type webhookInfo struct {
	*daemonv2.AddWebhookRequest
}

func (i webhookInfo) GetActionType() int {
	return int(i.AddWebhookRequest.Action)
}

type depositsInfo []ports.Deposit

func (i depositsInfo) toProto() []*daemonv2.Transaction {
	list := make([]*daemonv2.Transaction, 0, len(i))
	for _, info := range i {
		list = append(list, &daemonv2.Transaction{
			Txid:                info.GetTxid(),
			TotalAmountPerAsset: info.GetTotalAmountPerAsset(),
			Timestamp:           info.GetTimestamp(),
			Date:                timestampToString(info.GetTimestamp()),
		})
	}
	return list
}

type withdrawalsInfo []ports.Withdrawal

func (i withdrawalsInfo) toProto() []*daemonv2.Transaction {
	list := make([]*daemonv2.Transaction, 0, len(i))
	for _, info := range i {
		list = append(list, &daemonv2.Transaction{
			Txid:                info.GetTxid(),
			TotalAmountPerAsset: info.GetTotalAmountPerAsset(),
			Timestamp:           info.GetTimestamp(),
			Date:                timestampToString(info.GetTimestamp()),
		})
	}
	return list
}

type timeRangeInfo struct {
	*daemonv2.TimeRange
}

func (i timeRangeInfo) GetPredefinedPeriod() ports.PredefinedPeriod {
	return predefinedPeriodInfo(i.TimeRange.GetPredefinedPeriod())
}

func (i timeRangeInfo) GetCustomPeriod() ports.CustomPeriod {
	return i.TimeRange.GetCustomPeriod()
}

type predefinedPeriodInfo daemonv2.PredefinedPeriod

func (i predefinedPeriodInfo) IsLastHour() bool {
	return daemonv2.PredefinedPeriod(i) == daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_HOUR
}
func (i predefinedPeriodInfo) IsLastDay() bool {
	return daemonv2.PredefinedPeriod(i) == daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_DAY
}
func (i predefinedPeriodInfo) IsLastWeek() bool {
	return daemonv2.PredefinedPeriod(i) == daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_WEEK
}
func (i predefinedPeriodInfo) IsLastMonth() bool {
	return daemonv2.PredefinedPeriod(i) == daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_MONTH
}
func (i predefinedPeriodInfo) IsLastThreeMonths() bool {
	return daemonv2.PredefinedPeriod(i) == daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_THREE_MONTHS
}
func (i predefinedPeriodInfo) IsYearToDate() bool {
	return daemonv2.PredefinedPeriod(i) == daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_YEAR_TO_DATE
}
func (i predefinedPeriodInfo) IsLastYear() bool {
	return daemonv2.PredefinedPeriod(i) == daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_YEAR
}
func (i predefinedPeriodInfo) IsAll() bool {
	return daemonv2.PredefinedPeriod(i) == daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_ALL
}

type walletInfo struct {
	ports.WalletInfo
	ports.BuildData
}

func (i walletInfo) toProto() *daemonv2.GetInfoResponse {
	info := i.WalletInfo
	return &daemonv2.GetInfoResponse{
		RootPath:          info.GetRootPath(),
		MasterBlindingKey: info.GetMasterBlindingKey(),
		Network:           info.GetNetwork(),
		BuildData:         buildDataInfo{i.BuildData}.toProto(),
		AccountInfo:       accountsInfo(info.GetAccounts()).toProto(),
	}
}

type walletStatusInfo struct {
	ports.WalletStatus
}

func (i walletStatusInfo) toProto() *daemonv2.GetStatusResponse {
	info := i.WalletStatus
	return &daemonv2.GetStatusResponse{
		Initialized: info.IsInitialized(),
		Unlocked:    info.IsUnlocked(),
		Synced:      info.IsSynced(),
	}
}

type buildDataInfo struct {
	ports.BuildData
}

func (i buildDataInfo) toProto() *daemonv2.BuildInfo {
	info := i.BuildData
	return &daemonv2.BuildInfo{
		Version: info.GetVersion(),
		Commit:  info.GetCommit(),
		Date:    info.GetDate(),
	}
}

type accountsInfo []ports.WalletAccount

func (i accountsInfo) toProto() []*daemonv2.AccountInfo {
	list := make([]*daemonv2.AccountInfo, 0, len(i))
	for _, account := range i {
		list = append(list, &daemonv2.AccountInfo{
			AccountName:    account.GetName(),
			DerivationPath: account.GetDerivationPath(),
			Xpub:           account.GetXpub(),
		})
	}
	return list
}

type pageInfo struct {
	*daemonv2.Page
}

func (i pageInfo) GetNumber() int64 {
	if i.Page.GetNumber() <= 0 {
		return 1
	}
	return i.Page.GetNumber()
}

func (i pageInfo) GetSize() int64 {
	if i.Page.GetSize() <= 0 {
		return 1
	}
	return i.Page.GetSize()
}
