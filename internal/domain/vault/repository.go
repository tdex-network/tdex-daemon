package vault

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/storageutil/uow"
)

type Repository interface {
	GetOrCreateVault(ctx context.Context, mnemonic []string, passphrase string) (*Vault, error)
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
	Begin() (uow.Tx, error)
	ContextKey() interface{}
}
