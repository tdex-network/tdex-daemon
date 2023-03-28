package feeder

type marketInfo struct {
	baseAsset  string
	quoteAsset string
}

func (i marketInfo) GetBaseAsset() string {
	return i.baseAsset
}
func (i marketInfo) GetQuoteAsset() string {
	return i.quoteAsset
}
