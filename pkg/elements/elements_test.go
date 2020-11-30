package elements

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
)

var testKey []byte
var testAddress string

func newTestData() (string, []byte, error) {
	if len(testKey) > 0 && len(testAddress) > 0 {
		return testAddress, testKey, nil
	}

	key, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", nil, err
	}
	blindKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", nil, err
	}
	p2wpkh := payment.FromPublicKey(
		key.PubKey(),
		&network.Regtest,
		blindKey.PubKey(),
	)
	addr, err := p2wpkh.ConfidentialWitnessPubKeyHash()
	if err != nil {
		return "", nil, err
	}

	testAddress = addr
	testKey = blindKey.Serialize()
	return testAddress, testKey, nil
}

func TestGetUnspents(t *testing.T) {
	address, blindKey, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	explorerSvc := explorer.NewService("http://localhost:3001")
	_, err = explorerSvc.Faucet(address)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	elementsSvc, err := NewService(
		"localhost",
		7041,
		"admin1",
		"123",
	)
	if err != nil {
		t.Fatal(err)
	}

	utxos, err := elementsSvc.GetUnspents(address, [][]byte{blindKey})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(utxos))
	assert.Equal(t, true, len(utxos[0].Nonce()) > 1)
	assert.Equal(t, true, len(utxos[0].RangeProof()) > 0)
	assert.Equal(t, true, len(utxos[0].SurjectionProof()) > 0)

}
