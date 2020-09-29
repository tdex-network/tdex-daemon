package trademarket

import (
	"encoding/hex"
	"errors"
)

type Market struct {
	BaseAsset  string
	QuoteAsset string
}

var (
	// ErrInvalidBaseAsset ...
	ErrInvalidBaseAsset = errors.New(
		"base asset must be a 32-byte array in hex format",
	)
	// ErrInvalidQuoteAsset ...
	ErrInvalidQuoteAsset = errors.New(
		"quote asset must be a 32-byte array in hex format",
	)
)

// Validate cheks whether the current market is well-formed
func (m *Market) Validate() error {
	if buf, err := hex.DecodeString(m.BaseAsset); err != nil || len(buf) != 32 {
		return ErrInvalidBaseAsset
	}
	if buf, err := hex.DecodeString(m.QuoteAsset); err != nil || len(buf) != 32 {
		return ErrInvalidQuoteAsset
	}
	return nil
}
