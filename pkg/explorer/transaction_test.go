package explorer

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/magiconair/properties/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
)

func TestGetTransactionStatus(t *testing.T) {
	privkey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Fatal(err)
	}
	pubkey := privkey.PubKey()
	p2wpkh := payment.FromPublicKey(pubkey, &network.Regtest, nil)
	address, _ := p2wpkh.WitnessPubKeyHash()

	explorerSvc := NewService(config.GetString(config.ExplorerEndpointKey))

	// Fund sender address.
	txID, err := explorerSvc.Faucet(address)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	trxStatus, err := explorerSvc.GetTransactionStatus(txID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, trxStatus["confirmed"], true)

	isConfirmed, err := explorerSvc.IsTransactionConfirmed(txID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, isConfirmed, true)
}

func TestGetTransactionsForAddress(t *testing.T) {
	privkey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Fatal(err)
	}
	pubkey := privkey.PubKey()
	p2wpkh := payment.FromPublicKey(pubkey, &network.Regtest, nil)
	address, _ := p2wpkh.WitnessPubKeyHash()

	explorerSvc := NewService(config.GetString(config.ExplorerEndpointKey))

	// Fund sender address.
	if _, err := explorerSvc.Faucet(address); err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	txs, err := explorerSvc.GetTransactionsForAddress(address)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(txs))
}
