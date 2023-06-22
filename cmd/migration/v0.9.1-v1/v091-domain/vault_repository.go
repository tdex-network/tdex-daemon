package v091domain

import (
	"github.com/sekulicd/badgerhold/v2"
)

type VaultRepository interface {
	GetVault() (*Vault, error)
	GetAccountByAddress(addr string) (*Account, int, error)
}

type vaultRepositoryImpl struct {
	store *badgerhold.Store
}

func NewVaultRepositoryImpl(store *badgerhold.Store) VaultRepository {
	return &vaultRepositoryImpl{store}
}

func (v *vaultRepositoryImpl) GetVault() (*Vault, error) {
	var vault Vault
	if err := v.store.Get(vaultKey, &vault); err != nil {
		return nil, err
	}

	return &vault, nil
}

func (v *vaultRepositoryImpl) GetAccountByAddress(
	addr string,
) (*Account, int, error) {
	vault, err := v.GetVault()
	if err != nil {
		return nil, -1, err
	}

	account, accountIndex, err := vault.AccountByAddress(addr)
	if err != nil {
		return nil, -1, err
	}

	return account, accountIndex, nil
}
