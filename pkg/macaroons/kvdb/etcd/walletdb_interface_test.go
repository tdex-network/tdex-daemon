// +build kvdb_etcd

package etcd

/*
Modified from https://github.com/lightningnetwork/lnd/blob/master/macaroons/auth.go
Original Copyright 2017 Olaoluwa Osuntokun. All Rights Reserved. See LICENSE-MACAROON-LND for licensing terms.
*/

import (
	"testing"

	"github.com/btcsuite/btcwallet/walletdb/walletdbtest"
)

// TestWalletDBInterface performs the WalletDB interface test suite for the
// etcd database driver.
func TestWalletDBInterface(t *testing.T) {
	f := NewEtcdTestFixture(t)
	defer f.Cleanup()
	walletdbtest.TestInterface(t, dbType, f.BackendConfig())
}
