package walletservice

import (
	"context"
	"encoding/hex"

	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WalletAddress returns a new address for the wallet account
func (s *Service) WalletAddress(ctx context.Context, req *pb.WalletAddressRequest) (reply *pb.WalletAddressReply, err error) {
	if err = s.vaultRepository.UpdateVault(ctx, nil, "", func(v *vault.Vault) (*vault.Vault, error) {
		addr, _, blindingKey, err := v.DeriveNextExternalAddressForAccount(vault.WalletAccount)
		if err != nil {
			return nil, err
		}

		reply = &pb.WalletAddressReply{
			Address:  addr,
			Blinding: hex.EncodeToString(blindingKey),
		}

		return v, nil
	}); err != nil {
		err = status.Error(codes.Internal, err.Error())
		return
	}
	return
}
