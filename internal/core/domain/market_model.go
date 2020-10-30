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
	AccountIndex int
	BaseAsset    string
	QuoteAsset   string
	// Each Market has a different fee expressed in basis point of each swap
	Fee int64
	// The asset hash should be used to take a cut
	FeeAsset string
	// if curretly open for trades
	Tradable bool
	// Market Making strategy
	Strategy mm.MakingStrategy
	// how much 1 base asset is valued in quote asset.
	BasePrice Price
	// how much 1 quote asset is valued in base asset
	QuotePrice Price
}

// OutpointWithAsset contains the transaction outpoint (tx hash and vout) along with the asset hash
type OutpointWithAsset struct {
	Asset string
	Txid  string
	Vout  int
}

//Price ...
type Price decimal.Decimal

//StrategyType is the Market making strategy type
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
		AccountIndex: positiveAccountIndex,
		BaseAsset:    "",
		QuoteAsset:   "",

		BasePrice:  Price{},
		QuotePrice: Price{},

		Fee:      defaultFeeInBasisPoint,
		FeeAsset: defaultFeeAsset,

		Tradable: false,

		Strategy: mm.NewStrategyFromFormula(formula.BalancedReserves{}),
	}, nil
}
