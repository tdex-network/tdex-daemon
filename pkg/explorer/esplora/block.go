package esplora

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/tdex-network/tdex-daemon/pkg/httputil"
)

func (e *esplora) GetBlockHeight() (int, error) {
	url := fmt.Sprintf(
		"%v/blocks/tip/height",
		e.apiURL,
	)
	status, resp, err := httputil.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return -1, err
	}
	if status != http.StatusOK {
		return -1, fmt.Errorf(resp)
	}

	blockHeight, err := strconv.Atoi(resp)
	if err != nil {
		return -1, err
	}

	return blockHeight, nil
}
