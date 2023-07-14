package v0domain

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/shopspring/decimal"
)

type Market struct {
	AccountIndex        int
	BaseAsset           string
	QuoteAsset          string
	BaseAssetPrecision  uint
	QuoteAssetPrecision uint
	Fee                 int64
	FixedFee            FixedFee
	Tradable            bool
	Strategy            MakingStrategy
	Price               Prices
}

func (m Market) AccountName() string {
	buf, _ := hex.DecodeString(m.BaseAsset + m.QuoteAsset)
	return hex.EncodeToString(btcutil.Hash160(buf))[:5]
}

type Prices struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}

type FixedFee struct {
	BaseFee  int64
	QuoteFee int64
}

type MakingStrategy struct {
	Type int
}
