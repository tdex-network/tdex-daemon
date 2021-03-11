package esplora

import (
	"fmt"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type esplora struct {
	apiURL string
}

// NewService returns a new esplora service as an explorer.Service interface
func NewService(apiURL string) (explorer.Service, error) {
	service := &esplora{apiURL}

	if _, err := service.GetBlockHeight(); err != nil {
		return nil, fmt.Errorf("health check: %w", err)
	}

	return service, nil
}
