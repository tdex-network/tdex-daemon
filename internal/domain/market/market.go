package market

import (
	"errors"

	"github.com/tdex-network/tdex-daemon/config"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

var (
	//ErrNotFunded is thrown when a market requires being funded for a change
	ErrNotFunded = errors.New("market must be funded")
	//ErrNotTradable is thrown when a market requires being tradable for a change
	ErrNotTradable = errors.New("market must be opened")
	//ErrTradable is thrown when a market requires being NOT tradable for a change
	ErrTradable = errors.New("market must be closed")
	//ErrPriceExists is thrown when a price for that given timestamp already exists
	ErrPriceExists = errors.New("price has been inserted already")
	//ErrNotPriced is thrown when the price is still 0 (ie. not initialized)
	ErrNotPriced = errors.New("price must be inserted")
)

//Market defines the Market entity data structure for holding an asset pair state
type Market struct {
	// AccountIndex links a market to a HD wallet account derivation.
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
	strategy *mm.MakingStrategy
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

		strategy: mm.NewStrategyFromFormula(formula.BalancedReserves{}),
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
	println(m.IsFunded())
	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsStrategyPluggable() && !m.IsStrategyPluggableInitialized() {
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
