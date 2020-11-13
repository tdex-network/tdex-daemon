package domain

import (
	"context"
)

type VaultRepository interface {
	GetOrCreateVault(
		ctx context.Context,
		mnemonic []string,
		passphrase string,
	) (*Vault, error)
	UpdateVault(
		ctx context.Context,
		mnemonic []string,
		passphrase string,
		updateFn func(v *Vault) (*Vault, error),
	) error
	GetAccountByIndex(ctx context.Context, accountIndex int) (*Account, error)
	GetAccountByAddress(ctx context.Context, addr string) (*Account, int, error)
	GetAllDerivedAddressesAndBlindingKeysForAccount(
		ctx context.Context,
		accountIndex int,
	) ([]string, [][]byte, error)
	GetDerivationPathByScript(
		ctx context.Context,
		accountIndex int,
		scripts []string,
	) (map[string]string, error)
	GetAllDerivedExternalAddressesForAccount(
		ctx context.Context,
		accountIndex int,
	) ([]string, error)
}
