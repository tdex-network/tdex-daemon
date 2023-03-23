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

type marketInfo struct {
	domain.Market
	balance map[string]ports.Balance
}

func (i marketInfo) Ticker() string {
	//TODO implement me
	panic("implement me")
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
	balance map[string]ports.Balance
}

func (i previewInfo) GetPercentageFee() uint32 {
	return i.Market.PercentageFee
}
func (i previewInfo) GetFixedBaseFee() uint64 {
	return i.Market.FixedFee.BaseFee
}
func (i previewInfo) GetFixedQuoteFee() uint64 {
	return i.Market.FixedFee.QuoteFee
}
func (i previewInfo) GetAmount() uint64 {
	return i.PreviewInfo.Amount
}
func (i previewInfo) GetAsset() string {
	return i.PreviewInfo.Asset
}
func (i previewInfo) GetMarketBalance() map[string]ports.Balance {
	return i.balance
}
func (i previewInfo) GetMarketFee() ports.MarketFee {
	return i
}
func (i previewInfo) GetMarketPrice() ports.MarketPrice {
	return i.Market.Price
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
