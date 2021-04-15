package domain

import (
	"github.com/google/uuid"
)

// UnspentKey represent the ID of an Unspent, composed by its txid and vout.
type UnspentKey struct {
	TxID string
	VOut uint32
}

// Unspent is the data structure representing an Elements based UTXO with some
// other information like whether it is spent/unspent, confirmed/unconfirmed or
// locked/unlocked.
type Unspent struct {
	TxID            string
	VOut            uint32
	Value           uint64
	AssetHash       string
	ValueCommitment string
	AssetCommitment string
	ValueBlinder    []byte
	AssetBlinder    []byte
	ScriptPubKey    []byte
	Nonce           []byte
	Address         string
	Spent           bool
	Locked          bool
	LockedBy        *uuid.UUID
	Confirmed       bool
}
