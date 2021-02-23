package elements

import (
	"encoding/json"
	"fmt"
)

func (e *elements) GetBlockHeight() (int, error) {
	r, err := e.client.call("getblockcount", nil)
	if err = handleError(err, &r); err != nil {
		return -1, err
	}

	var height int
	if err := json.Unmarshal(r.Result, &height); err != nil {
		return -1, fmt.Errorf("unmarshal: %w", err)
	}
	return height, nil
}
