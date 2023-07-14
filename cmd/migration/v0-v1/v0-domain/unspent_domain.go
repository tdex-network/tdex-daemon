package v0domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrUnspentAlreadyLocked = errors.New("cannot lock an already locked unspent")
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
	ValueBlinder    []byte
	AssetBlinder    []byte
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

func (u *Unspent) IsLocked() bool {
	return u.Locked
}

func (u *Unspent) Lock(tradeID *uuid.UUID) error {
	if u.IsLocked() {
		if tradeID.String() != u.LockedBy.String() {
			return ErrUnspentAlreadyLocked
		}
		return nil
	}

	u.Locked = true
	u.LockedBy = tradeID
	return nil
}

func (u *Unspent) Key() UnspentKey {
	return UnspentKey{
		TxID: u.TxID,
		VOut: u.VOut,
	}
}
