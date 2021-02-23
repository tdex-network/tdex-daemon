package trade

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
)

var explorerSvc explorer.Service

func initExplorer() error {
	var err error
	explorerSvc, err = esplora.NewService("http://localhost:3001")
	return err
}

func TestNewTrade(t *testing.T) {
	if err := initExplorer(); err != nil {
		t.Fatal(err)
	}

	client, err := tradeclient.NewTradeClient("localhost", 9000)
	if err != nil {
		t.Fatal(err)
	}

	tt, err := NewTrade(NewTradeOpts{
		Chain:           "regtest",
		ExplorerService: explorerSvc,
		Client:          client,
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
				Chain:           "",
				ExplorerService: explorerSvc,
				Client:          client,
			},
			err: ErrInvalidChain,
		},
		{
			opts: NewTradeOpts{
				Chain:           "bitcoin",
				ExplorerService: explorerSvc,
				Client:          client,
			},
			err: ErrInvalidChain,
		},
		{
			opts: NewTradeOpts{
				Chain:           "regtest",
				ExplorerService: nil,
				Client:          client,
			},
			err: ErrNullExplorer,
		},
		{
			opts: NewTradeOpts{
				Chain:           "regtest",
				ExplorerService: explorerSvc,
				Client:          nil,
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
		Chain:           "regtest",
		ExplorerService: explorerSvc,
		Client:          client,
	})
	return
}
