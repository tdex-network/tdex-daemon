package v0domain

import "github.com/google/uuid"

type Trade struct {
	ID                  uuid.UUID
	MarketBaseAsset     string
	MarketQuoteAsset    string
	MarketPrice         Prices
	MarketFee           int64
	MarketFixedBaseFee  int64
	MarketFixedQuoteFee int64
	TraderPubkey        []byte
	Status              Status
	PsetBase64          string
	TxID                string
	TxHex               string
	ExpiryTime          uint64
	SettlementTime      uint64
	SwapRequest         Swap
	SwapAccept          Swap
	SwapComplete        Swap
	SwapFail            Swap
}

type Status struct {
	Code   int
	Failed bool
}

type Swap struct {
	ID        string
	Message   []byte
	Timestamp uint64
}
