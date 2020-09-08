package walletservice

import (
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
)

// Service is used to implement Wallet service.
type Service struct {
	vaultRepository vault.Repository
	pb.UnimplementedWalletServer
}

// NewService returns a Wallet Service
func NewService(vaultRepo vault.Repository, wallet *wallet.Wallet) *Service {
	return &Service{
		vaultRepository: vaultRepo,
	}
}
