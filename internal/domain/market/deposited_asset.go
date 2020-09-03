package market

import (
	"errors"

	"github.com/tdex-network/tdex-daemon/config"
)

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

// FundMarket adds funding details given an array of outpoints and recognize quote asset
func (m *Market) FundMarket(fundingTxs []OutpointWithAsset) error {
	var baseAssetHash string = config.GetString(config.BaseAssetKey)
	var otherAssetHash string

	var baseAssetTxs []OutpointWithAsset
	var otherAssetTxs []OutpointWithAsset

	assetCount := make(map[string]int)
	for _, o := range fundingTxs {
		assetCount[o.Asset]++
		if o.Asset == baseAssetHash {
			baseAssetTxs = append(baseAssetTxs, o)
		} else {
			// Potentially here could be different assets mixed
			// We chek if unique quote asset later after the loop
			otherAssetTxs = append(otherAssetTxs, o)
			otherAssetHash = o.Asset
		}
	}

	if _, ok := assetCount[baseAssetHash]; !ok {
		return errors.New("base asset is missing")
	}

	if keysNumber := len(assetCount); keysNumber != 2 {
		return errors.New("must be deposited 2 unique assets")
	}

	m.baseAsset = &depositedAsset{
		assetHash:  baseAssetHash,
		fundingTxs: baseAssetTxs,
	}

	m.quoteAsset = &depositedAsset{
		assetHash:  otherAssetHash,
		fundingTxs: otherAssetTxs,
	}

	return nil
}

// IsNotZero ...
func (d depositedAsset) IsNotZero() bool {
	return len(d.assetHash) > 0 && len(d.fundingTxs) > 0
}
