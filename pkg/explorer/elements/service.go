package elements

import (
	"encoding/json"
	"fmt"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type elements struct {
	client *RPCClient
}

// NewService returns and Elements based explorer
func NewService(host string, port int, user, passwd string) (explorer.Service, error) {
	client, err := NewClient(host, port, user, passwd, false, 30)
	if err != nil {
		return nil, err
	}

	return &elements{client}, nil
}

func (e *elements) mine() (string, error) {
	r, err := e.client.call("getnewaddress", nil)
	if err = handleError(err, &r); err != nil {
		return "", fmt.Errorf("info: %w", err)
	}
	var addr string
	if err := json.Unmarshal(r.Result, &addr); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}

	r, err = e.client.call("generatetoaddress", []interface{}{1, addr})
	if err = handleError(err, &r); err != nil {
		return "", fmt.Errorf("info: %w", err)
	}

	var blockHash []string
	if err := json.Unmarshal(r.Result, &blockHash); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}
	return blockHash[0], nil
}

func (e *elements) importAddress(addr, label string) error {
	r, err := e.client.call("importaddress", []interface{}{addr, label, false})
	if err = handleError(err, &r); err != nil {
		return err
	}
	return nil
}

func (e *elements) importBlindKey(addr, blindKey string) error {
	r, err := e.client.call("importblindingkey", []interface{}{addr, blindKey})
	if err = handleError(err, &r); err != nil {
		return err
	}
	return nil
}

func (e *elements) isAddressImported(targetLabel string) (bool, error) {
	r, err := e.client.call("listlabels", nil)
	if err = handleError(err, &r); err != nil {
		return false, err
	}

	var labels []interface{}
	if err := json.Unmarshal(r.Result, &labels); err != nil {
		return false, fmt.Errorf("unmarshal: %w", err)
	}

	for _, label := range labels {
		if label.(string) == targetLabel {
			return true, nil
		}
	}
	return false, nil
}

// isBlindKeyImported returns whethet the blinding private key relative to an
// address has already been imported. It accomplishes that by checking if the
// `dumpblindingkey` RPC returns an error in its response.
func (e *elements) isBlindKeyImported(addr string) (bool, error) {
	r, err := e.client.call("dumpblindingkey", []interface{}{addr})
	if err != nil {
		return false, err
	}
	return r.Err == nil, nil
}
