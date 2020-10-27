package grpchandler

import (
	"context"
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
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
	mnemonic := req.GetSeedMnemonic()
	if err := validateMnemonic(mnemonic); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	password := req.GetWalletPassword()
	if err := validatePassword(password); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if err := w.walletSvc.InitWallet(
		stream.Context(),
		mnemonic,
		string(password),
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
	password := req.GetWalletPassword()
	if err := validatePassword(password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := w.walletSvc.UnlockWallet(ctx, string(password)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UnlockWalletReply{}, nil
}

func (w walletHandler) ChangePassword(
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
	b, err := w.walletSvc.GetWalletBalance(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	balance := make(map[string]*pb.BalanceInfo)
	for k, v := range b {
		balance[k] = &pb.BalanceInfo{
			TotalBalance:       v.TotalBalance,
			ConfirmedBalance:   v.ConfirmedBalance,
			UnconfirmedBalance: v.UnconfirmedBalance,
		}
	}

	return &pb.WalletBalanceReply{
		Balance: balance,
	}, nil
}

func (w walletHandler) SendToMany(
	ctx context.Context,
	req *pb.SendToManyRequest,
) (*pb.SendToManyReply, error) {
	outs := req.GetOutputs()
	if err := validateOutputs(outs); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	msatPerByte := req.GetMillisatPerByte()
	if err := validateMillisatPerByte(msatPerByte); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	outputs := make([]application.TxOut, 0)
	for _, v := range outs {
		output := application.TxOut{
			Asset:   v.GetAsset(),
			Value:   v.GetValue(),
			Address: v.GetAddress(),
		}
		outputs = append(outputs, output)
	}

	walletReq := application.SendToManyRequest{
		Outputs:         outputs,
		MillisatPerByte: msatPerByte,
		Push:            req.GetPush(),
	}
	rawTx, err := w.walletSvc.SendToMany(ctx, walletReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SendToManyReply{
		RawTx: rawTx,
	}, nil
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

func validateOutputs(outputs []*pb.TxOut) error {
	if len(outputs) <= 0 {
		return errors.New("output list is empty")
	}
	for _, o := range outputs {
		if o == nil ||
			len(o.GetAsset()) <= 0 ||
			o.GetValue() <= 0 ||
			len(o.GetAddress()) <= 0 {
			return errors.New("output list is malformed")
		}
	}
	return nil
}

func validateMillisatPerByte(satPerByte int64) error {
	if satPerByte < domain.MinMilliSatPerByte {
		return errors.New("milli sats per byte is too low")
	}
	return nil
}
