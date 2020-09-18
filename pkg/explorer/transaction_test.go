package explorer

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/magiconair/properties/assert"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"testing"
	"time"
)

func TestGetTransactionStatus(t *testing.T) {
	privkey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Fatal(err)
	}
	pubkey := privkey.PubKey()
	p2wpkh := payment.FromPublicKey(pubkey, &network.Regtest, nil)
	address, _ := p2wpkh.WitnessPubKeyHash()

	// Fund sender address.
	txID, err := Faucet(address)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	explorerSvc := NewService()

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
