package priceupdater

type Market struct {
	BaseAsset    string
	QuoteAsset   string
	MarketTicker string
}

func (m *Market) GetBaseAsset() string {
	return m.BaseAsset
}

func (m *Market) GetQuoteAsset() string {
	return m.QuoteAsset
}

func (m *Market) Ticker() string {
	return m.MarketTicker
}
