package operator

import "errors"

var (
	ErrMarketNotExist    = errors.New("market does not exist")
	ErrInvalidTimeFormat = errors.New("fromTime must be valid RFC3339 format")
	ErrInvalidTimeFrame  = errors.New("timeFrame must be smaller than timePeriod")
	ErrInvalidTime       = errors.New("must be a valid time.Time")
)
