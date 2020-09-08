package walletservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InitWallet creates or restores a wallet for the daemon
func (s *Service) InitWallet(ctx context.Context, req *pb.InitWalletRequest) (*pb.InitWalletResponse, error) {
	// var mnemonic string
	var err error

	if mnemonics := req.GetSeedMnemonic(); len(mnemonics) > 0 {
		_, _, err = s.vaultRepository.CreateOrRestoreVault(ctx, mnemonics[0])
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	} else {
		_, _, err = s.vaultRepository.CreateOrRestoreVault(ctx, "")
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if passphrase := string(req.GetWalletPassword()); len(passphrase) > 0 {
		err = s.vaultRepository.UpdateVault(ctx, func(v *vault.Vault) (*vault.Vault, error) {
			err := v.Lock(passphrase)
			if err != nil {
				return nil, err
			}
			return v, nil
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.InitWalletResponse{
		// Mnemonic: mnemonic,
	}, nil
}
