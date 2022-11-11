package oceanwallet

import (
	"context"
	"strings"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/ocean/v1"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"google.golang.org/grpc"
)

type wallet struct {
	client pb.WalletServiceClient
}

func newWallet(conn *grpc.ClientConn) *wallet {
	return &wallet{pb.NewWalletServiceClient(conn)}
}

func (m *wallet) GenSeed(ctx context.Context) ([]string, error) {
	res, err := m.client.GenSeed(ctx, &pb.GenSeedRequest{})
	if err != nil {
		return nil, err
	}
	mnemonic := strings.Split(res.GetMnemonic(), " ")

	return mnemonic, nil
}

func (m *wallet) InitWallet(
	ctx context.Context, mnemonic []string, password string,
) error {
	_, err := m.client.CreateWallet(ctx, &pb.CreateWalletRequest{
		Mnemonic: strings.Join(mnemonic, " "),
		Password: password,
	})
	return err
}

func (m *wallet) RestoreWallet(
	ctx context.Context, mnemonic []string, password string,
) error {
	_, err := m.client.RestoreWallet(ctx, &pb.RestoreWalletRequest{
		Mnemonic: strings.Join(mnemonic, " "),
		Password: password,
	})
	return err
}

func (m *wallet) Unlock(ctx context.Context, password string) error {
	_, err := m.client.Unlock(ctx, &pb.UnlockRequest{
		Password: password,
	})
	return err
}

func (m *wallet) Lock(ctx context.Context, password string) error {
	_, err := m.client.Lock(ctx, &pb.LockRequest{
		Password: password,
	})
	return err
}

func (m *wallet) ChangePassword(
	ctx context.Context, oldPwd, newPwd string,
) error {
	_, err := m.client.ChangePassword(ctx, &pb.ChangePasswordRequest{
		CurrentPassword: oldPwd,
		NewPassword:     newPwd,
	})
	return err
}

func (m *wallet) Status(
	ctx context.Context,
) (ports.WalletStatus, error) {
	res, err := m.client.Status(ctx, &pb.StatusRequest{})
	if err != nil {
		return nil, err
	}
	return walletStatus{res}, nil
}

func (m *wallet) Info(ctx context.Context) (ports.WalletInfo, error) {
	res, err := m.client.GetInfo(ctx, &pb.GetInfoRequest{})
	if err != nil {
		return nil, err
	}
	return walletInfo{res}, nil
}

type walletStatus struct {
	*pb.StatusResponse
}

func (w walletStatus) IsInitialized() bool {
	return w.StatusResponse.GetInitialized()
}
func (w walletStatus) IsUnlocked() bool {
	return w.StatusResponse.GetUnlocked()
}
func (w walletStatus) IsSynced() bool {
	return w.StatusResponse.GetSynced()
}

type walletInfo struct {
	*pb.GetInfoResponse
}

func (w walletInfo) GetNetwork() string {
	switch w.GetInfoResponse.GetNetwork() {
	case pb.GetInfoResponse_NETWORK_REGTEST:
		return "regtest"
	case pb.GetInfoResponse_NETWORK_TESTNET:
		return "testnet"
	case pb.GetInfoResponse_NETWORK_MAINNET:
		fallthrough
	default:
		return "mainnet"
	}
}

func (w walletInfo) GetAccounts() []ports.WalletAccount {
	accountInfo := w.GetInfoResponse.GetAccounts()
	accounts := make([]ports.WalletAccount, 0, len(accountInfo))
	for _, i := range accountInfo {
		accounts = append(accounts, i)
	}
	return accounts
}
