package market

// OutpointWithAsset contains the transaction outpoint (tx hash and vout) along with the asset hash
type OutpointWithAsset struct {
	Asset string
	Txid  string
	Vout  int
}

type depositedAsset struct {
	assetHash  string
	fundingTxs []OutpointWithAsset
}

// IsNotZero ...
func (d depositedAsset) IsNotZero() bool {
	return len(d.assetHash) > 0 && len(d.fundingTxs) > 0
}
