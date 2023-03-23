package coinbasefeeder

import (
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type price struct {
	basePrice  decimal.Decimal
	quotePrice decimal.Decimal
}

func (p *price) GetBasePrice() decimal.Decimal {
	return p.basePrice
}

func (p *price) GetQuotePrice() decimal.Decimal {
	return p.quotePrice
}

type priceFeed struct {
	market ports.Market
	price  *price
}

func (p *priceFeed) GetMarket() ports.Market {
	return p.market
}

func (p *priceFeed) GetPrice() ports.MarketPrice {
	return p.price
}

type market struct {
	baseAsset  string
	quoteAsset string
	ticker     string
}

func (m market) GetBaseAsset() string {
	return m.baseAsset
}

func (m market) GetQuoteAsset() string {
	return m.quoteAsset
}

func (m market) Ticker() string {
	return m.ticker
}
