package grpchandler

import (
	"context"
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex-daemon/v1"
)

type walletHandler struct {
	walletSvc application.WalletService
}

func NewWalletHandler(
	walletSvc application.WalletService,
) daemonv1.WalletServiceServer {
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
	req *daemonv1.WalletAddressRequest,
) (*daemonv1.WalletAddressResponse, error) {
	return w.walletAddress(ctx, req)
}

func (w walletHandler) WalletBalance(
	ctx context.Context,
	req *daemonv1.WalletBalanceRequest,
) (*daemonv1.WalletBalanceResponse, error) {
	return w.walletBalance(ctx, req)
}

func (w walletHandler) SendToMany(
	ctx context.Context,
	req *daemonv1.SendToManyRequest,
) (*daemonv1.SendToManyResponse, error) {
	return w.sendToMany(ctx, req)
}

func (w walletHandler) walletAddress(
	ctx context.Context,
	req *daemonv1.WalletAddressRequest,
) (*daemonv1.WalletAddressResponse, error) {
	addr, blindingKey, err := w.walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		return nil, err
	}

	return &daemonv1.WalletAddressResponse{
		Address:  addr,
		Blinding: blindingKey,
	}, nil
}

func (w walletHandler) walletBalance(
	ctx context.Context,
	req *daemonv1.WalletBalanceRequest,
) (*daemonv1.WalletBalanceResponse, error) {
	b, err := w.walletSvc.GetWalletBalance(ctx)
	if err != nil {
		return nil, err
	}

	balance := make(map[string]*daemonv1.BalanceInfo)
	for k, v := range b {
		balance[k] = &daemonv1.BalanceInfo{
			TotalBalance:       v.TotalBalance,
			ConfirmedBalance:   v.ConfirmedBalance,
			UnconfirmedBalance: v.UnconfirmedBalance,
		}
	}

	return &daemonv1.WalletBalanceResponse{Balance: balance}, nil
}

func (w walletHandler) sendToMany(
	ctx context.Context,
	req *daemonv1.SendToManyRequest,
) (*daemonv1.SendToManyResponse, error) {
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
		outputs = append(outputs, application.TxOut{
			Asset:   v.GetAsset(),
			Value:   v.GetValue(),
			Address: v.GetAddress(),
		})
	}

	walletReq := application.SendToManyRequest{
		Outputs:         outputs,
		MillisatPerByte: msatPerByte,
		Push:            true,
	}
	rawTx, txid, err := w.walletSvc.SendToMany(ctx, walletReq)
	if err != nil {
		return nil, err
	}

	return &daemonv1.SendToManyResponse{RawTx: rawTx, Txid: txid}, nil
}

func validateOutputs(outputs []*daemonv1.TxOutput) error {
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
