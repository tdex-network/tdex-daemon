package walletservice

import (
	"context"
	"encoding/hex"

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
func (s *Service) SendToMany(ctx context.Context, req *pb.SendToManyRequest) (reply *pb.SendToManyReply, err error) {
	outputs, outputsBlindingKeys, err := parseRequestOutputs(req.GetOutputs())
	if err != nil {
		err = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	walletDerivedAddresses, _, err := s.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, vault.WalletAccount)
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
		return
	}

	unspents, err := s.getUnspents(walletDerivedAddresses, nil)
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
		return
	}

	if err = s.vaultRepository.UpdateVault(ctx, nil, "", func(v *vault.Vault) (*vault.Vault, error) {
		mnemonic, err := v.Mnemonic()
		if err != nil {
			return nil, err
		}
		account, err := v.AccountByIndex(vault.WalletAccount)
		if err != nil {
			return nil, err
		}

		changePathsByAsset := map[string]string{}
		for _, asset := range getAssetsOfOutputs(outputs) {
			_, script, _, err := v.DeriveNextInternalAddressForAccount(vault.WalletAccount)
			if err != nil {
				return nil, err
			}
			derivationPath, _ := account.DerivationPathByScript(script)
			changePathsByAsset[asset] = derivationPath
		}

		txHex, err := sendToMany(
			mnemonic,
			account,
			unspents,
			outputs,
			outputsBlindingKeys,
			int(req.GetMillisatPerByte()),
			changePathsByAsset,
		)
		if err != nil {
			return nil, err
		}

		if req.GetPush() {
			if _, err := s.explorerService.BroadcastTransaction(txHex); err != nil {
				return nil, err
			}
		}

		rawTx, _ := hex.DecodeString(txHex)
		reply = &pb.SendToManyReply{
			RawTx: rawTx,
		}

		return v, nil
	}); err != nil {
		err = status.Error(codes.Internal, err.Error())
		return
	}

	return
}

func (s *Service) getUnspents(addresses []string, blindingKeys [][]byte) ([]explorer.Utxo, error) {
	chUnspents := make(chan []explorer.Utxo)
	chErr := make(chan error, 1)
	unspents := make([]explorer.Utxo, 0)

	for _, addr := range addresses {
		go s.getUnspentsForAddress(addr, blindingKeys, chUnspents, chErr)

		select {
		case err := <-chErr:
			close(chErr)
			close(chUnspents)
			return nil, err
		case unspentsForAddress := <-chUnspents:
			unspents = append(unspents, unspentsForAddress...)
		}
	}

	return unspents, nil
}

func (s *Service) getUnspentsForAddress(addr string, blindingKeys [][]byte, chUnspents chan []explorer.Utxo, chErr chan error) {
	unspents, err := s.explorerService.GetUnspents(addr, blindingKeys)
	if err != nil {
		chErr <- err
		return
	}
	chUnspents <- unspents
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
