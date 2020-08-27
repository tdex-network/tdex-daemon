package market

import (
	"errors"

	"github.com/tdex-network/tdex-daemon/config"
)

//Market defines the Market entity data structure for holding an asset pair state
type Market struct {
	// AccountIndex links a market to a HD wallet account derivation.
	// Each Market could receive assets on any of those child addresses
	accountIndex int
	baseAsset    *depositedAsset
	quoteAsset   *depositedAsset
	// Each Market could have a different fee expressed in percentage of each swap
	fee      int64
	tradable bool
}

//NewMarket returns an empty market with a reference to an account index
func NewMarket(positiveAccountIndex int) (*Market, error) {

	if err := validateAccountIndex(positiveAccountIndex); err != nil {
		return nil, err
	}

	// Here we convert the float to integer indicating basis point to take from each swap
	defaultFeeInDecimals := config.GetFloat(config.DefaultFeeKey)
	defaultFeeInBasisPoint := int64(defaultFeeInDecimals * 100)

	return &Market{
		accountIndex: positiveAccountIndex,
		baseAsset:    &depositedAsset{},
		quoteAsset:   &depositedAsset{},
		fee:          defaultFeeInBasisPoint,
		tradable:     false,
	}, nil
}

// IsFunded method returns true if the market contains a non empty funding tx outpoint for each asset
func (m *Market) IsFunded() bool {
	return m.baseAsset.IsNotZero() && m.quoteAsset.IsNotZero()
}

// IsTradable returns true if the market is available for trading
func (m *Market) IsTradable() bool {
	return m.tradable
}

// AccountIndex returns the account index
func (m *Market) AccountIndex() int {
	return m.accountIndex
}

// BaseAssetHash returns the base asset hash
func (m *Market) BaseAssetHash() string {
	return m.baseAsset.assetHash
}

// QuoteAssetHash returns the quote asset hash
func (m *Market) QuoteAssetHash() string {
	return m.quoteAsset.assetHash
}

// Fee returns the selected fee
func (m *Market) Fee() int64 {
	return m.fee
}

// MakeTradable ...
func (m *Market) MakeTradable() error {
	if !m.IsFunded() {
		return errors.New("Market must be funded before change the status")
	}

	m.tradable = true
	return nil
}

// MakeNotTradable ...
func (m *Market) MakeNotTradable() error {
	if !m.IsFunded() {
		return errors.New("Market must be funded before change the trading status")
	}

	m.tradable = false
	return nil
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

type depositedAsset struct {
	assetHash  string
	fundingTxs []OutpointWithAsset
}

// OutpointWithAsset contains the transaction outpoint (tx hash and vout) along with the asset hash
type OutpointWithAsset struct {
	Asset string
	Txid  string
	Vout  int
}

// IsNotZero ...
func (d depositedAsset) IsNotZero() bool {
	return len(d.assetHash) > 0 && len(d.fundingTxs) > 0
}

func validateAccountIndex(accIndex int) error {
	if accIndex < 0 {
		return errors.New("Account index must be a positive integer number")
	}

	return nil
}
