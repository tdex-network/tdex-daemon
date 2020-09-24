package unspent

import "github.com/google/uuid"

type UnspentKey struct {
	TxID string
	VOut uint32
}

type Unspent struct {
	txID         string
	vOut         uint32
	value        uint64
	assetHash    string
	address      string
	spent        bool
	locked       bool
	scriptPubKey []byte
	lockedBy     *uuid.UUID
}

func NewUnspent(
	txID, assetHash, address string,
	vOut uint32,
	value uint64,
	spent, locked bool,
	scriptPubKey []byte,
	lockedBy *uuid.UUID,
) Unspent {
	return Unspent{
		txID:         txID,
		vOut:         vOut,
		value:        value,
		assetHash:    assetHash,
		address:      address,
		spent:        spent,
		locked:       locked,
		scriptPubKey: scriptPubKey,
	}
}

func (u *Unspent) Address() string {
	return u.address
}

func (u *Unspent) AssetHash() string {
	return u.assetHash
}

func (u *Unspent) Value() uint64 {
	return u.value
}

func (u *Unspent) TxID() string {
	return u.txID
}

func (u *Unspent) VOut() uint32 {
	return u.vOut
}

func (u *Unspent) Lock(tradeID *uuid.UUID) {
	u.locked = true
	u.lockedBy = tradeID
}

func (u *Unspent) UnLock() {
	u.locked = false
	u.lockedBy = nil
}

func (u *Unspent) IsLocked() bool {
	return u.locked
}

func (u *Unspent) Spend() {
	u.spent = true
}

func (u *Unspent) IsSpent() bool {
	return u.spent
}

func (u *Unspent) GetKey() UnspentKey {
	return UnspentKey{
		TxID: u.txID,
		VOut: u.vOut,
	}
}

func (u *Unspent) IsKeyEqual(key UnspentKey) bool {
	return u.txID == key.TxID && u.vOut == key.VOut
}

func (u *Unspent) LockedBy() *uuid.UUID {
	return u.lockedBy
}
