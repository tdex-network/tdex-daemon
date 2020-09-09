package walletservice

import (
	"context"
	"encoding/hex"

	"github.com/tdex-network/tdex-daemon/internal/constant"
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"github.com/vulpemventures/go-elements/transaction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SendToMany creates, signs and eventually broadcasts a transaction sending the
// amounts of the assets to the receiving addresses listed in the request
func (s *Service) SendToMany(ctx context.Context, req *pb.SendToManyRequest) (*pb.SendToManyReply, error) {
	outputs, outputsBlindingKeys, err := parseRequestOutputs(req.GetOutputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	walletDerivedAddresses, err := s.vaultRepository.GetAllDerivedAddressesForAccount(ctx, constant.WalletAccount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	unspents, err := getUnspents(walletDerivedAddresses)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var txHex string
	err = s.vaultRepository.UpdateVault(ctx, func(v *vault.Vault) (*vault.Vault, error) {
		txHex, err = v.SendToMany(
			constant.WalletAccount,
			unspents,
			outputs,
			outputsBlindingKeys,
			int(req.GetSatPerKw()),
		)
		if err != nil {
			return nil, err
		}

		if req.GetPush() {
			explorer.BroadcastTransaction(txHex)
		}

		return v, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rawTx, _ := hex.DecodeString(txHex)
	return &pb.SendToManyReply{
		RawTx: rawTx,
	}, nil
}

func parseRequestOutputs(reqOutputs []*pb.TxOut) ([]*transaction.TxOutput, [][]byte, error) {
	outputs := make([]*transaction.TxOutput, 0, len(reqOutputs))
	blindingKeys := make([][]byte, 0, len(reqOutputs))

	for _, out := range reqOutputs {
		asset, err := bufferutil.AssetHashToBytes(out.GetAsset())
		if err != nil {
			return nil, nil, err
		}
		value, err := bufferutil.ValueToBytes(uint64(out.GetValue()))
		if err != nil {
			return nil, nil, err
		}
		script, blindingKey, err := parseConfidentialAddress(out.GetAddress())
		if err != nil {
			return nil, nil, err
		}

		output := transaction.NewTxOutput(asset, value, script)
		outputs = append(outputs, output)
		blindingKeys = append(blindingKeys, blindingKey)
	}
	return outputs, blindingKeys, nil
}
