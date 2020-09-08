package walletservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GenSeed creates and returns a new mnemonic for the wallet
func (s *Service) GenSeed(ctx context.Context, req *pb.GenSeedRequest) (*pb.GenSeedResponse, error) {
	_, mnemonic, err := s.vaultRepository.CreateOrRestoreVault(ctx, "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GenSeedResponse{SeedMnemonic: []string{mnemonic}}, nil
}
