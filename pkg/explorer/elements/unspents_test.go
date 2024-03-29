package elements

import (
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
)

var oneLbtc = float64(1)

func TestGetUnspents(t *testing.T) {
	elementsSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	address, blindKey, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := elementsSvc.Faucet(address, oneLbtc, ""); err != nil {
		t.Fatal(err)
	}

	utxos, err := elementsSvc.GetUnspents(address, [][]byte{blindKey})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(utxos))
	if len(utxos) > 0 {
		assert.Equal(t, true, utxos[0].IsConfidential())
		assert.Equal(t, true, utxos[0].IsRevealed())
		assert.Equal(t, true, len(utxos[0].Nonce()) > 0)
		assert.Equal(t, true, len(utxos[0].RangeProof()) > 0)
		assert.Equal(t, true, len(utxos[0].SurjectionProof()) > 0)
	}
}

func TestGetUnspentsForAddresses(t *testing.T) {
	elementsSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	addr1, blindKey1, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}
	addr2, blindKey2, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := elementsSvc.Faucet(addr1, oneLbtc, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := elementsSvc.Mint(addr2, 100); err != nil {
		t.Fatal(err)
	}

	utxos, err := elementsSvc.GetUnspentsForAddresses(
		[]string{addr1, addr2},
		[][]byte{blindKey1, blindKey2},
	)

	assert.Equal(t, 2, len(utxos))
	for _, utxo := range utxos {
		assert.Equal(t, true, utxo.IsRevealed())
		assert.Equal(t, true, utxo.IsConfidential())
	}
}

func newService() (explorer.Service, error) {
	rpcEndpoint := os.Getenv("TEST_ELEMENTS_ENDPOINT")
	if rpcEndpoint == "" {
		rpcEndpoint = "http://admin1:123@127.0.0.1:7041"
	}
	return NewService(rpcEndpoint, 0)
}

func newTestData() (string, []byte, error) {
	key, err := btcec.NewPrivateKey()
	if err != nil {
		return "", nil, err
	}
	blindKey, err := btcec.NewPrivateKey()
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
