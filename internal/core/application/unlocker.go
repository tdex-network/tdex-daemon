package application

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application/pubsub"
	"github.com/tdex-network/tdex-daemon/internal/core/application/unlocker"
	"github.com/tdex-network/tdex-daemon/internal/core/application/wallet"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type UnlockerService interface {
	GenSeed(ctx context.Context) ([]string, error)
	InitWallet(ctx context.Context, mnemonic []string, password string) error
	RestoreWallet(ctx context.Context, mnemonic []string, password string) error
	UnlockWallet(ctx context.Context, password string) error
	LockWallet(ctx context.Context, password string) error
	ChangePassword(ctx context.Context, oldPwd, newPwd string) error
	Status(ctx context.Context) (ports.WalletStatus, error)
	Info(ctx context.Context) (ports.WalletInfo, error)
}

func NewUnlockerService(
	walletSvc WalletService, pubsubSvc PubSubService,
) (UnlockerService, error) {
	w := walletSvc.(*wallet.Service)
	p := pubsubSvc.(*pubsub.Service)
	return unlocker.NewService(w, p)
}
