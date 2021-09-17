package domain

import (
	"context"

	"github.com/google/uuid"
)

// UnspentRepository is the abstraction for any kind of database intended to
// persist Unspents.
type UnspentRepository interface {
	// AddUnspents adds the provided unspents to the repository. Those already
	// existing won't be re-added
	AddUnspents(ctx context.Context, unspents []Unspent) (int, error)
	// GetAllUnspents returns the entire UTXO set, included those locked or
	// already spent.
	GetAllUnspents(ctx context.Context) []Unspent
	// GetAvailableUnspents returns all unlocked unspent UTXOs.
	GetAvailableUnspents(ctx context.Context) ([]Unspent, error)
	// GetAllUnspentsForAddresses returns the entire UTXO set (locked and spent
	// included) for the provided list of addresses.
	GetAllUnspentsForAddresses(ctx context.Context, addresses []string) ([]Unspent, error)
	// GetAllUnspentsForAddressesAndPage returns a subset of the entire UTXO set
	// (locked and spent included) for the provided list of addresses.
	GetAllUnspentsForAddressesAndPage(ctx context.Context, addresses []string, page Page) ([]Unspent, error)
	// GetUnspentsForAddresses returns the list of all unspent UTXOs for the
	// provided list of address (locked unspents included).
	GetUnspentsForAddresses(ctx context.Context, addresses []string) ([]Unspent, error)
	// GetAvailableUnspentsForAddresses returns the list of spendable UTXOs for the
	// provided list of addresses (locked unspents excluded).
	GetAvailableUnspentsForAddresses(ctx context.Context, addresses []string) ([]Unspent, error)
	// GetUnspentWithKey returns all the info about an UTXO, if existing in the
	// repository.
	GetUnspentWithKey(ctx context.Context, unspentKey UnspentKey) (*Unspent, error)
	// GetBalance returns the current balance of a certain asset for the provided
	// list of addresses (locked unspents included)
	GetBalance(ctx context.Context, addresses []string, assetHash string) (uint64, error)
	// GetUnlockedBalance returns the current balance of a certain asset for the
	// provided list of addresses (locked unspents exlcuded).
	GetUnlockedBalance(ctx context.Context, addresses []string, assetHash string) (uint64, error)
	// SpendUnspents let mark the provided list of unspent UTXOs (identified by their
	// keys) as spent.
	SpendUnspents(ctx context.Context, unspentKeys []UnspentKey) (int, error)
	// ConfirmUnspents let mark the provided list of unconfirmed unspent UTXOs as
	// confirmed.
	ConfirmUnspents(ctx context.Context, unspentKeys []UnspentKey) (int, error)
	// LockUnspents let lock the provided list of unlocked, unspent UTXOs,
	// referring to a certain trade by its UUID.
	LockUnspents(ctx context.Context, unspentKeys []UnspentKey, tradeID uuid.UUID) (int, error)
	// UnlockUnspents let unlock the provided list of locked, unspent UTXOs.
	UnlockUnspents(ctx context.Context, unspentKeys []UnspentKey) (int, error)
}
