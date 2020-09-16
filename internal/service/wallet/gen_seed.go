package walletservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GenSeed creates and returns a new mnemonic for the wallet
func (s *Service) GenSeed(ctx context.Context, req *pb.GenSeedRequest) (*pb.GenSeedReply, error) {
	mnemonic, err := wallet.NewMnemonic(wallet.NewMnemonicOpts{EntropySize: 256})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GenSeedReply{SeedMnemonic: mnemonic}, nil
}
