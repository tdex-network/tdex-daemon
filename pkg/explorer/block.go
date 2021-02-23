package explorer

import (
	"fmt"
	"github.com/tdex-network/tdex-daemon/pkg/httputil"
	"net/http"
	"strconv"
)

func (e *explorer) GetBlockHeight() (int, error) {
	url := fmt.Sprintf(
		"%v/blocks/tip/height",
		e.apiUrl,
	)
	status, resp, err := httputil.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return 0, err
	}
	if status != http.StatusOK {
		return 0, fmt.Errorf(resp)
	}

	blockHeight, err := strconv.Atoi(resp)
	if err != nil {
		return 0, err
	}

	return blockHeight, nil
}
