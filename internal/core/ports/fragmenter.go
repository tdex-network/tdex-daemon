package ports

import (
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type AssetValuePair interface {
	BaseValue() uint64
	BaseAsset() string
	QuoteValue() uint64
	QuoteAsset() string
}

type TxOut interface {
	Asset() string
	Value() uint64
	Address() string
}

type Fragmenter interface {
	Address() string
	Keys() ([]byte, []byte)
	FragmentAmount(args ...interface{}) ([]uint64, []uint64, error)
	CraftTransaction(
		ins []explorer.Utxo, outs []TxOut, feeAmount uint64, lbtc string,
	) (string, error)
}
