package unlocker

import (
	"context"
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/core/application/pubsub"
	"github.com/tdex-network/tdex-daemon/internal/core/application/wallet"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type service struct {
	wallet *wallet.Service
	pubsub *pubsub.Service
}

func NewService(
	walletSvc *wallet.Service, pubsubSvc *pubsub.Service,
) (*service, error) {
	if walletSvc == nil {
		return nil, fmt.Errorf("missing wallet service")
	}
	if pubsubSvc == nil {
		return nil, fmt.Errorf("missing pubsub service")
	}

	return &service{walletSvc, pubsubSvc}, nil
}

func (s *service) GenSeed(ctx context.Context) ([]string, error) {
	return s.wallet.Wallet().GenSeed(ctx)
}

func (s *service) InitWallet(
	ctx context.Context, mnemonic []string, password string,
) error {
	if err := s.wallet.Wallet().InitWallet(
		ctx, mnemonic, password,
	); err != nil {
		return err
	}
	// nolint
	s.pubsub.SecurePubSub().Store().Init(password)
	return nil
}

func (s *service) RestoreWallet(
	ctx context.Context, mnemonic []string, password string,
) error {
	if err := s.wallet.Wallet().RestoreWallet(
		ctx, mnemonic, password,
	); err != nil {
		return err
	}
	// nolint
	s.pubsub.SecurePubSub().Store().Init(password)
	return nil
}

func (s *service) UnlockWallet(ctx context.Context, password string) error {
	if err := s.wallet.Wallet().Unlock(ctx, password); err != nil {
		return err
	}
	// nolint
	s.pubsub.SecurePubSub().Store().Unlock(password)
	return nil
}

func (s *service) LockWallet(ctx context.Context, password string) error {
	if err := s.wallet.Wallet().Lock(ctx, password); err != nil {
		return err
	}
	s.pubsub.SecurePubSub().Store().Lock()
	return nil
}

func (s *service) ChangePassword(
	ctx context.Context, oldPwd, newPwd string,
) error {
	if err := s.wallet.Wallet().ChangePassword(ctx, oldPwd, newPwd); err != nil {
		return err
	}
	// nolint
	s.pubsub.SecurePubSub().Store().ChangePassword(oldPwd, newPwd)
	return nil
}

func (s *service) Status(ctx context.Context) (ports.WalletStatus, error) {
	return s.wallet.Wallet().Status(ctx)
}

func (s *service) Info(ctx context.Context) (ports.WalletInfo, error) {
	return s.wallet.Wallet().Info(ctx)
}
