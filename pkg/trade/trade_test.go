package trade

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
)

func TestNewTrade(t *testing.T) {
	client, err := tradeclient.NewTradeClient("localhost", 9000)
	if err != nil {
		t.Fatal(err)
	}

	tt, err := NewTrade(NewTradeOpts{
		Chain:       "regtest",
		ExplorerURL: "http://localhost:3001",
		Client:      client,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, tt)
}

func TestFailingNewTrade(t *testing.T) {
	client, err := tradeclient.NewTradeClient("localhost", 9000)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		opts NewTradeOpts
		err  error
	}{
		{
			opts: NewTradeOpts{
				Chain:       "",
				ExplorerURL: "http://localhost:3001",
				Client:      client,
			},
			err: ErrInvalidChain,
		},
		{
			opts: NewTradeOpts{
				Chain:       "bitcoin",
				ExplorerURL: "http://localhost:3001",
				Client:      client,
			},
			err: ErrInvalidChain,
		},
		{
			opts: NewTradeOpts{
				Chain:       "regtest",
				ExplorerURL: "",
				Client:      client,
			},
			err: ErrInvalidExplorerURL,
		},
		{
			opts: NewTradeOpts{
				Chain:       "regtest",
				ExplorerURL: "localhost:3001",
				Client:      client,
			},
			err: ErrInvalidExplorerURL,
		},
		{
			opts: NewTradeOpts{
				Chain:       "regtest",
				ExplorerURL: "http://localhost:3001",
				Client:      nil,
			},
			err: ErrNullClient,
		},
	}

	for _, tt := range tests {
		_, err := NewTrade(tt.opts)
		assert.NotNil(t, err)
		assert.Equal(t, tt.err, err)
	}
}

var client *tradeclient.Client

func newTestTrade() (t *Trade, err error) {
	if client == nil {
		client, err = tradeclient.NewTradeClient("localhost", 9000)
		if err != nil {
			return
		}
	}

	t, err = NewTrade(NewTradeOpts{
		Chain:       "regtest",
		ExplorerURL: "http://localhost:3001",
		Client:      client,
	})
	return
}
