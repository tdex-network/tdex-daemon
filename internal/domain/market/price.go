package market

import (
	"sort"
	"time"
)

//Price ...
type Price float32

//PriceByTime ...
type PriceByTime map[uint64]Price

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

// ChangeBasePrice ...
func (m *Market) ChangeBasePrice(price float32) error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	// TODO add logic to be sure that the price do not change to much from the latest one

	timestamp := uint64(time.Now().Unix())
	if _, ok := m.basePrice[timestamp]; ok {
		return ErrPriceExists
	}

	m.basePrice[timestamp] = Price(price)
	return nil
}

// ChangeQuotePrice ...
func (m *Market) ChangeQuotePrice(price float32) error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	//TODO check if the previous price is changing too much as security measure

	timestamp := uint64(time.Now().Unix())
	if _, ok := m.quotePrice[timestamp]; ok {
		return ErrPriceExists
	}

	m.quotePrice[timestamp] = Price(price)
	return nil
}

// IsZero ...
func (pt PriceByTime) IsZero() bool {
	return len(pt) == 0
}

// IsZero ...
func (p Price) IsZero() bool {
	return p == Price(0)
}

func getLatestPrice(keyValue PriceByTime) (uint64, Price) {
	if keyValue.IsZero() {
		return uint64(0), Price(0)
	}

	keys := make([]uint64, 0, len(keyValue))
	for k := range keyValue {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	latestKey := keys[len(keys)-1]
	latestValue := keyValue[latestKey]
	return latestKey, latestValue
}
