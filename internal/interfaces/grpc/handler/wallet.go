package grpchandler

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type walletHandler struct {
	pb.UnimplementedWalletServer
	walletSvc application.WalletService
}

func NewWalletHandler(walletSvc application.WalletService) pb.WalletServer {
	return &walletHandler{
		walletSvc: walletSvc,
	}
}

func (w walletHandler) GenSeed(
	ctx context.Context,
	req *pb.GenSeedRequest,
) (*pb.GenSeedReply, error) {
	mnemonic, err := w.walletSvc.GenSeed(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GenSeedReply{SeedMnemonic: mnemonic}, nil
}

func (w walletHandler) InitWallet(
	req *pb.InitWalletRequest,
	stream pb.Wallet_InitWalletServer,
) error {
	if req.SeedMnemonic == nil || req.WalletPassword == nil {
		return status.Error(
			codes.InvalidArgument,
			"seed and password must be populated",
		)
	}

	if err := w.walletSvc.InitWallet(
		context.Background(),
		req.SeedMnemonic,
		string(req.WalletPassword),
	); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if err := stream.Send(&pb.InitWalletReply{}); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (w walletHandler) UnlockWallet(
	ctx context.Context,
	req *pb.UnlockWalletRequest,
) (*pb.UnlockWalletReply, error) {
	err := w.walletSvc.UnlockWallet(
		ctx,
		string(req.WalletPassword),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UnlockWalletReply{}, nil
}

func (w walletHandler) ChangePassword(
	ctx context.Context,
	req *pb.ChangePasswordRequest,
) (*pb.ChangePasswordReply, error) {
	err := w.walletSvc.ChangePassword(
		ctx,
		string(req.CurrentPassword),
		string(req.NewPassword),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ChangePasswordReply{}, nil
}

func (w walletHandler) WalletAddress(
	ctx context.Context,
	req *pb.WalletAddressRequest,
) (*pb.WalletAddressReply, error) {
	address, blindingKey, err := w.walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.WalletAddressReply{
		Address:  address,
		Blinding: blindingKey,
	}, nil
}

func (w walletHandler) WalletBalance(
	ctx context.Context,
	req *pb.WalletBalanceRequest,
) (*pb.WalletBalanceReply, error) {
	panic("implement me")
}

func (w walletHandler) SendToMany(
	ctx context.Context,
	req *pb.SendToManyRequest,
) (*pb.SendToManyReply, error) {

	outputs := make([]application.TxOut, 0)
	for _, v := range req.Outputs {
		output := application.TxOut{
			Asset:   v.Asset,
			Value:   v.Value,
			Address: v.Address,
		}
		outputs = append(outputs, output)
	}

	walletReq := application.SendToManyRequest{
		Outputs:         outputs,
		MillisatPerByte: req.MillisatPerByte,
		Push:            req.Push,
	}
	rawTx, err := w.walletSvc.SendToMany(ctx, walletReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SendToManyReply{
		RawTx: rawTx,
	}, nil
}
