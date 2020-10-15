package domain

import (
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type UnspentKey struct {
	TxID string
	VOut uint32
}

type Unspent struct {
	TxID            string
	VOut            uint32
	Value           uint64
	AssetHash       string
	ValueCommitment string
	AssetCommitment string
	ScriptPubKey    []byte
	Nonce           []byte
	RangeProof      []byte
	SurjectionProof []byte
	Address         string
	Spent           bool
	Locked          bool
	LockedBy        *uuid.UUID
	Confirmed       bool
}

type BalanceInfo struct {
	TotalBalance       uint64
	ConfirmedBalance   uint64
	UnconfirmedBalance uint64
}

func (u *Unspent) Lock(tradeID *uuid.UUID) {
	u.Locked = true
	u.LockedBy = tradeID
}

func (u *Unspent) UnLock() {
	u.Locked = false
	u.LockedBy = nil
}

func (u *Unspent) IsLocked() bool {
	return u.Locked
}

func (u *Unspent) Spend() {
	u.Spent = true
}

func (u *Unspent) IsSpent() bool {
	return u.Spent
}

func (u *Unspent) IsConfirmed() bool {
	return u.Confirmed
}

func (u *Unspent) Key() UnspentKey {
	return UnspentKey{
		TxID: u.TxID,
		VOut: u.VOut,
	}
}

func (u *Unspent) IsKeyEqual(key UnspentKey) bool {
	return u.TxID == key.TxID && u.VOut == key.VOut
}

func (u *Unspent) ToUtxo() explorer.Utxo {
	return explorer.NewWitnessUtxo(
		u.TxID,
		u.VOut,
		u.Value,
		u.AssetHash,
		u.ValueCommitment,
		u.AssetCommitment,
		u.ScriptPubKey,
		u.Nonce,
		u.RangeProof,
		u.SurjectionProof,
		u.Confirmed,
	)
}
