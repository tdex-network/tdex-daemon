package vault

import "context"

type Repository interface {
	GetOrCreateMnemonic(ctx context.Context) (string, error)
	RestoreFromMnemonic(ctx context.Context, mnemonic string) error
	Lock(ctx context.Context, passphrase string) error
	Unlock(ctx context.Context, passphrase string) error
	ChangePassphrase(ctx context.Context, oldPassphrase, newPassphrase string) error
	GetOrCreateAccount(ctx context.Context, accountIndex uint32) (*Account, error)
	UpdateAccount(
		ctx context.Context,
		accountIndex uint32,
		updateFn func(a *Account) (*Account, error),
	) error
	GetAccountByAddress(ctx context.Context, addr string) (*Account, int, error)
	AddAccountByAddress(ctx context.Context, addr string, accountIndex uint32) error
}
