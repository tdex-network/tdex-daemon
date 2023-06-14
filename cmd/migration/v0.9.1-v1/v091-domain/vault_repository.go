package v091domain

import (
	"context"

	"github.com/sekulicd/badgerhold/v2"
)

type VaultRepository interface {
	GetVault(ctx context.Context) (*Vault, error)
}

type vaultRepositoryImpl struct {
	store *badgerhold.Store
}

func NewVaultRepositoryImpl(store *badgerhold.Store) VaultRepository {
	return &vaultRepositoryImpl{store}
}

func (v *vaultRepositoryImpl) GetVault(
	ctx context.Context,
) (*Vault, error) {
	var vault *Vault
	if err := v.store.Get(vaultKey, vault); err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return vault, nil
}
