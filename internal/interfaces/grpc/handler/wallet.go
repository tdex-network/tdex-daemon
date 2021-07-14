package grpchandler

import (
	"context"
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/wallet"
)

type walletHandler struct {
	pb.UnimplementedWalletServer
	walletSvc application.WalletService
}

func NewWalletHandler(
	walletSvc application.WalletService,
) pb.WalletServer {
	return newWalletHandler(walletSvc)
}

func newWalletHandler(
	walletSvc application.WalletService,
) *walletHandler {
	return &walletHandler{
		walletSvc: walletSvc,
	}
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

func (w walletHandler) walletAddress(
	ctx context.Context,
	req *pb.WalletAddressRequest,
) (*pb.WalletAddressReply, error) {
	addr, blindingKey, err := w.walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.WalletAddressReply{
		Address:  addr,
		Blinding: blindingKey,
	}, nil
}

func (w walletHandler) walletBalance(
	ctx context.Context,
	req *pb.WalletBalanceRequest,
) (*pb.WalletBalanceReply, error) {
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
}

func (w walletHandler) sendToMany(
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
		Push:            true,
	}
	rawTx, err := w.walletSvc.SendToMany(ctx, walletReq)
	if err != nil {
		return nil, err
	}

	return &pb.SendToManyReply{RawTx: rawTx}, nil
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
