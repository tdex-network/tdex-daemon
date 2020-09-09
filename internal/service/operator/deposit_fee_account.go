package operatorservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/constant"
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DepositFeeAccount returns a new address for the fee account
func (s *Service) DepositFeeAccount(ctx context.Context, req *pb.DepositFeeAccountRequest) (*pb.DepositFeeAccountReply, error) {
	var ctAddress string
	err := s.vaultRepository.UpdateVault(ctx, func(v *vault.Vault) (*vault.Vault, error) {
		addr, _, err := v.DeriveNextExternalAddressForAccount(constant.FeeAccount)
		if err != nil {
			return nil, err
		}
		ctAddress = addr
		return v, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DepositFeeAccountReply{
		Address: ctAddress,
	}, nil
}