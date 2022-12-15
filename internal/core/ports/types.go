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
	GetScript() []byte
	GetAssetBlinder() []byte
	GetValueBlinder() []byte
	GetConfirmedStatus() UtxoStatus
	GetSpentStatus() UtxoStatus
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
	GetPercentageFee() uint32
	GetFixedBaseFee() uint64
	GetFixedQuoteFee() uint64
}

type MarketStartegy interface {
	IsBalanced() bool
	IsPluggable() bool
}

type MarketInfo interface {
	GetMarket() Market
	GetAccountName() string
	IsTradable() bool
	GetStrategyType() MarketStartegy
	GetBalance() map[string]Balance
	GetFee() MarketFee
	GetPrice() MarketPrice
}

type TradeType interface {
	IsBuy() bool
	IsSell() bool
}

type Trade interface {
	GetId() string
	GetStatus() TradeStatus
	GetSwapInfo() SwapRequest
	GetSwapFailInfo() SwapFail
	GetRequestTimestamp() int64
	GetAcceptTimestamp() int64
	GetCompleteTimestamp() int64
	GetSettleTimestamp() int64
	GetExpiryTimestamp() int64
	GetMarket() Market
	GetMarketFee() MarketFee
	GetMarketPrice() MarketPrice
}

type TradePreview interface {
	GetAmount() uint64
	GetAsset() string
	GetMarketFee() MarketFee
	GetMarketPrice() MarketPrice
	GetMarketBalance() map[string]Balance
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
	GetActionType() int
	GetEndpoint() string
	GetSecret() string
}

type WebhookInfo interface {
	GetId() string
	GetActionType() int
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
