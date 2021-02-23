package esplora

import (
	"math"
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
	explorerSvc, err := NewService(explorerURL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = explorerSvc.Faucet(address)
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
	address, key1, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}
	explorerSvc, err := NewService(explorerURL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = explorerSvc.Faucet(address)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	utxos, err := explorerSvc.GetUnspents(address, [][]byte{key1})
	if err != nil {
		t.Fatal(err)
	}

	targetAmount := uint64(1.5 * math.Pow10(8))

	selectedUtxos, change, err := explorer.SelectUnspents(
		utxos,
		targetAmount,
		network.Regtest.AssetID,
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedChange := uint64(0.5 * math.Pow10(8))
	assert.Equal(t, 2, len(selectedUtxos))
	assert.Equal(t, expectedChange, change)
}
