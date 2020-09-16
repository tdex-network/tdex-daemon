package walletservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InitWallet creates or restores a wallet for the daemon
func (s *Service) InitWallet(ctx context.Context, req *pb.InitWalletRequest) (*pb.InitWalletReply, error) {
	mnemonic := req.GetSeedMnemonic()
	passphrase := string(req.GetWalletPassword())
	if err := s.vaultRepository.UpdateVault(ctx, mnemonic, passphrase, func(v *vault.Vault) (*vault.Vault, error) {
		return v, nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.InitWalletReply{}, nil
}
