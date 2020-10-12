package domain

import (
	"github.com/google/uuid"
)

type UnspentKey struct {
	TxID string
	VOut uint32
}

type Unspent struct {
	TxID         string
	VOut         uint32
	Value        uint64
	AssetHash    string
	Address      string `badgerhold:"Address"`
	Spent        bool
	Locked       bool
	ScriptPubKey []byte
	LockedBy     *uuid.UUID
	Confirmed    bool
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
