package macaroons

/*
Modified from https://github.com/lightningnetwork/lnd/blob/master/macaroons/auth.go
Original Copyright 2017 Olaoluwa Osuntokun. All Rights Reserved. See LICENSE-MACAROON-LND for licensing terms.
*/

import "github.com/btcsuite/btcwallet/snacl"

var (
	// Below are the default scrypt parameters that are used when creating
	// the encryption key for the macaroon database with snacl.NewSecretKey.
	scryptN = snacl.DefaultN
	scryptR = snacl.DefaultR
	scryptP = snacl.DefaultP
)
