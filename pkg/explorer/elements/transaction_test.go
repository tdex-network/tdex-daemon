package elements

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTransaction(t *testing.T) {
	elementsSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	txid, err := elementsSvc.Faucet(addr, oneLbtc, "")
	if err != nil {
		t.Fatal(err)
	}

	tx, err := elementsSvc.GetTransaction(txid)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, tx != nil)
}

func TestGetTransactionHex(t *testing.T) {
	elementsSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	txid, err := elementsSvc.Faucet(addr, oneLbtc, "")
	if err != nil {
		t.Fatal(err)
	}

	txhex, err := elementsSvc.GetTransactionHex(txid)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, len(txhex) > 0)
}

func TestIsTransactionConfirmed(t *testing.T) {
	elementsSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	txid, err := elementsSvc.Faucet(addr, oneLbtc, "")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	isConfirmed, err := elementsSvc.IsTransactionConfirmed(txid)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, isConfirmed)
}
func TestGetTransactionStatus(t *testing.T) {
	elementsSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	txid, err := elementsSvc.Faucet(addr, oneLbtc, "")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	status, err := elementsSvc.GetTransactionStatus(txid)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, status["confirmed"].(bool))
	assert.Equal(t, true, len(status["block_hash"].(string)) > 0)
	assert.Equal(t, true, status["block_height"].(float64) > 0)
	assert.Equal(t, true, status["block_time"].(float64) > 0)
}

func TestGetTransactionsForAddress(t *testing.T) {
	elementsSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	addr, blindKey, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := elementsSvc.Faucet(addr, oneLbtc, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := elementsSvc.Mint(addr, 10); err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	txs, err := elementsSvc.GetTransactionsForAddress(addr, blindKey)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(txs))
}
