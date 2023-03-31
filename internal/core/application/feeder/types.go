package feeder

import "github.com/tdex-network/tdex-daemon/internal/core/domain"

type marketInfo domain.Market

func (i marketInfo) GetBaseAsset() string {
	return i.BaseAsset
}
func (i marketInfo) GetQuoteAsset() string {
	return i.QuoteAsset
}
