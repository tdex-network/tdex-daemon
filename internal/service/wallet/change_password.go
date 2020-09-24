package walletservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ChangePassword changes the wallet passphrase
func (s *Service) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordReply, error) {
	currentPassphrase := string(req.GetCurrentPassword())
	newPassphrase := string(req.GetNewPassword())

	if err := s.vaultRepository.UpdateVault(ctx, nil, "", func(v *vault.Vault) (*vault.Vault, error) {
		if err := v.ChangePassphrase(currentPassphrase, newPassphrase); err != nil {
			return nil, err
		}
		return v, nil
	}); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.ChangePasswordReply{}, nil
}
