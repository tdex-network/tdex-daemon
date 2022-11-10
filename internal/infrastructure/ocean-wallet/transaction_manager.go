package oceanwallet

import (
	"context"
	"fmt"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/ocean/v1"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/vulpemventures/go-elements/psetv2"
	"google.golang.org/grpc"
)

type txManager struct {
	client pb.TransactionServiceClient
}

func newTxManager(conn *grpc.ClientConn) *txManager {
	return &txManager{pb.NewTransactionServiceClient(conn)}
}

func (m *txManager) GetTransaction(
	ctx context.Context, txid string,
) (string, error) {
	res, err := m.client.GetTransaction(ctx, &pb.GetTransactionRequest{
		Txid: txid,
	})
	if err != nil {
		return "", err
	}
	return res.GetTxHex(), nil
}

func (m *txManager) EstimateFees(
	ctx context.Context, ins []ports.TxInput, outs []ports.TxOutput,
) (uint64, error) {
	res, err := m.client.EstimateFees(ctx, &pb.EstimateFeesRequest{
		Inputs:  inputList(ins).toProto(),
		Outputs: outputList(outs).toProto(),
	})
	if err != nil {
		return 0, err
	}
	return res.GetFeeAmount(), nil
}

func (m *txManager) SelectUtxos(
	ctx context.Context, accountName, asset string, amount uint64,
) ([]ports.Utxo, uint64, int64, error) {
	res, err := m.client.SelectUtxos(ctx, &pb.SelectUtxosRequest{
		AccountName:  accountName,
		TargetAsset:  asset,
		TargetAmount: amount,
	})
	if err != nil {
		return nil, 0, -1, err
	}
	change := res.GetChange()
	expiryTime := res.GetExpirationDate()
	return utxoList(res.GetUtxos()).toPortableList(), change, expiryTime, nil
}

func (m *txManager) CreatePset(
	ctx context.Context, ins []ports.TxInput, outs []ports.TxOutput,
) (string, error) {
	res, err := m.client.CreatePset(ctx, &pb.CreatePsetRequest{
		Inputs:  inputList(ins).toProto(),
		Outputs: outputList(outs).toProto(),
	})
	if err != nil {
		return "", err
	}
	return res.GetPset(), nil
}

func (m *txManager) UpdatePset(
	ctx context.Context, pset string,
	ins []ports.TxInput, outs []ports.TxOutput,
) (string, error) {
	res, err := m.client.UpdatePset(ctx, &pb.UpdatePsetRequest{
		Pset:    pset,
		Inputs:  inputList(ins).toProto(),
		Outputs: outputList(outs).toProto(),
	})
	if err != nil {
		return "", err
	}
	return res.GetPset(), nil
}

func (m *txManager) BlindPset(
	ctx context.Context, pset string, extraUnblindedIns []ports.UnblindedInput,
) (string, error) {
	res, err := m.client.BlindPset(ctx, &pb.BlindPsetRequest{
		Pset:                 pset,
		LastBlinder:          true,
		ExtraUnblindedInputs: unblindedInputList(extraUnblindedIns).toProto(),
	})
	if err != nil {
		return "", err
	}
	return res.GetPset(), nil
}

func (m *txManager) SignPset(
	ctx context.Context, pset string, extractRawTx bool,
) (string, error) {
	res, err := m.client.SignPset(ctx, &pb.SignPsetRequest{
		Pset: pset,
	})
	if err != nil {
		return "", err
	}
	signedPset := res.GetPset()
	if !extractRawTx {
		return signedPset, nil
	}

	ptx, _ := psetv2.NewPsetFromBase64(signedPset)
	if err := psetv2.FinalizeAll(ptx); err != nil {
		return "", fmt.Errorf("failed to finalize signed pset: %s", err)
	}
	rawTx, err := psetv2.Extract(ptx)
	if err != nil {
		return "", fmt.Errorf(
			"failed to extract final tx from finalized pset: %s", err,
		)
	}
	return rawTx.ToHex()
}

func (m *txManager) Transfer(
	ctx context.Context,
	accountName string, outs []ports.TxOutput, millisatsPerByte uint64,
) (string, error) {
	res, err := m.client.Transfer(ctx, &pb.TransferRequest{
		AccountName:      accountName,
		Receivers:        outputList(outs).toProto(),
		MillisatsPerByte: millisatsPerByte,
	})
	if err != nil {
		return "", err
	}
	return res.GetTxHex(), nil
}

func (m *txManager) BroadcastTransaction(
	ctx context.Context, txHex string,
) (string, error) {
	res, err := m.client.BroadcastTransaction(
		ctx, &pb.BroadcastTransactionRequest{
			TxHex: txHex,
		},
	)
	if err != nil {
		return "", err
	}
	return res.GetTxid(), nil
}

type inputList []ports.TxInput

func (l inputList) toProto() []*pb.Input {
	list := make([]*pb.Input, 0, len(l))
	for _, in := range l {
		list = append(list, &pb.Input{
			Txid:          in.GetTxid(),
			Index:         in.GetIndex(),
			Script:        in.GetScript(),
			ScriptsigSize: uint64(in.GetScriptSigSize()),
			WitnessSize:   uint64(in.GetWitnessSize()),
		})
	}
	return list
}

type outputList []ports.TxOutput

func (l outputList) toProto() []*pb.Output {
	list := make([]*pb.Output, 0, len(l))
	for _, out := range l {
		list = append(list, &pb.Output{
			Asset:          out.GetAsset(),
			Amount:         out.GetAmount(),
			Script:         out.GetScript(),
			BlindingPubkey: out.GetBlindingKey(),
		})
	}
	return list
}

type unblindedInputList []ports.UnblindedInput

func (l unblindedInputList) toProto() []*pb.UnblindedInput {
	list := make([]*pb.UnblindedInput, 0, len(l))
	for _, in := range l {
		list = append(list, &pb.UnblindedInput{
			Index:         in.GetIndex(),
			Asset:         in.GetAsset(),
			Amount:        in.GetAmount(),
			AssetBlinder:  in.GetAssetBlinder(),
			AmountBlinder: in.GetAmountBlinder(),
		})
	}
	return list
}
