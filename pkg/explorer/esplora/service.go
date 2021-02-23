package esplora

import (
	"fmt"
	"net/http"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/httputil"
)

type esplora struct {
	apiURL string
}

// NewService returns a new esplora service as an explorer.Service interface
func NewService(apiURL string) (explorer.Service, error) {
	service := &esplora{apiURL}

	if err := service.healtCheck(); err != nil {
		return nil, fmt.Errorf("health check: %w", err)
	}

	return service, nil
}

func (e *esplora) healtCheck() error {
	url := fmt.Sprintf("%s/blocks/tip/height", e.apiURL)
	status, resp, err := httputil.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf(resp)
	}
	return nil
}
