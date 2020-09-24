package walletservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnlockWallet attempts to unlock the wallet database with the given password
func (s *Service) UnlockWallet(ctx context.Context, req *pb.UnlockWalletRequest) (*pb.UnlockWalletReply, error) {
	passphrase := string(req.GetWalletPassword())
	if err := s.vaultRepository.UpdateVault(ctx, nil, "", func(v *vault.Vault) (*vault.Vault, error) {
		if err := v.Unlock(passphrase); err != nil {
			return nil, err
		}
		return v, nil
	}); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.UnlockWalletReply{}, nil
}
