package domain

import (
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/config"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
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
	strategy mm.MakingStrategy
	// how much 1 base asset is valued in quote asset.
	// It's a map  timestamp -> price, so it's easier to do historical price change.
	basePrice PriceByTime
	// how much 1 quote asset is valued in base asset
	// It's a map  timestamp -> price, so it's easier to do historical price change.
	quotePrice PriceByTime
}

// OutpointWithAsset contains the transaction outpoint (tx hash and vout) along with the asset hash
type OutpointWithAsset struct {
	Asset string
	Txid  string
	Vout  int
}

type depositedAsset struct {
	assetHash string
}

//Price ...
type Price decimal.Decimal

//PriceByTime ...
type PriceByTime map[uint64]Price

//Market strategy type
type StrategyType int32

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

//NewMarketFromFields is necessary because newly introduced badger persistent
//implementation is not able to work with domain Market object since fields
//are not exported
func NewMarketFromFields(
	accountIndex int,
	baseAsset string,
	quoteAsset string,
	fee int64,
	feeAsset string,
	tradable bool,
	formula mm.MakingFormula,
	basePrice map[uint64]decimal.Decimal,
	quotePrice map[uint64]decimal.Decimal,
) *Market {

	basePriceCopy := make(map[uint64]Price)
	for k, v := range basePrice {
		basePriceCopy[k] = Price(v)
	}

	quotePriceCopy := make(map[uint64]Price)
	for k, v := range quotePrice {
		quotePriceCopy[k] = Price(v)
	}

	return &Market{
		accountIndex: accountIndex,
		baseAsset: &depositedAsset{
			assetHash: baseAsset,
		},
		quoteAsset: &depositedAsset{
			assetHash: quoteAsset,
		},
		fee:        fee,
		feeAsset:   feeAsset,
		tradable:   tradable,
		strategy:   mm.NewStrategyFromFormula(formula),
		basePrice:  basePriceCopy,
		quotePrice: quotePriceCopy,
	}
}
