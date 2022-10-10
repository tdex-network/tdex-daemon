package grpchandler

import (
	"context"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
)

type walletUnlockerHandler struct {
	walletUnlockerSvc application.WalletUnlockerService
	adminMacaroonPath string
}

func NewWalletUnlockerHandler(
	walletUnlockerSvc application.WalletUnlockerService,
	adminMacPath string,
) daemonv1.WalletUnlockerServiceServer {
	return newWalletUnlockerHandler(walletUnlockerSvc, adminMacPath)
}

func newWalletUnlockerHandler(
	walletUnlockerSvc application.WalletUnlockerService,
	adminMacPath string,
) *walletUnlockerHandler {
	return &walletUnlockerHandler{
		walletUnlockerSvc: walletUnlockerSvc,
		adminMacaroonPath: adminMacPath,
	}
}

func (w *walletUnlockerHandler) GenSeed(
	ctx context.Context, req *daemonv1.GenSeedRequest,
) (*daemonv1.GenSeedResponse, error) {
	return w.genSeed(ctx, req)
}

func (w *walletUnlockerHandler) InitWallet(
	req *daemonv1.InitWalletRequest, stream daemonv1.WalletUnlockerService_InitWalletServer,
) error {
	return w.initWallet(req, stream)
}

func (w *walletUnlockerHandler) UnlockWallet(
	ctx context.Context, req *daemonv1.UnlockWalletRequest,
) (*daemonv1.UnlockWalletResponse, error) {
	return w.unlockWallet(ctx, req)
}

func (w *walletUnlockerHandler) ChangePassword(
	ctx context.Context, req *daemonv1.ChangePasswordRequest,
) (*daemonv1.ChangePasswordResponse, error) {
	return w.changePassword(ctx, req)
}

func (w *walletUnlockerHandler) IsReady(
	ctx context.Context, req *daemonv1.IsReadyRequest,
) (*daemonv1.IsReadyResponse, error) {
	return w.isReady(ctx, req)
}

func (w *walletUnlockerHandler) genSeed(
	ctx context.Context, req *daemonv1.GenSeedRequest,
) (*daemonv1.GenSeedResponse, error) {
	mnemonic, err := w.walletUnlockerSvc.GenSeed(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &daemonv1.GenSeedResponse{SeedMnemonic: mnemonic}, nil
}

func (w *walletUnlockerHandler) initWallet(
	req *daemonv1.InitWalletRequest, stream daemonv1.WalletUnlockerService_InitWalletServer,
) error {
	mnemonic := req.GetSeedMnemonic()
	if err := validateMnemonic(mnemonic); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	password := req.GetWalletPassword()
	if err := validatePassword(password); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	chReplies := make(chan application.InitWalletReply)
	go w.walletUnlockerSvc.InitWallet(
		stream.Context(),
		mnemonic,
		string(password),
		req.GetRestore(),
		chReplies,
	)

	noReplies := true
	for reply := range chReplies {
		if err := reply.Err; err != nil {
			return err
		}

		noReplies = false
		if err := stream.Send(&daemonv1.InitWalletResponse{
			Account: reply.AccountIndex,
			Status:  daemonv1.InitWalletResponse_Status(reply.Status),
			Data:    reply.Data,
		}); err != nil {
			return err
		}
	}

	// Inject admin.macaroon to InitWalletResponse only if the app service has
	// actually initialized the internal wallet.
	// If the reply channel didn't contain any message before closing, it means
	// that the app service skipped the operation because the wallet was already
	// initialized.
	if !noReplies && w.adminMacaroonPath != "" {
		var mac []byte
		// Retry checking the admin.macaroon file exists until it's found in the datadir.
	macExist:
		for {
			if _, err := os.Stat(w.adminMacaroonPath); err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return err
			}
			break
		}
		mac, err := ioutil.ReadFile(w.adminMacaroonPath)
		if err != nil {
			return nil
		}

		if len(mac) > 0 {
			macStr := hex.EncodeToString(mac)
			if err := stream.Send(&daemonv1.InitWalletResponse{
				Data:    macStr,
				Account: -1,
			}); err != nil {
				return err
			}
		} else {
			goto macExist
		}
	}

	return nil
}

func (w *walletUnlockerHandler) unlockWallet(
	ctx context.Context, req *daemonv1.UnlockWalletRequest,
) (*daemonv1.UnlockWalletResponse, error) {
	password := req.GetWalletPassword()
	if err := validatePassword(password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := w.walletUnlockerSvc.UnlockWallet(ctx, string(password)); err != nil {
		return nil, err
	}

	return &daemonv1.UnlockWalletResponse{}, nil
}

func (w *walletUnlockerHandler) changePassword(
	ctx context.Context, req *daemonv1.ChangePasswordRequest,
) (*daemonv1.ChangePasswordResponse, error) {
	currentPwd := req.GetCurrentPassword()
	if err := validatePassword(currentPwd); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	newPwd := req.GetNewPassword()
	if err := validatePassword(newPwd); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := w.walletUnlockerSvc.ChangePassword(
		ctx,
		string(currentPwd),
		string(newPwd),
	); err != nil {
		return nil, err
	}

	return &daemonv1.ChangePasswordResponse{}, nil
}

func (w *walletUnlockerHandler) isReady(
	ctx context.Context, _ *daemonv1.IsReadyRequest,
) (*daemonv1.IsReadyResponse, error) {
	status := w.walletUnlockerSvc.IsReady(ctx)

	return &daemonv1.IsReadyResponse{
		Initialized: status.Initialized,
		Unlocked:    status.Unlocked,
		Synced:      status.Synced,
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
