package marketmaking

//SpotPriceOpts defines the parameter neeede to calculate the spot price
type SpotPriceOpts struct {
	balanceIn, balanceOut, weightIn, weightOut, fee uint64
	// Defines id the fee should be cahrged on the way in (ie. on )
	chargeFeeOnTheWayIn bool
}

// MakingFormula defines the interface for implmenting the formula to derive the spot price
type MakingFormula interface {
	SpotPrice() (spotPriceOpts SpotPriceOpts)
	OutGivenIn(spotPriceOpts SpotPriceOpts, amountIn uint64)
	OutGivenOut(spotPriceOpts SpotPriceOpts, amountOn uint64)
}
