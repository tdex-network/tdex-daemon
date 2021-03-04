package elements

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type elements struct {
	client          *RPCClient
	rescanTimestamp interface{}
}

// NewService returns the Elements implementation of the Explorer interface.
// It establishes an insecure connection with the JSON-RPC interface of the
// node with no TLS termination.
func NewService(endpoint string, rescanTimestamp interface{}) (
	explorer.Service,
	error,
) {
	if err := validateEndpoint(endpoint); err != nil {
		return nil, err
	}
	if err := validateRescanTimestamp(rescanTimestamp); err != nil {
		return nil, err
	}

	parsedEndpoint, _ := url.Parse(endpoint)
	host := parsedEndpoint.Hostname()
	port, _ := strconv.Atoi(parsedEndpoint.Port())
	user := parsedEndpoint.User.Username()
	password, _ := parsedEndpoint.User.Password()
	if rescanTimestamp == nil {
		rescanTimestamp = "now"
	}

	client, err := NewClient(host, port, user, password, false, 30)

	if err != nil {
		return nil, err
	}

	service := &elements{client, rescanTimestamp}
	if _, err := service.GetBlockHeight(); err != nil {
		return nil, fmt.Errorf("healt check: %w", err)
	}
	return service, nil
}

func (e *elements) importAddress(addr, label string, rescan bool) error {
	rescanTime := e.rescanTimestamp
	if !rescan {
		rescanTime = "now"
	}
	r, err := e.client.call("importmulti", []interface{}{
		[]map[string]interface{}{
			{
				"scriptPubKey": map[string]string{
					"address": addr,
				},
				"watchonly": true,
				"label":     label,
				"timestamp": rescanTime,
			},
		},
		map[string]bool{
			"rescan": rescan,
		},
	})
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

// isBlindKeyImported returns whether the blinding private key relative to an
// address has already been imported. It accomplishes that by checking if the
// `dumpblindingkey` RPC returns an error in its response.
func (e *elements) isBlindKeyImported(addr string) (bool, error) {
	r, err := e.client.call("dumpblindingkey", []interface{}{addr})
	if err != nil {
		return false, err
	}
	return r.Err == nil, nil
}

/**** Regtest only ****/

// mine adds 1 block to the blockchain and returns its hash. It makes use of
// the getnewaddress RPC to derive a new address of the node's wallet and
// generatetoaddress to create 1 block and send the reward to the generated
// address.
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

func validateEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("missing endpoint")
	}
	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}

	if parsedEndpoint.Hostname() == "" {
		return ErrMissingRPCHost
	}
	if parsedEndpoint.Port() == "" {
		return ErrMissingRPCPort
	}
	if parsedEndpoint.User.Username() == "" {
		return ErrMissingRPCUser
	}
	if _, ok := parsedEndpoint.User.Password(); !ok {
		return ErrMissingRPCPassword
	}

	return nil
}

func validateRescanTimestamp(timestamp interface{}) error {
	if timestamp == nil {
		return nil
	}

	switch timestamp.(type) {
	case int:
		if timestamp.(int) < 0 {
			return ErrInvalidRescaTimestamp
		}
	default:
		return ErrInvalidRescaTimestamp
	}
	return nil
}
