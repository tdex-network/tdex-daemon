package vault

import "context"

type Repository interface {
	GetOrCreateVault(mnemonic []string, passphrase string) (*Vault, error)
	UpdateVault(
		ctx context.Context,
		mnemonic []string,
		passphrase string,
		updateFn func(v *Vault) (*Vault, error),
	) error
	GetAccountByIndex(ctx context.Context, accountIndex int) (*Account, error)
	GetAccountByAddress(ctx context.Context, addr string) (*Account, int, error)
	GetAllDerivedAddressesForAccount(ctx context.Context, accountIndex int) ([]string, error)
}
