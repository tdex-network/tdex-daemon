package trade

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
	"github.com/vulpemventures/go-elements/network"
)

var (
	// ErrInvalidChain ...
	ErrInvalidChain = fmt.Errorf(
		"chain must be either '%s' or '%s", network.Liquid.Name, network.Regtest.Name,
	)
	// ErrInvalidProviderURL ...
	ErrInvalidProviderURL = fmt.Errorf(
		"provider url must be a valid url in the form 'host:port'",
	)
	// ErrInvalidExplorerURL ...
	ErrInvalidExplorerURL = fmt.Errorf("explorer url must be a valid url")
	// ErrNullClient ...
	ErrNullClient = fmt.Errorf("client must not be null")
)

type Trade struct {
	network  *network.Network
	explorer explorer.Service
	client   *tradeclient.Client
}

// NewTradeOpts is the struct given to NewTrade method
type NewTradeOpts struct {
	Chain       string
	ExplorerURL string
	Client      *tradeclient.Client
}

func (o NewTradeOpts) validate() error {
	if !isValidChain(o.Chain) {
		return ErrInvalidChain
	}
	if !isValidURL(o.ExplorerURL) {
		return ErrInvalidExplorerURL
	}
	if o.Client == nil {
		return ErrNullClient
	}
	return nil
}

// NewTrade returns a new trade initialized with the given arguments
func NewTrade(opts NewTradeOpts) (*Trade, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	return &Trade{
		network:  networkFromString(opts.Chain),
		explorer: explorer.NewService(opts.ExplorerURL),
		client:   opts.Client,
	}, nil
}

func isValidChain(chain string) bool {
	return chain == network.Liquid.Name || chain == network.Regtest.Name
}

func isValidURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil && strings.HasPrefix(urlStr, "http")
}

func networkFromString(chain string) *network.Network {
	if chain == network.Liquid.Name {
		return &network.Liquid
	}
	return &network.Regtest
}
