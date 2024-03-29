package ports

import "github.com/shopspring/decimal"

type Balance interface {
	GetConfirmedBalance() uint64
	GetUnconfirmedBalance() uint64
	GetLockedBalance() uint64
	GetTotalBalance() uint64
}

type UtxoKey interface {
	GetTxid() string
	GetIndex() uint32
}

type UtxoStatus interface {
	GetTxid() string
	GetBlockInfo() BlockInfo
}

type Utxo interface {
	UtxoKey
	GetAsset() string
	GetValue() uint64
	GetScript() string
	GetAssetBlinder() string
	GetValueBlinder() string
	GetConfirmedStatus() UtxoStatus
	GetSpentStatus() UtxoStatus
	GetRedeemScript() string
}

type UnblindedInput interface {
	GetIndex() uint32
	GetAsset() string
	GetAmount() uint64
	GetAssetBlinder() string
	GetAmountBlinder() string
}

type TxInput interface {
	UtxoKey
	GetScript() string
	GetScriptSigSize() int
	GetWitnessSize() int
}

type TxOutput interface {
	GetAsset() string
	GetAmount() uint64
	GetScript() string
	GetBlindingKey() string
}

type BlockInfo interface {
	GetHash() string
	GetHeight() uint64
	GetTimestamp() int64
}

type Market interface {
	GetBaseAsset() string
	GetQuoteAsset() string
}

type MarketPrice interface {
	GetBasePrice() decimal.Decimal
	GetQuotePrice() decimal.Decimal
}

type MarketFee interface {
	GetBaseAsset() uint64
	GetQuoteAsset() uint64
}

type MarketStrategy interface {
	IsBalanced() bool
	IsPluggable() bool
}

type MarketInfo interface {
	GetMarket() Market
	GetName() string
	IsTradable() bool
	GetStrategyType() MarketStrategy
	GetBalance() map[string]Balance
	GetPercentageFee() MarketFee
	GetFixedFee() MarketFee
	GetPrice() MarketPrice
	GetBaseAssetPrecision() uint32
	GetQuoteAssetPrecision() uint32
}

type MarketReport interface {
	GetMarket() Market
	GetCollectedFees() MarketCollectedFees
	GetTotalVolume() MarketVolume
	GetVolumesPerFrame() []MarketVolume
}

type MarketCollectedFees interface {
	GetBaseAmount() uint64
	GetQuoteAmount() uint64
	GetTradeFeeInfo() []TradeFeeInfo
	GetStartDate() string
	GetEndDate() string
}

type MarketVolume interface {
	GetBaseVolume() uint64
	GetQuoteVolume() uint64
	GetStartDate() string
	GetEndDate() string
}

type TradeFeeInfo interface {
	GetTradeId() string
	GetPercentageFee() uint64
	GetFixedFee() uint64
	GetFeeAsset() string
	GetFeeAmount() uint64
	GetMarketPrice() decimal.Decimal
	GetTimestamp() int64
}

type TradeType interface {
	IsBuy() bool
	IsSell() bool
}

type Trade interface {
	GetId() string
	GetType() TradeType
	GetStatus() TradeStatus
	GetSwapInfo() SwapRequest
	GetSwapFailInfo() SwapFail
	GetRequestTimestamp() int64
	GetAcceptTimestamp() int64
	GetCompleteTimestamp() int64
	GetSettleTimestamp() int64
	GetExpiryTimestamp() int64
	GetMarket() Market
	GetMarketPercentageFee() MarketFee
	GetMarketFixedFee() MarketFee
	GetMarketPrice() MarketPrice
	GetTxid() string
	GetTxHex() string
	GetFeeAsset() string
	GetFeeAmount() uint64
}

type TradePreview interface {
	GetAmount() uint64
	GetAsset() string
	GetMarketPercentageFee() MarketFee
	GetMarketFixedFee() MarketFee
	GetMarketPrice() MarketPrice
	GetFeeAmount() uint64
	GetFeeAsset() string
}

type TimeRange interface {
	GetPredefinedPeriod() PredefinedPeriod
	GetCustomPeriod() CustomPeriod
}

type PredefinedPeriod interface {
	IsLastHour() bool
	IsLastDay() bool
	IsLastWeek() bool
	IsLastMonth() bool
	IsLastThreeMonths() bool
	IsYearToDate() bool
	IsLastYear() bool
	IsAll() bool
}

type CustomPeriod interface {
	GetStartDate() string
	GetEndDate() string
}

type Page interface {
	GetNumber() int64
	GetSize() int64
}

type TradeStatus interface {
	IsRequest() bool
	IsAccept() bool
	IsComplete() bool
	IsSettled() bool
	IsExpired() bool
	IsFailed() bool
}

type SwapRequest interface {
	GetId() string
	GetAssetP() string
	GetAmountP() uint64
	GetAssetR() string
	GetAmountR() uint64
	GetTransaction() string
	GetFeeAsset() string
	GetFeeAmount() uint64
	GetUnblindedInputs() []UnblindedInput
}

type SwapAccept interface {
	GetId() string
	GetRequestId() string
	GetTransaction() string
	GetUnblindedInputs() []UnblindedInput
}

type SwapComplete interface {
	GetId() string
	GetAcceptId() string
	GetTransaction() string
}

type SwapFail interface {
	GetId() string
	GetMessageId() string
	GetFailureCode() uint32
	GetFailureMessage() string
}

type Deposit interface {
	GetTxid() string
	GetTotalAmountPerAsset() map[string]uint64
	GetTimestamp() int64
}

type Withdrawal interface {
	GetTxid() string
	GetTotalAmountPerAsset() map[string]uint64
	GetTimestamp() int64
}

type Webhook interface {
	GetId() string
	GetEvent() WebhookEvent
	GetEndpoint() string
	GetSecret() string
}

type WebhookEvent interface {
	IsUnspecified() bool
	IsTradeSettled() bool
	IsAccountLowBalance() bool
	IsAccountWithdraw() bool
	IsAccountDeposit() bool
	IsAny() bool
}

type WebhookInfo interface {
	GetId() string
	GetEvent() WebhookEvent
	GetEndpoint() string
	IsSecured() bool
}

type FragmenterReply interface {
	GetMessage() string
	GetError() error
}

type BuildData interface {
	GetVersion() string
	GetCommit() string
	GetDate() string
}

type PriceFeedInfo interface {
	GetId() string
	GetMarket() Market
	GetSource() string
	GetTicker() string
	IsStarted() bool
}

type PriceFeed interface {
	GetMarket() Market
	GetPrice() MarketPrice
}
