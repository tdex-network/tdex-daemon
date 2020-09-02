package market

import (
	"errors"

	"github.com/tdex-network/tdex-daemon/config"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
)

var (
	//ErrNotFunded is thrown when a market requires being funded for a change
	ErrNotFunded = errors.New("market must be funded")
	//ErrNotTradable is thrown when a market requires being tradable for a change
	ErrNotTradable = errors.New("market must be tradable")
	//ErrPriceExists is thrown when a price for that given timestamp already exists
	ErrPriceExists = errors.New("price has been inserted already")
	//ErrNotPriced is thrown when the price is still 0 (ie. not initialized)
	ErrNotPriced = errors.New("price must be inserted")
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
	// Market Making strategy
	strategy mm.MakingStrategy
	// how much 1 base asset is valued in quote asset.
	// It's a map  timestamp -> price, so it's easier to do historical price change.
	basePrice PriceByTime
	// how much 1 quote asset is valued in base asset
	// It's a map  timestamp -> price, so it's easier to do historical price change.
	quotePrice PriceByTime
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

		basePrice:  map[uint64]Price{},
		quotePrice: map[uint64]Price{},

		fee:      defaultFeeInBasisPoint,
		feeAsset: defaultFeeAsset,

		tradable: false,

		strategy: mm.MakingStrategy{},
	}, nil
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

// BaseAssetPrice returns the latest price for the base asset
func (m *Market) BaseAssetPrice() float32 {
	_, price := getLatestPrice(m.basePrice)

	return float32(price)
}

// QuoteAssetPrice returns the latest price for the quote asset
func (m *Market) QuoteAssetPrice() float32 {
	_, price := getLatestPrice(m.quotePrice)

	return float32(price)
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
		return ErrNotFunded
	}

	if m.strategy.IsZero() && (m.basePrice.IsZero() || m.quotePrice.IsZero()) {
		return ErrNotPriced
	}

	m.tradable = true
	return nil
}

// MakeNotTradable ...
func (m *Market) MakeNotTradable() error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	m.tradable = false
	return nil
}

// ChangeFee ...
func (m *Market) ChangeFee(fee int64) error {

	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsTradable() {
		return ErrNotTradable
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
		return ErrNotFunded
	}

	if m.IsTradable() {
		return ErrNotTradable
	}

	if asset != m.BaseAssetHash() && asset != m.QuoteAssetHash() {
		return errors.New("The given asset must be either the base or quote asset in the pair")
	}

	m.feeAsset = asset
	return nil
}

// Strategy ...
func (m *Market) Strategy() mm.MakingStrategy {
	return m.strategy
}

// ChangeStrategyWith

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

// IsFunded method returns true if the market contains a non empty funding tx outpoint for each asset
func (m *Market) IsFunded() bool {
	return m.baseAsset.IsNotZero() && m.quoteAsset.IsNotZero()
}

// IsTradable returns true if the market is available for trading
func (m *Market) IsTradable() bool {
	return m.tradable
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
