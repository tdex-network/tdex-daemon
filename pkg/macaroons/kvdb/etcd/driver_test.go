// +build kvdb_etcd

package etcd

/*
Modified from https://github.com/lightningnetwork/lnd/blob/master/macaroons/auth.go
Original Copyright 2017 Olaoluwa Osuntokun. All Rights Reserved. See LICENSE-MACAROON-LND for licensing terms.
*/

import (
	"testing"

	"github.com/btcsuite/btcwallet/walletdb"
	"github.com/stretchr/testify/require"
)

func TestOpenCreateFailure(t *testing.T) {
	t.Parallel()

	db, err := walletdb.Open(dbType)
	require.Error(t, err)
	require.Nil(t, db)

	db, err = walletdb.Open(dbType, "wrong")
	require.Error(t, err)
	require.Nil(t, db)

	db, err = walletdb.Create(dbType)
	require.Error(t, err)
	require.Nil(t, db)

	db, err = walletdb.Create(dbType, "wrong")
	require.Error(t, err)
	require.Nil(t, db)
}
