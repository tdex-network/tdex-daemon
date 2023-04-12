package feeder

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type marketPrice struct {
	ports.MarketPrice
}

func (mp marketPrice) toDomain() domain.MarketPrice {
	return domain.MarketPrice{
		BasePrice:  mp.GetBasePrice().String(),
		QuotePrice: mp.GetQuotePrice().String(),
	}
}
