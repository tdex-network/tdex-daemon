package walletservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ChangePassword changes the wallet passphrase
func (s *Service) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	currentPassphrase := string(req.GetCurrentPassword())
	newPassphrase := string(req.GetNewPassword())

	err := s.vaultRepository.ChangePassphrase(ctx, currentPassphrase, newPassphrase)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.ChangePasswordResponse{}, nil
}
