package esplora

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
)

var oneLbtc = 100000000

func TestGetUnspents(t *testing.T) {
	address, blindKey, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}
	explorerSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	_, err = explorerSvc.Faucet(address, oneLbtc)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	utxos, err := explorerSvc.GetUnspents(address, [][]byte{blindKey})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(utxos))

	if len(utxos) > 0 {
		assert.Equal(t, true, len(utxos[0].Nonce()) > 1)
		assert.Equal(t, true, len(utxos[0].RangeProof()) > 0)
		assert.Equal(t, true, len(utxos[0].SurjectionProof()) > 0)
	}
}

func TestSelectUnspents(t *testing.T) {
	addr, blindKey, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}
	explorerSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := explorerSvc.Faucet(addr, oneLbtc); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	utxos, err := explorerSvc.GetUnspents(addr, [][]byte{blindKey})
	if err != nil {
		t.Fatal(err)
	}

	targetAmount := uint64(0.7 * math.Pow10(8))

	selectedUtxos, change, err := explorer.SelectUnspents(
		utxos,
		targetAmount,
		network.Regtest.AssetID,
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedChange := uint64(0.3 * math.Pow10(8))
	assert.Equal(t, 1, len(selectedUtxos))
	assert.Equal(t, expectedChange, change)
}

func newService() (explorer.Service, error) {
	endpoint := os.Getenv("TEST_EXPLORER_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:3001"
	}
	requestTimeout := 5000
	return NewService(endpoint, requestTimeout)
}

func newTestData() (string, []byte, error) {
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

	return addr, blindKey.Serialize(), nil
}
