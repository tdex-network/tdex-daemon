package operatorservice

import (
	"context"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

func (s *Service) DepositFeeAccount(
	ctx context.Context,
	depositFeeAccountRequest *pb.DepositFeeAccountRequest,
) (*pb.DepositFeeAccountReply, error) {
	//just create fee account address
	//return address
	return nil, nil
}
