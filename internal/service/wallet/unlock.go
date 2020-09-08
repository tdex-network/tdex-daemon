package walletservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnlockWallet attempts to unlock the wallet database with the given password
func (s *Service) UnlockWallet(ctx context.Context, req *pb.UnlockWalletRequest) (*pb.UnlockWalletResponse, error) {
	passphrase := string(req.GetWalletPassword())
	if err := s.vaultRepository.Unlock(ctx, passphrase); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.UnlockWalletResponse{}, nil
}
