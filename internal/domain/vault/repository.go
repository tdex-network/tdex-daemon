package vault

import "context"

type Repository interface {
	CreateOrRestoreVault(ctx context.Context, mnemonic string) (*Vault, string, error)
	UpdateVault(
		ctx context.Context,
		updateFn func(v *Vault) (*Vault, error),
	) error
	GetAccountByIndex(ctx context.Context, accountIndex int) (*Account, error)
	GetAccountByAddress(ctx context.Context, addr string) (*Account, int, error)
	GetAllDerivedAddressesForAccount(ctx context.Context, accountIndex int) ([]string, error)
}
