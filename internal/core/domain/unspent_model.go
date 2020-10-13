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
	txID            string
	vOut            uint32
	value           uint64
	assetHash       string
	valueCommitment string
	assetCommitment string
	scriptPubKey    []byte
	nonce           []byte
	rangeProof      []byte
	surjectionProof []byte
	address         string
	spent           bool
	locked          bool
	lockedBy        *uuid.UUID
	confirmed       bool
}

func NewUnspent(
	txID string, vOut uint32,
	value uint64, assetHash string, scriptPubKey []byte,
	valueCommitment, assetCommitment string,
	nonce, rangeProof, surjectionProof []byte,
	address string, confirmed bool,
) Unspent {
	return Unspent{
		txID:            txID,
		vOut:            vOut,
		value:           value,
		assetHash:       assetHash,
		valueCommitment: valueCommitment,
		assetCommitment: assetCommitment,
		nonce:           nonce,
		rangeProof:      rangeProof,
		surjectionProof: surjectionProof,
		address:         address,
		scriptPubKey:    scriptPubKey,
		confirmed:       confirmed,
	}
}

type BalanceInfo struct {
	TotalBalance       uint64
	ConfirmedBalance   uint64
	UnconfirmedBalance uint64
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

func (u *Unspent) ValueCommitment() string {
	return u.valueCommitment
}

func (u *Unspent) AssetCommitment() string {
	return u.assetCommitment
}

func (u *Unspent) Script() []byte {
	return u.scriptPubKey
}

func (u *Unspent) Nonce() []byte {
	return u.nonce
}

func (u *Unspent) RangeProof() []byte {
	return u.rangeProof
}

func (u *Unspent) SurjectionProof() []byte {
	return u.surjectionProof
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

func (u *Unspent) IsConfirmed() bool {
	return u.confirmed
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

func (u *Unspent) ToUtxo() explorer.Utxo {
	return explorer.NewWitnessUtxo(
		u.txID,
		u.vOut,
		u.value,
		u.assetHash,
		u.valueCommitment,
		u.assetCommitment,
		u.scriptPubKey,
		u.nonce,
		u.rangeProof,
		u.surjectionProof,
		u.confirmed,
	)
}
