package grpchandler

import (
	"context"
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type walletHandler struct {
	pb.UnimplementedWalletServer
	walletSvc application.WalletService
	dbManager ports.DbManager
}

func NewWalletHandler(
	walletSvc application.WalletService,
	dbManager ports.DbManager,
) pb.WalletServer {
	return newWalletHandler(walletSvc, dbManager)
}

func newWalletHandler(
	walletSvc application.WalletService,
	dbManager ports.DbManager,
) *walletHandler {
	return &walletHandler{
		walletSvc: walletSvc,
		dbManager: dbManager,
	}
}

func (w walletHandler) GenSeed(
	ctx context.Context,
	req *pb.GenSeedRequest,
) (*pb.GenSeedReply, error) {
	return w.genSeed(ctx, req)
}

func (w walletHandler) InitWallet(
	req *pb.InitWalletRequest,
	stream pb.Wallet_InitWalletServer,
) error {
	return w.initWallet(req, stream)
}

func (w walletHandler) UnlockWallet(
	ctx context.Context,
	req *pb.UnlockWalletRequest,
) (*pb.UnlockWalletReply, error) {
	return w.unlockWallet(ctx, req)
}

func (w walletHandler) ChangePassword(
	ctx context.Context,
	req *pb.ChangePasswordRequest,
) (*pb.ChangePasswordReply, error) {
	return w.changePassword(ctx, req)
}

func (w walletHandler) WalletAddress(
	ctx context.Context,
	req *pb.WalletAddressRequest,
) (*pb.WalletAddressReply, error) {
	return w.walletAddress(ctx, req)
}

func (w walletHandler) WalletBalance(
	ctx context.Context,
	req *pb.WalletBalanceRequest,
) (*pb.WalletBalanceReply, error) {
	return w.walletBalance(ctx, req)
}

func (w walletHandler) SendToMany(
	ctx context.Context,
	req *pb.SendToManyRequest,
) (*pb.SendToManyReply, error) {
	return w.sendToMany(ctx, req)
}

func (w walletHandler) genSeed(
	ctx context.Context,
	req *pb.GenSeedRequest,
) (*pb.GenSeedReply, error) {
	mnemonic, err := w.walletSvc.GenSeed(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GenSeedReply{SeedMnemonic: mnemonic}, nil
}

func (w walletHandler) initWallet(
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
			close(chErr)
			close(chReplies)
			return err
		case reply := <-chReplies:
			if reply == nil {
				close(chErr)
				close(chReplies)
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

func (w walletHandler) unlockWallet(
	reqCtx context.Context,
	req *pb.UnlockWalletRequest,
) (*pb.UnlockWalletReply, error) {
	password := req.GetWalletPassword()
	if err := validatePassword(password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := w.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			if err := w.walletSvc.UnlockWallet(ctx, string(password)); err != nil {
				return nil, err
			}

			return &pb.UnlockWalletReply{}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.UnlockWalletReply), nil
}

func (w walletHandler) changePassword(
	reqCtx context.Context,
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

	res, err := w.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			if err := w.walletSvc.ChangePassword(
				ctx,
				string(currentPwd),
				string(newPwd),
			); err != nil {
				return nil, err
			}

			return &pb.ChangePasswordReply{}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.ChangePasswordReply), nil
}

func (w walletHandler) walletAddress(
	reqCtx context.Context,
	req *pb.WalletAddressRequest,
) (*pb.WalletAddressReply, error) {
	res, err := w.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			addr, blindingKey, err := w.walletSvc.GenerateAddressAndBlindingKey(ctx)
			if err != nil {
				return nil, err
			}

			return &pb.WalletAddressReply{
				Address:  addr,
				Blinding: blindingKey,
			}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.WalletAddressReply), nil
}

func (w walletHandler) walletBalance(
	reqCtx context.Context,
	req *pb.WalletBalanceRequest,
) (*pb.WalletBalanceReply, error) {
	res, err := w.dbManager.RunTransaction(
		reqCtx,
		readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			b, err := w.walletSvc.GetWalletBalance(ctx)
			if err != nil {
				return nil, err
			}

			balance := make(map[string]*pb.BalanceInfo)
			for k, v := range b {
				balance[k] = &pb.BalanceInfo{
					TotalBalance:       v.TotalBalance,
					ConfirmedBalance:   v.ConfirmedBalance,
					UnconfirmedBalance: v.UnconfirmedBalance,
				}
			}

			return &pb.WalletBalanceReply{Balance: balance}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.WalletBalanceReply), nil
}

func (w walletHandler) sendToMany(
	reqCtx context.Context,
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

	res, err := w.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			walletReq := application.SendToManyRequest{
				Outputs:         outputs,
				MillisatPerByte: msatPerByte,
				Push:            true,
			}
			rawTx, err := w.walletSvc.SendToMany(ctx, walletReq)
			if err != nil {
				return nil, err
			}

			return &pb.SendToManyReply{RawTx: rawTx}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.SendToManyReply), nil
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
