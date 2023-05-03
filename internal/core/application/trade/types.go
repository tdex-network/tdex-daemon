package trade

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type marketList []domain.Market

func (l marketList) toPortableList() []ports.MarketInfo {
	list := make([]ports.MarketInfo, 0, len(l))
	for _, m := range l {
		list = append(list, marketInfo{m, nil})
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
	return i.Market.Price
}
func (i marketInfo) GetBalance() map[string]ports.Balance {
	return i.balance
}

type marketStrategyInfo int

func (i marketStrategyInfo) IsBalanced() bool {
	return i == domain.StrategyTypeBalanced
}
func (i marketStrategyInfo) IsPluggable() bool {
	return i == domain.StrategyTypePluggable
}

type previewInfo struct {
	domain.Market
	domain.PreviewInfo
}

func (i previewInfo) GetMarketPercentageFee() ports.MarketFee {
	return marketFeeInfo(i.Market.PercentageFee)
}
func (i previewInfo) GetMarketFixedFee() ports.MarketFee {
	return marketFeeInfo(i.Market.FixedFee)
}
func (i previewInfo) GetAmount() uint64 {
	return i.PreviewInfo.Amount
}
func (i previewInfo) GetAsset() string {
	return i.PreviewInfo.Asset
}
func (i previewInfo) GetMarketPrice() ports.MarketPrice {
	return i.Market.Price
}
func (i previewInfo) GetFeeAmount() uint64 {
	return i.PreviewInfo.FeeAmount
}
func (i previewInfo) GetFeeAsset() string {
	return i.PreviewInfo.FeeAsset
}

type tradeTypeInfo struct {
	ports.TradeType
}

func (i tradeTypeInfo) toDomain() domain.TradeType {
	if i.TradeType.IsBuy() {
		return domain.TradeBuy
	}
	return domain.TradeSell
}

type swapRequestInfo struct {
	ports.SwapRequest
}

func (i swapRequestInfo) toDomain() domain.SwapRequest {
	info := i.SwapRequest
	list := make([]domain.UnblindedInput, 0, len(info.GetUnblindedInputs()))
	for _, in := range info.GetUnblindedInputs() {
		list = append(list, domain.UnblindedInput{
			Index:         in.GetIndex(),
			Asset:         in.GetAsset(),
			Amount:        in.GetAmount(),
			AssetBlinder:  in.GetAssetBlinder(),
			AmountBlinder: in.GetAmountBlinder(),
		})
	}
	return domain.SwapRequest{
		Id:              info.GetId(),
		AssetP:          info.GetAssetP(),
		AssetR:          info.GetAssetR(),
		AmountP:         info.GetAmountP(),
		AmountR:         info.GetAmountR(),
		Transaction:     info.GetTransaction(),
		UnblindedInputs: list,
		FeeAsset:        info.GetFeeAsset(),
		FeeAmount:       info.GetFeeAmount(),
	}
}

type swapAcceptInfo struct {
	*domain.SwapAccept
}

func (i swapAcceptInfo) GetUnblindedInputs() []ports.UnblindedInput {
	info := i.SwapAccept
	list := make([]ports.UnblindedInput, 0, len(info.GetUnblindedInputs()))
	for _, in := range info.GetUnblindedInputs() {
		list = append(list, in)
	}
	return list
}

type utxo struct {
	txid  string
	index uint32
}

func (u utxo) GetTxid() string {
	return u.txid
}
func (u utxo) GetIndex() uint32 {
	return u.index
}
func (u utxo) GetAsset() string {
	return ""
}
func (u utxo) GetValue() uint64 {
	return 0
}
func (u utxo) GetScript() string {
	return ""
}
func (u utxo) GetAssetBlinder() string {
	return ""
}
func (u utxo) GetValueBlinder() string {
	return ""
}
func (u utxo) GetSpentStatus() ports.UtxoStatus {
	return nil
}
func (u utxo) GetConfirmedStatus() ports.UtxoStatus {
	return nil
}
func (u utxo) GetRedeemScript() string {
	return ""
}
