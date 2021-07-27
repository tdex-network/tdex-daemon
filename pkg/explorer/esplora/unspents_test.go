package esplora

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
)

var oneLbtc = float64(1)

func TestGetUnspents(t *testing.T) {
	address, blindKey, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}
	explorerSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	_, err = explorerSvc.Faucet(address, oneLbtc, "")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	utxos, err := explorerSvc.GetUnspents(address, [][]byte{blindKey})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 1, len(utxos))

	utxo := utxos[0]
	require.Equal(t, true, len(utxo.Nonce()) > 1)
	require.Equal(t, true, len(utxo.RangeProof()) > 0)
	require.Equal(t, true, len(utxo.SurjectionProof()) > 0)

	status, err := explorerSvc.GetUnspentStatus(utxo.Hash(), utxo.Index())
	require.NoError(t, err)
	require.False(t, status.Spent())
	require.Empty(t, status.Hash())
	require.Equal(t, -1, status.Index())
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

	if _, err := explorerSvc.Faucet(addr, oneLbtc, ""); err != nil {
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

func TestGetUnspentStatus(t *testing.T) {
	addr, blindKey, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}
	explorerSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := explorerSvc.Faucet(addr, oneLbtc, ""); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	utxos, err := explorerSvc.GetUnspents(addr, [][]byte{blindKey})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(utxos))

	hash := utxos[0].Hash()
	index := utxos[0].Index()
	status, err := explorerSvc.GetUnspentStatus(hash, index)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, status.Spent())
	assert.Empty(t, status.Hash())
	assert.Equal(t, -1, status.Index())

	tx, err := explorerSvc.GetTransaction(hash)
	if err != nil {
		t.Fatal(err)
	}
	spentUtxo := tx.Inputs()[0]
	spentUtxoHash := bufferutil.TxIDFromBytes(spentUtxo.Hash)
	spentUtxoIndex := spentUtxo.Index
	spentUtxoStatus, err := explorerSvc.GetUnspentStatus(spentUtxoHash, spentUtxoIndex)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, spentUtxoStatus.Spent())
	assert.Equal(t, hash, spentUtxoStatus.Hash())
	assert.Equal(t, 0, spentUtxoStatus.Index())
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
