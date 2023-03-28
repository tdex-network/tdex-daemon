package feeder

import "fmt"

var (
	ErrMarketNotPluggable = fmt.Errorf("market is not pluggable")
	ErrMarketPluggable    = fmt.Errorf("cant remove pluggable market")
	ErrFeedOn             = fmt.Errorf("feed needs to be stopped before updating it")
)
