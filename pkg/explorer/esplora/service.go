package esplora

import (
	"fmt"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

const (
	minRequestTimeout = 5 * time.Second
)

type esplora struct {
	apiURL string
	client *Client
}

// NewService returns a new esplora service as an explorer.Service interface
func NewService(apiURL string, requestTimeout int) (explorer.Service, error) {
	d := time.Duration(requestTimeout) * time.Millisecond
	if d < minRequestTimeout {
		return nil, fmt.Errorf("request timeout must be at least 5 seconds")
	}
	client := NewHTTPClient(d)
	service := &esplora{apiURL, client}

	if _, err := service.GetBlockHeight(); err != nil {
		return nil, fmt.Errorf("health check: %w", err)
	}

	return service, nil
}
