package grpchandler

import (
	"context"
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/walletunlocker"
)

type walletUnlockerHandler struct {
	pb.UnimplementedWalletUnlockerServer
	walletSvc application.WalletService
}

func NewWalletUnlockerHandler(
	walletSvc application.WalletService,
) pb.WalletUnlockerServer {
	return newWalletUnlockerHandler(walletSvc)
}

func newWalletUnlockerHandler(
	walletSvc application.WalletService,
) *walletUnlockerHandler {
	return &walletUnlockerHandler{
		walletSvc: walletSvc,
	}
}

func (w walletUnlockerHandler) GenSeed(
	ctx context.Context,
	req *pb.GenSeedRequest,
) (*pb.GenSeedReply, error) {
	return w.genSeed(ctx, req)
}

func (w walletUnlockerHandler) InitWallet(
	req *pb.InitWalletRequest,
	stream pb.WalletUnlocker_InitWalletServer,
) error {
	return w.initWallet(req, stream)
}

func (w walletUnlockerHandler) UnlockWallet(
	ctx context.Context,
	req *pb.UnlockWalletRequest,
) (*pb.UnlockWalletReply, error) {
	return w.unlockWallet(ctx, req)
}

func (w walletUnlockerHandler) ChangePassword(
	ctx context.Context,
	req *pb.ChangePasswordRequest,
) (*pb.ChangePasswordReply, error) {
	return w.changePassword(ctx, req)
}

func (w walletUnlockerHandler) genSeed(
	ctx context.Context,
	req *pb.GenSeedRequest,
) (*pb.GenSeedReply, error) {
	mnemonic, err := w.walletSvc.GenSeed(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GenSeedReply{SeedMnemonic: mnemonic}, nil
}

func (w walletUnlockerHandler) initWallet(
	req *pb.InitWalletRequest,
	stream pb.WalletUnlocker_InitWalletServer,
) error {
	mnemonic := req.GetSeedMnemonic()
	if err := validateMnemonic(mnemonic); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	password := req.GetWalletPassword()
	if err := validatePassword(password); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	chReplies := make(chan *application.InitWalletReply)
	chErr := make(chan error, 1)
	go w.walletSvc.InitWallet(
		stream.Context(),
		mnemonic,
		string(password),
		req.GetRestore(),
		chReplies,
		chErr,
	)

	for {
		select {
		case err := <-chErr:
			return err
		case reply, ok := <-chReplies:
			if !ok {
				return nil
			}
			if err := stream.Send(&pb.InitWalletReply{
				Account: uint64(reply.AccountIndex),
				Index:   uint64(reply.AddressIndex),
				Status:  pb.InitWalletReply_Status(reply.Status),
				Data:    reply.Data,
			}); err != nil {
				return err
			}
		}
	}
}

func (w walletUnlockerHandler) unlockWallet(
	ctx context.Context,
	req *pb.UnlockWalletRequest,
) (*pb.UnlockWalletReply, error) {
	password := req.GetWalletPassword()
	if err := validatePassword(password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := w.walletSvc.UnlockWallet(ctx, string(password)); err != nil {
		return nil, err
	}

	return &pb.UnlockWalletReply{}, nil
}

func (w walletUnlockerHandler) changePassword(
	ctx context.Context,
	req *pb.ChangePasswordRequest,
) (*pb.ChangePasswordReply, error) {
	currentPwd := req.GetCurrentPassword()
	if err := validatePassword(currentPwd); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	newPwd := req.GetNewPassword()
	if err := validatePassword(newPwd); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := w.walletSvc.ChangePassword(
		ctx,
		string(currentPwd),
		string(newPwd),
	); err != nil {
		return nil, err
	}

	return &pb.ChangePasswordReply{}, nil

}

func validateMnemonic(mnemonic []string) error {
	if len(mnemonic) <= 0 {
		return errors.New("mnemonic is null")
	}
	return nil
}

func validatePassword(password []byte) error {
	if len(password) <= 0 {
		return errors.New("password is null")
	}
	return nil
}
