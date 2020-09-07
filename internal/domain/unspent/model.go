package unspent

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
}

func NewUnspent(
	txID, assetHash, address string,
	vOut uint32,
	value uint64,
	spent, locked bool,
	scriptPubKey []byte,
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

func (u *Unspent) GetAddress() string {
	return u.address
}

func (u *Unspent) GetAssetHash() string {
	return u.assetHash
}

func (u *Unspent) GetValue() uint64 {
	return u.value
}

func (u *Unspent) GetTxID() string {
	return u.txID
}

func (u *Unspent) GetVOut() uint32 {
	return u.vOut
}

func (u *Unspent) Lock() {
	u.locked = true
}

func (u *Unspent) UnLock() {
	u.locked = false
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
