package pricefeederinfra

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
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

func validateAddPriceFeed(market ports.Market, source, ticker string) error {
	if len(market.GetBaseAsset()) != 32 {
		return fmt.Errorf(
			"invalid baseAsset: %s, must be 32 length", market.GetBaseAsset(),
		)
	}

	if len(market.GetQuoteAsset()) != 32 {
		return fmt.Errorf(
			"invalid quoteAsset: %s, must be 32 length",
			market.GetQuoteAsset(),
		)
	}

	if _, ok := sources[source]; !ok {
		return fmt.Errorf(
			"invalid source: %s, must be one of %v", source, sources,
		)
	}

	return nil
}

func ValidateUpdatePriceFeed(id, source, ticker string) error {
	if id == "" {
		return fmt.Errorf("id must not be empty")
	}

	if _, ok := sources[source]; !ok {
		return fmt.Errorf(
			"invalid source: %s, must be one of %v", source, sources,
		)
	}

	return nil
}
