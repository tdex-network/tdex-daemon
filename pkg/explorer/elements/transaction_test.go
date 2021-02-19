package elements

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTransactionHex(t *testing.T) {
	elementsSvc, err := NewService("localhost", 7041, "admin1", "123")
	if err != nil {
		t.Fatal(err)
	}

	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	txid, err := elementsSvc.Faucet(addr)
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
	elementsSvc, err := NewService("localhost", 7041, "admin1", "123")
	if err != nil {
		t.Fatal(err)
	}

	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	txid, err := elementsSvc.Faucet(addr)
	if err != nil {
		t.Fatal(err)
	}

	isConfirmed, err := elementsSvc.IsTransactionConfirmed(txid)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, isConfirmed)
}
func TestGetTransactionStatus(t *testing.T) {
	elementsSvc, err := NewService("localhost", 7041, "admin1", "123")
	if err != nil {
		t.Fatal(err)
	}

	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	txid, err := elementsSvc.Faucet(addr)
	if err != nil {
		t.Fatal(err)
	}

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
	elementsSvc, err := NewService("localhost", 7041, "admin1", "123")
	if err != nil {
		t.Fatal(err)
	}

	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := elementsSvc.Faucet(addr); err != nil {
		t.Fatal(err)
	}
	if _, _, err := elementsSvc.Mint(addr, 10); err != nil {
		t.Fatal(err)
	}

	txs, err := elementsSvc.GetTransactionsForAddress(addr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(txs))
}
