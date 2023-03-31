package pricefeederinfra

import (
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type PriceFeed struct {
	ID      string
	Market  Market
	Source  string
	Started bool
}

func (p PriceFeed) GetId() string {
	return p.ID
}

func (p PriceFeed) GetMarket() ports.Market {
	return p.Market
}

func (p PriceFeed) GetSource() string {
	return p.Source
}

func (p PriceFeed) GetTicker() string {
	return p.Market.Ticker
}

func (p PriceFeed) IsStarted() bool {
	return p.Started
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
		ID:      uuid.New().String(),
		Market:  market,
		Source:  source,
		Started: false,
	}
}
