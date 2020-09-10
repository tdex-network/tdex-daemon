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
func (s *Service) WalletBalance(ctx context.Context, req *pb.WalletBalanceRequest) (*pb.WalletBalanceResponse, error) {
	derivedAddresses, prvBlindingKeys, err := s.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, vault.WalletAccount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	unspents, err := s.getUnspents(derivedAddresses, prvBlindingKeys)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.WalletBalanceResponse{
		Balance: getBalancesByAsset(unspents),
	}, nil
}

func getBalancesByAsset(unspents []explorer.Utxo) (balance map[string]*pb.BalanceInfo) {
	for _, unspent := range unspents {
		balance[unspent.Asset()].TotalBalance += unspent.Value()
		if unspent.IsConfirmed() {
			balance[unspent.Asset()].ConfirmedBalance += unspent.Value()
		} else {
			balance[unspent.Asset()].UnconfirmedBalance += unspent.Value()
		}
	}

	return
}
