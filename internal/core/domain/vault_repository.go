package domain

import (
	"context"

	"github.com/vulpemventures/go-elements/network"
)

// VaultRepository is the abstraction for any kind of database intended to
// persist a Vault.
type VaultRepository interface {
	// GetOrCreateVault returns the Vault stored in the repo. If not yet created,
	// a new one is created the provided mnemonic, passphrase, and network.
	GetOrCreateVault(
		ctx context.Context,
		mnemonic []string,
		passphrase string,
		network *network.Network,
	) (*Vault, error)
	// GetAccountByIndex returns the account with the given index, if it
	// exists.
	GetAccountByIndex(ctx context.Context, accountIndex int) (*Account, error)
	// GetAccountByAddress returns the account with the given index, if it
	// exists.
	GetAccountByAddress(ctx context.Context, addr string) (*Account, int, error)
	// GetAllDerivedAddressesAndBlindingKeysForAccount returns the list of all
	// external and internal (change) addresses derived for the provided account
	// along with the respective private blinding keys.
	GetAllDerivedAddressesAndBlindingKeysForAccount(
		ctx context.Context,
		accountIndex int,
	) ([]string, [][]byte, error)
	// GetDerivationPathByScript returns the derivation paths for the given account
	// index and the given list of scripts.
	GetDerivationPathByScript(
		ctx context.Context,
		accountIndex int,
		scripts []string,
	) (map[string]string, error)
	// GetAllDerivedExternalAddressesForAccount returns all receiving addresses
	// derived for the provided account so far.
	GetAllDerivedExternalAddressesForAccount(
		ctx context.Context,
		accountIndex int,
	) ([]string, error)
	// UpdateVault is the method allowing to make multiple changes to a vault in
	// a transactional way.
	UpdateVault(
		ctx context.Context,
		updateFn func(v *Vault) (*Vault, error),
	) error
}
