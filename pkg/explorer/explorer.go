package explorer

import (
	"github.com/vulpemventures/go-elements/transaction"
)

// Utxo represents a transaction output in the elements chain.
type Utxo interface {
	Hash() string
	Index() uint32
	Value() uint64
	Asset() string
	ValueCommitment() string
	AssetCommitment() string
	ValueBlinder() []byte
	AssetBlinder() []byte
	Script() []byte
	Nonce() []byte
	RangeProof() []byte
	SurjectionProof() []byte
	IsConfidential() bool
	IsConfirmed() bool
	IsRevealed() bool
	Parse() (*transaction.TxInput, *transaction.TxOutput, error)
}

// UtxoStatus represents whether a UTXO is spent. If it's actually spent,
// it provides methods to know the hash and input index of the tx that spent it.
type UtxoStatus interface {
	Spent() bool
	Hash() string
	Index() int
}

// Transaction represents a transaction in the elements chain.
type Transaction interface {
	Hash() string
	Version() int
	Locktime() int
	Inputs() []*transaction.TxInput
	Outputs() []*transaction.TxOutput
	Size() int
	Weight() int
	Confirmed() bool
}

type TransactionStatus interface {
	Confirmed() bool
	BlockHash() string
	BlockHeight() int
	BlockTime() int
}

// Service is representation of an explorer that allows to fetch data from the
// blockchain, to broadcast transactions, and for regtest ONLY, to fund and
// and address with LBTC or some other asset.
type Service interface {
	// GetUnspents fetches and optionally unblinds utxos for the given address.
	GetUnspents(addr string, blindKeys [][]byte) (unspents []Utxo, err error)
	// GetUnspentsForAddresses fetches and optionally unblinds utxos of the given
	// list of addresses.
	GetUnspentsForAddresses(
		addresses []string,
		blindingKeys [][]byte,
	) (unspents []Utxo, err error)
	// GetUnspentStatus returns the status of the provided unspent txid:index.
	GetUnspentStatus(txid string, index uint32) (status UtxoStatus, err error)
	// GetTransaction fetches the transaction given its hash.
	GetTransaction(txid string) (tx Transaction, err error)
	// GetTransactionHex fetches the transaction in hex format given its hash.
	GetTransactionHex(txid string) (txhex string, err error)
	// IsTransactionConfirmed returns whether the tx identified by its hash has
	// been included in the blockchain.
	IsTransactionConfirmed(txid string) (confirmed bool, err error)
	// GetTransactionStatus returns the status of the tx identified by its hash.
	GetTransactionStatus(txid string) (status TransactionStatus, err error)
	// GetTransactionsForAddress returns the list of all txs relative to the
	// given address.
	GetTransactionsForAddress(address string, blindingKey []byte) (txs []Transaction, err error)
	// BroadcastTransaction attempts to add the given tx in hex format to the
	// mempool and returns its tx hash.
	BroadcastTransaction(txhex string) (txid string, err error)
	// GetBlockHeight returns the the number of block of the blockchain.
	GetBlockHeight() (int, error)
	/**** REGTEST ONLY ****/
	// Faucet funds the given address with the amount (in BTC) of provided asset
	Faucet(address string, amount float64, asset string) (txid string, err error)
	// Mint funds the given address with a certain amount (in BTC) of a new issued asset.
	Mint(address string, amount float64) (txid string, asset string, err error)
}
