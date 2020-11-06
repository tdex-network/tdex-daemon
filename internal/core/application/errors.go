package application

import "errors"

// ErrCrawlerDoesNotObserveAddresses occurs when the crawler does not observe any address
var	ErrCrawlerDoesNotObserveAddresses = errors.New("crawler does not observe any addresses")