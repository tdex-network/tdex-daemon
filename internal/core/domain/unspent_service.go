package domain

import (
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
)

// IsKeyEqual returns whether the provided UnspentKey matches that or the
// current unspent.
func (u *Unspent) IsKeyEqual(key UnspentKey) bool {
	return u.TxID == key.TxID && u.VOut == key.VOut
}

// IsSpent returns whether the unspent is already spent.
func (u *Unspent) IsSpent() bool {
	return u.Spent
}

// IsConfirmed returns whether the unspent is already confirmed.
func (u *Unspent) IsConfirmed() bool {
	return u.Confirmed
}

// IsLocked returns whether the unspent is already locked - used in some not yet
// broadcasted trade.
func (u *Unspent) IsLocked() bool {
	return u.Locked
}

// Key returns the UnspentKey of the current unspent.
func (u *Unspent) Key() UnspentKey {
	return UnspentKey{
		TxID: u.TxID,
		VOut: u.VOut,
	}
}

// Spend marks the unspents as spent.
func (u *Unspent) Spend() {
	u.Spent = true
}

// Confirm marks the unspents as confirmed.
func (u *Unspent) Confirm() {
	u.Confirmed = true
}

// Lock marks the current unspent as locked, referring to some trade by its
// UUID.
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

// Unlock marks the current locked unspent as unlocked.
func (u *Unspent) Unlock() {
	u.Locked = false
	u.LockedBy = nil
}

// ToUtxo returns the current unpsent as an explorer.Utxo interface
func (u *Unspent) ToUtxo() explorer.Utxo {
	return esplora.NewWitnessUtxo(
		u.TxID,
		u.VOut,
		u.Value,
		u.AssetHash,
		u.ValueCommitment,
		u.AssetCommitment,
		u.ValueBlinder,
		u.AssetBlinder,
		u.ScriptPubKey,
		u.Nonce,
		u.RangeProof,
		u.SurjectionProof,
		u.Confirmed,
	)
}
