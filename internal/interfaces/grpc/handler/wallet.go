package grpchandler

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type walletHandler struct {
	walletSvc    ports.Wallet
	buildData    ports.BuildData
	macaroonPath string

	onInit      func(pwd string)
	onUnlock    func(pwd string)
	onLock      func(pwd string)
	onChangePwd func(oldPwd, newPwd string)
}

func NewWalletHandler(
	walletSvc ports.Wallet, buildData ports.BuildData, macPath string,
	onInit, onUnlock, onLock func(pwd string),
	onChangePwd func(oldPwd, newPwd string),
) daemonv2.WalletServiceServer {
	return newWalletHandler(
		walletSvc, buildData, macPath, onInit, onUnlock, onLock, onChangePwd,
	)
}

func newWalletHandler(
	walletSvc ports.Wallet, buildData ports.BuildData, macPath string,
	onInit, onUnlock, onLock func(pwd string),
	onChangePwd func(oldPwd, newPwd string),
) *walletHandler {
	return &walletHandler{
		walletSvc, buildData, macPath, onInit, onUnlock, onLock, onChangePwd,
	}
}

func (h *walletHandler) GenSeed(
	ctx context.Context, _ *daemonv2.GenSeedRequest,
) (*daemonv2.GenSeedResponse, error) {
	mnemonic, err := h.walletSvc.GenSeed(ctx)
	if err != nil {
		return nil, err
	}
	return &daemonv2.GenSeedResponse{
		SeedMnemonic: mnemonic,
	}, nil
}

func (h *walletHandler) InitWallet(
	req *daemonv2.InitWalletRequest,
	stream daemonv2.WalletService_InitWalletServer,
) error {
	password, err := parsePassword(req.GetPassword())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	mnemonic, err := parseMnemonic(req.GetSeedMnemonic())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	isRestore := req.GetRestore()
	ctx := stream.Context()

	if isRestore {
		if err := h.walletSvc.RestoreWallet(ctx, mnemonic, password); err != nil {
			return err
		}
	}
	if err := h.walletSvc.InitWallet(ctx, mnemonic, password); err != nil {
		return err
	}

	go h.onInit(password)

	if len(h.macaroonPath) <= 0 {
		return nil
	}

	for {
		if _, err := os.Stat(h.macaroonPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		mac, err := ioutil.ReadFile(h.macaroonPath)
		if err != nil {
			return err
		}
		if len(mac) > 0 {
			return stream.Send(&daemonv2.InitWalletResponse{
				Message: fmt.Sprintf("macaroon: %s", hex.EncodeToString(mac)),
			})
		}
	}
}

func (h *walletHandler) UnlockWallet(
	ctx context.Context, req *daemonv2.UnlockWalletRequest,
) (*daemonv2.UnlockWalletResponse, error) {
	password, err := parsePassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := h.walletSvc.Unlock(ctx, password); err != nil {
		return nil, err
	}

	go h.onUnlock(password)

	return &daemonv2.UnlockWalletResponse{}, nil
}

func (h *walletHandler) LockWallet(
	ctx context.Context, req *daemonv2.LockWalletRequest,
) (*daemonv2.LockWalletResponse, error) {
	password, err := parsePassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := h.walletSvc.Lock(ctx, password); err != nil {
		return nil, err
	}

	go h.onLock(password)

	return &daemonv2.LockWalletResponse{}, nil
}

func (h *walletHandler) ChangePassword(
	ctx context.Context, req *daemonv2.ChangePasswordRequest,
) (*daemonv2.ChangePasswordResponse, error) {
	currentPassword, err := parsePassword(req.GetCurrentPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	newPassword, err := parsePassword(req.GetNewPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := h.walletSvc.ChangePassword(
		ctx, currentPassword, newPassword,
	); err != nil {
		return nil, err
	}

	go h.onChangePwd(currentPassword, newPassword)

	return &daemonv2.ChangePasswordResponse{}, nil
}

func (h *walletHandler) GetInfo(
	ctx context.Context, _ *daemonv2.GetInfoRequest,
) (*daemonv2.GetInfoResponse, error) {
	info, err := h.walletSvc.Info(ctx)
	if err != nil {
		return nil, err
	}
	return walletInfo{info, h.buildData}.toProto(), nil
}

func (h *walletHandler) GetStatus(
	ctx context.Context, _ *daemonv2.GetStatusRequest,
) (*daemonv2.GetStatusResponse, error) {
	status, err := h.walletSvc.Status(ctx)
	if err != nil {
		return nil, err
	}
	return walletStatusInfo{status}.toProto(), nil
}
