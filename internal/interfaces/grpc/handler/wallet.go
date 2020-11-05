package grpchandler

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
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
		log.Debug("trying to generate new seed: ", err)
		return nil, status.Error(codes.Internal, ErrCannotServeRequest)
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

	for {
		tx := w.dbManager.NewTransaction()
		ctx := context.WithValue(stream.Context(), "tx", tx)

		if err := w.walletSvc.InitWallet(
			ctx,
			mnemonic,
			string(password),
		); err != nil {
			log.Debug("trying to initialize wallet: ", err)
			return status.Error(codes.Internal, ErrCannotServeRequest)
		}

		if err := tx.Commit(); err != nil {
			if !w.dbManager.IsTransactionConflict(err) {
				log.Debug("trying to commit changes after initializing wallet: ", err)
				return status.Error(codes.Internal, ErrCannotServeRequest)
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}
		break
	}

	if err := stream.Send(&pb.InitWalletReply{}); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (w walletHandler) unlockWallet(
	reqCtx context.Context,
	req *pb.UnlockWalletRequest,
) (*pb.UnlockWalletReply, error) {
	password := req.GetWalletPassword()
	if err := validatePassword(password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	for {
		tx := w.dbManager.NewTransaction()
		ctx := context.WithValue(reqCtx, "tx", tx)

		if err := w.walletSvc.UnlockWallet(ctx, string(password)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if err := tx.Commit(); err != nil {
			if !w.dbManager.IsTransactionConflict(err) {
				log.Debug("trying to commit changes after unlocking wallet: ", err)
				return nil, status.Error(codes.Internal, ErrCannotServeRequest)
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}
		break
	}

	return &pb.UnlockWalletReply{}, nil
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

	for {
		tx := w.dbManager.NewTransaction()
		ctx := context.WithValue(reqCtx, "tx", tx)

		if err := w.walletSvc.ChangePassword(
			ctx,
			string(currentPwd),
			string(newPwd),
		); err != nil {
			log.Debug("trying to change password: ", err)
			return nil, status.Error(codes.Internal, ErrCannotServeRequest)
		}

		if err := tx.Commit(); err != nil {
			if !w.dbManager.IsTransactionConflict(err) {
				log.Debug("trying to commit changes after unlocking wallet: ", err)
				return nil, status.Error(codes.Internal, ErrCannotServeRequest)
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}
		break
	}

	return &pb.ChangePasswordReply{}, nil
}

func (w walletHandler) walletAddress(
	reqCtx context.Context,
	req *pb.WalletAddressRequest,
) (*pb.WalletAddressReply, error) {
	var reply *pb.WalletAddressReply

	for {
		tx := w.dbManager.NewTransaction()
		ctx := context.WithValue(reqCtx, "tx", tx)

		addr, blindingKey, err := w.walletSvc.GenerateAddressAndBlindingKey(ctx)
		if err != nil {
			log.Debug("trying to derive new address: ", err)
			return nil, status.Error(codes.Internal, ErrCannotServeRequest)
		}

		if err := tx.Commit(); err != nil {
			if !w.dbManager.IsTransactionConflict(err) {
				log.Debug("trying to commit changes after deriving new address: ", err)
				return nil, status.Error(codes.Internal, ErrCannotServeRequest)
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}

		reply = &pb.WalletAddressReply{
			Address:  addr,
			Blinding: blindingKey,
		}
		break
	}

	return reply, nil
}

func (w walletHandler) walletBalance(
	ctx context.Context,
	req *pb.WalletBalanceRequest,
) (*pb.WalletBalanceReply, error) {
	b, err := w.walletSvc.GetWalletBalance(ctx)
	if err != nil {
		log.Debug("trying to get balance: ", err)
		return nil, status.Error(codes.Internal, ErrCannotServeRequest)
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

	var reply *pb.SendToManyReply
	for {
		tx := w.dbManager.NewTransaction()
		ctx := context.WithValue(reqCtx, "tx", tx)

		walletReq := application.SendToManyRequest{
			Outputs:         outputs,
			MillisatPerByte: msatPerByte,
			Push:            req.GetPush(),
		}
		rawTx, err := w.walletSvc.SendToMany(ctx, walletReq)
		if err != nil {
			log.Debug("trying to send to many: ", err)
			return nil, status.Error(codes.Internal, ErrCannotServeRequest)
		}

		if err := tx.Commit(); err != nil {
			if !w.dbManager.IsTransactionConflict(err) {
				log.Debug("trying to commit changes after sending to many: ", err)
				return nil, status.Error(codes.Internal, ErrCannotServeRequest)
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}

		reply = &pb.SendToManyReply{RawTx: rawTx}
		break
	}

	return reply, nil
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
