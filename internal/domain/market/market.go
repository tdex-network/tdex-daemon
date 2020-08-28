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
	// Each Market has a different fee expressed in basis point of each swap
	fee int64
	// The asset hash should be used to take a cut
	feeAsset string
	// if curretly open for trades
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
	// Default asset fee is the base asset
	defaultFeeAsset := config.GetString(config.BaseAssetKey)

	return &Market{
		accountIndex: positiveAccountIndex,
		baseAsset:    &depositedAsset{},
		quoteAsset:   &depositedAsset{},
		fee:          defaultFeeInBasisPoint,
		feeAsset:     defaultFeeAsset,
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

// FeeAsset returns the selected asset to be used for market fee collection
func (m *Market) FeeAsset() string {
	return m.feeAsset
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

// ChangeFee ...
func (m *Market) ChangeFee(fee int64) error {

	if !m.IsFunded() {
		return errors.New("Market must be funded before change the fee")
	}

	if m.IsTradable() {
		return errors.New("Cannot change the fee when market is open for trading")
	}

	if err := validateFee(fee); err != nil {
		return err
	}

	m.fee = fee
	return nil
}

// ChangeFeeAsset ...
func (m *Market) ChangeFeeAsset(asset string) error {
	// In case of empty asset hash, no updates happens and therefore it exit without error
	if asset == "" {
		return nil
	}

	if !m.IsFunded() {
		return errors.New("Market must be funded before change the fee asset")
	}

	if m.IsTradable() {
		return errors.New("Cannot change the fee asset when market is open for trading")
	}

	if asset != m.BaseAssetHash() && asset != m.QuoteAssetHash() {
		return errors.New("The given asset must be either the base or quote asset in the pair")
	}

	m.feeAsset = asset
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

func validateFee(basisPoint int64) error {
	if basisPoint < 1 || basisPoint > 9999 {
		return errors.New("percentage of the fee on each swap must be > 0.01 and < 99")
	}

	return nil
}
