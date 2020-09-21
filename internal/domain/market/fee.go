package market

import "errors"

// ChangeFee ...
func (m *Market) ChangeFee(fee int64) error {

	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsTradable() {
		return ErrMarketMustBeClose
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
		return ErrMarketMustBeClose
	}

	if asset != m.BaseAssetHash() && asset != m.QuoteAssetHash() {
		return errors.New("The given asset must be either the base or quote asset in the pair")
	}

	m.feeAsset = asset
	return nil
}
