package pricefeederinfra

import (
	"github.com/google/uuid"
)

type PriceFeed struct {
	ID     string
	Market Market
	Source string
	On     bool
}

type Market struct {
	BaseAsset  string
	QuoteAsset string
	Ticker     string
}

func (m Market) GetBaseAsset() string {
	return m.BaseAsset
}

func (m Market) GetQuoteAsset() string {
	return m.QuoteAsset
}

func NewPriceFeed(market Market, source string) PriceFeed {
	return PriceFeed{
		ID:     uuid.New().String(),
		Market: market,
		Source: source,
		On:     false,
	}
}
