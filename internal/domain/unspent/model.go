package unspent

type Unspent struct {
	Txid      string
	Vout      uint32
	Value     uint64
	AssetHash string
	Address   string
	Spent     bool
	Locked    bool
}
