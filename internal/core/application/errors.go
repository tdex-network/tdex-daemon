package application

import "errors"

// ErrCrawlerDoesNotObserveAddresses occurs when the crawler does not observe any address
var	ErrCrawlerDoesNotObserveFeeAccount = errors.New("fee account needs to be funded to open a market")
