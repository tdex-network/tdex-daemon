package v1domain

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
)

type Utxo struct {
	UtxoKey
	Value               uint64
	Asset               string
	ValueCommitment     []byte
	AssetCommitment     []byte
	ValueBlinder        []byte
	AssetBlinder        []byte
	Script              []byte
	Nonce               []byte
	RangeProof          []byte
	SurjectionProof     []byte
	AccountName         string
	LockTimestamp       int64
	LockExpiryTimestamp int64
	SpentStatus         UtxoStatus
	ConfirmedStatus     UtxoStatus
}

type UtxoKey struct {
	TxID string
	VOut uint32
}

type UtxoStatus struct {
	Txid        string
	BlockHeight uint64
	BlockTime   int64
	BlockHash   string
}

func (u *Utxo) Key() UtxoKey {
	return u.UtxoKey
}

func (k UtxoKey) Hash() string {
	buf, _ := hex.DecodeString(k.TxID)
	buf = append(buf, byte(k.VOut))
	return hex.EncodeToString(btcutil.Hash160(buf))
}
