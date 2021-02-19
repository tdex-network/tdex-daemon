package esplora

import (
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type esplora struct {
	apiURL string
}

// NewService returns a new esplora service as an explorer.Service interface
func NewService(apiURL string) explorer.Service {
	return &esplora{apiURL}
}
