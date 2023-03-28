package pricefeederinfra

import (
	"fmt"
	"regexp"

	"github.com/shopspring/decimal"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

const (
	marketTickerFormat = "%s/%s"
)

type Feed struct {
	Market Market
	Price  Price
}

func (f Feed) GetMarket() ports.Market {
	return f.Market
}

func (f Feed) GetPrice() ports.MarketPrice {
	return f.Price
}

type Price struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}

func (p Price) GetBasePrice() decimal.Decimal {
	return p.BasePrice
}

func (p Price) GetQuotePrice() decimal.Decimal {
	return p.QuotePrice
}

type AddPriceFeedReq ports.AddPriceFeedReq

func (a AddPriceFeedReq) Validate() error {
	if len(a.MarketBaseAsset) != 32 {
		return fmt.Errorf(
			"invalid baseAsset: %s, must be 32 length", a.MarketBaseAsset,
		)
	}

	if len(a.MarketQuoteAsset) != 32 {
		return fmt.Errorf(
			"invalid quoteAsset: %s, must be 32 length",
			a.MarketQuoteAsset,
		)
	}

	if _, ok := sources[a.Source]; !ok {
		return fmt.Errorf(
			"invalid source: %s, must be one of %v", a.Source, sources,
		)
	}

	regex := regexp.MustCompile(marketTickerFormat)
	if !regex.MatchString(a.Ticker) {
		return fmt.Errorf(
			"invalid ticker: %s, must be in format %s",
			a.Ticker,
			marketTickerFormat,
		)
	}

	return nil
}

type UpdatePriceFeedReq ports.UpdatePriceFeedReq

func (u UpdatePriceFeedReq) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("id must not be empty")
	}

	if _, ok := sources[u.Source]; !ok {
		return fmt.Errorf(
			"invalid source: %s, must be one of %v", u.Source, sources,
		)
	}

	regex := regexp.MustCompile(marketTickerFormat)
	if !regex.MatchString(u.Ticker) {
		return fmt.Errorf(
			"invalid ticker: %s, must be in format %s",
			u.Ticker,
			marketTickerFormat,
		)
	}

	return nil
}
