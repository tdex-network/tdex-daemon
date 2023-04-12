package pricefeeder

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"
)

type PriceFeedInfo struct {
	ID      string
	Market  Market
	Source  string
	Ticker  string
	Started bool
}

func (p PriceFeedInfo) GetId() string {
	return p.ID
}

func (p PriceFeedInfo) GetMarket() ports.Market {
	return p.Market
}

func (p PriceFeedInfo) GetSource() string {
	return p.Source
}

func (p PriceFeedInfo) GetTicker() string {
	return p.Ticker
}

func (p PriceFeedInfo) IsStarted() bool {
	return p.Started
}

type Market struct {
	BaseAsset  string
	QuoteAsset string
}

func (m Market) GetBaseAsset() string {
	return m.BaseAsset
}

func (m Market) GetQuoteAsset() string {
	return m.QuoteAsset
}

func NewPriceFeedInfo(market ports.Market, source, ticker string) (*PriceFeedInfo, error) {
	if market == nil {
		return nil, fmt.Errorf("missing market")
	}
	if len(source) <= 0 {
		return nil, fmt.Errorf("missing price source")
	}
	if _, ok := feederFactory[source]; !ok {
		return nil, fmt.Errorf("unknown price source")
	}
	if len(ticker) <= 0 {
		return nil, fmt.Errorf("missing market ticker")
	}

	return &PriceFeedInfo{
		ID: uuid.New().String(),
		Market: Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
		},
		Source:  source,
		Ticker:  ticker,
		Started: false,
	}, nil
}

func (p *PriceFeedInfo) toMarketList() []pricefeeder.Market {
	return []pricefeeder.Market{
		{
			BaseAsset:  p.Market.BaseAsset,
			QuoteAsset: p.Market.QuoteAsset,
			Ticker:     p.Ticker,
		},
	}
}

type priceFeedInfo pricefeeder.PriceFeed

func (i priceFeedInfo) GetMarket() ports.Market {
	return Market{
		BaseAsset:  i.Market.BaseAsset,
		QuoteAsset: i.Market.QuoteAsset,
	}
}
func (i priceFeedInfo) GetBasePrice() decimal.Decimal {
	return i.Price.BasePrice
}
func (i priceFeedInfo) GetQuotePrice() decimal.Decimal {
	return i.Price.QuotePrice
}
func (i priceFeedInfo) GetPrice() ports.MarketPrice {
	return i
}
