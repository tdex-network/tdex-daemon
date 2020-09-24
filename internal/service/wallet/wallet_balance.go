package walletservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WalletBalance returns info about the balance or balances of the wallet
// account.
// For each asset owned by the wallet account, the balance info reports both
// the confirmed and unconfirmed balances and the total balance that is their
// sum
func (s *Service) WalletBalance(ctx context.Context, req *pb.WalletBalanceRequest) (*pb.WalletBalanceReply, error) {
	derivedAddresses, prvBlindingKeys, err := s.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, vault.WalletAccount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	unspents, err := s.getUnspents(derivedAddresses, prvBlindingKeys)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.WalletBalanceReply{
		Balance: getBalancesByAsset(unspents),
	}, nil
}

func getBalancesByAsset(unspents []explorer.Utxo) map[string]*pb.BalanceInfo {
	balances := map[string]*pb.BalanceInfo{}
	for _, unspent := range unspents {
		if _, ok := balances[unspent.Asset()]; !ok {
			balances[unspent.Asset()] = &pb.BalanceInfo{}
		}

		balance := balances[unspent.Asset()]
		balance.TotalBalance += unspent.Value()
		if unspent.IsConfirmed() {
			balance.ConfirmedBalance += unspent.Value()
		} else {
			balance.UnconfirmedBalance += unspent.Value()
		}
		balances[unspent.Asset()] = balance
	}
	return balances
}
