package walletservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InitWallet creates or restores a wallet for the daemon
func (s *Service) InitWallet(ctx context.Context, req *pb.InitWalletRequest) (*pb.InitWalletResponse, error) {
	mnemonics := req.GetSeedMnemonic()
	if len(mnemonics) > 0 {
		err := s.vaultRepository.RestoreFromMnemonic(ctx, mnemonics[0])
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	passphrase := req.GetWalletPassword()
	if len(passphrase) > 0 {
		if err := s.vaultRepository.Lock(ctx, string(passphrase)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.InitWalletResponse{}, nil
}
