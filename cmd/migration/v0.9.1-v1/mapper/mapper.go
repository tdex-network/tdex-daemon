package mapper

import (
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"
	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-domain"
)

type Service interface {
	FromV091VaultToV1Wallet(
		vault v091domain.Vault, walletPass string,
	) (*v1domain.Wallet, error)
}

type mapperService struct {
	v091RepoManager v091domain.Repository
}

func NewService(v091RepoManager v091domain.Repository) Service {
	return &mapperService{
		v091RepoManager: v091RepoManager,
	}
}
