package oceanwallet

import (
	"context"
	"strings"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/ocean/v1"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"google.golang.org/grpc"
)

type account struct {
	client pb.AccountServiceClient
}

func newAccount(conn *grpc.ClientConn) *account {
	return &account{pb.NewAccountServiceClient(conn)}
}

func (m *account) CreateAccount(
	ctx context.Context, accountName string, isMarket bool,
) (ports.WalletAccount, error) {
	// If the account being created is for market, try to create a multisig
	// account if supported, otherwise fallback to singlesig one.
	if isMarket {
		res, err := m.client.CreateAccountMultiSig(ctx, &pb.CreateAccountMultiSigRequest{
			Label: accountName,
		})
		if err != nil && !strings.Contains(err.Error(), "not implemented") {
			return nil, err
		}
		if res != nil {
			return accountInfo{res.GetInfo()}, nil
		}
	}

	res, err := m.client.CreateAccountBIP44(ctx, &pb.CreateAccountBIP44Request{
		Label: accountName,
	})
	if err != nil {
		return nil, err
	}
	return accountInfo{res.GetInfo()}, nil
}

func (m *account) DeriveAddresses(
	ctx context.Context, accountName string, num int,
) ([]string, error) {
	res, err := m.client.DeriveAddresses(ctx, &pb.DeriveAddressesRequest{
		AccountName:    accountName,
		NumOfAddresses: uint64(num),
	})
	if err != nil {
		return nil, err
	}
	return res.GetAddresses(), nil
}

func (m *account) DeriveChangeAddresses(
	ctx context.Context, accountName string, num int,
) ([]string, error) {
	res, err := m.client.DeriveChangeAddresses(ctx, &pb.DeriveChangeAddressesRequest{
		AccountName:    accountName,
		NumOfAddresses: uint64(num),
	})
	if err != nil {
		return nil, err
	}
	return res.GetAddresses(), nil
}

func (m *account) ListAddresses(
	ctx context.Context, accountName string,
) ([]string, error) {
	res, err := m.client.ListAddresses(ctx, &pb.ListAddressesRequest{
		AccountName: accountName,
	})
	if err != nil {
		return nil, err
	}
	return res.GetAddresses(), nil
}

func (m *account) GetBalance(
	ctx context.Context, accountName string,
) (map[string]ports.Balance, error) {
	res, err := m.client.Balance(ctx, &pb.BalanceRequest{
		AccountName: accountName,
	})
	if err != nil {
		return nil, err
	}
	balance := make(map[string]ports.Balance)
	for asset, bal := range res.GetBalance() {
		balance[asset] = bal
	}
	return balance, nil
}

func (m *account) ListUtxos(
	ctx context.Context, accountName string,
) (spendableUtxos, lockedUtxos []ports.Utxo, err error) {
	res, err := m.client.ListUtxos(ctx, &pb.ListUtxosRequest{
		AccountName: accountName,
	})
	if err != nil {
		return nil, nil, err
	}
	if res.GetSpendableUtxos() != nil {
		spendableUtxos = utxoList(res.GetSpendableUtxos().GetUtxos()).toPortableList()
	}
	if res.GetLockedUtxos() != nil {
		lockedUtxos = utxoList(res.GetLockedUtxos().GetUtxos()).toPortableList()
	}
	return
}

func (m *account) DeleteAccount(
	ctx context.Context, accountName string,
) error {
	_, err := m.client.DeleteAccount(ctx, &pb.DeleteAccountRequest{
		AccountName: accountName,
	})
	return err
}

type accountInfo struct {
	*pb.AccountInfo
}

func (i accountInfo) GetMarket() ports.Market {
	return nil
}

type utxoList []*pb.Utxo

func (l utxoList) toPortableList() []ports.Utxo {
	utxos := make([]ports.Utxo, 0, len(l))
	for _, u := range l {
		utxos = append(utxos, utxoInfo{u})
	}
	return utxos
}
