package v091domain

import (
	"github.com/sekulicd/badgerhold/v2"
)

type VaultRepository interface {
	GetVault() (*Vault, error)
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
