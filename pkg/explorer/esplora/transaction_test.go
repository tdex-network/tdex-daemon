package esplora

import (
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
)

func TestGetTransaction(t *testing.T) {
	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	explorerSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	// Fund sender address.
	// Empty asset hash defaults to LBTC.
	txID, err := explorerSvc.Faucet(addr, oneLbtc, "")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	tx, err := explorerSvc.GetTransaction(txID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, tx != nil)
}

func TestGetTransactionStatus(t *testing.T) {
	addr, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	explorerSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	// Fund sender address.
	txID, err := explorerSvc.Faucet(addr, oneLbtc, "")
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
	addr, blindKey, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	explorerSvc, err := newService()
	if err != nil {
		t.Fatal(err)
	}

	// Fund sender address.
	if _, err := explorerSvc.Faucet(addr, oneLbtc, ""); err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	txs, err := explorerSvc.GetTransactionsForAddress(addr, blindKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(txs))
}
