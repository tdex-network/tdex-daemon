package swap

import (
	"encoding/hex"
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"
	"github.com/thanhpk/randstr"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
	"google.golang.org/protobuf/proto"
)

// RequestOpts is the struct to be given to the Request method
type RequestOpts struct {
	Id                 string
	AssetToSend        string
	AmountToSend       uint64
	AssetToReceive     string
	AmountToReceive    uint64
	PsetBase64         string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
}

func (o RequestOpts) validate() error {
	return checkTxAndBlindKeys(
		o.PsetBase64,
		o.InputBlindingKeys,
		o.OutputBlindingKeys,
	)
}

// Request takes a RequestOpts struct and returns a serialized protobuf message.
func Request(opts RequestOpts) ([]byte, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	id := opts.Id
	if len(id) <= 0 {
		id = randstr.Hex(8)
	}
	msg := &tdexv1.SwapRequest{
		Id: id,
		// Proposer
		AssetP:  opts.AssetToSend,
		AmountP: opts.AmountToSend,
		// Receiver
		AssetR:  opts.AssetToReceive,
		AmountR: opts.AmountToReceive,
		//PSET
		Transaction: opts.PsetBase64,
		// Blinding keys
		InputBlindingKey:  opts.InputBlindingKeys,
		OutputBlindingKey: opts.InputBlindingKeys,
	}

	return proto.Marshal(msg)
}

func checkTxAndBlindKeys(
	psetBase64 string,
	inBlindKeys, outBlindKeys map[string][]byte,
) error {
	ptx, err := pset.NewPsetFromBase64(psetBase64)
	if err != nil {
		return fmt.Errorf("pset is not in a valid base64 format")
	}

	checkInputKeys := inBlindKeys != nil
	for i, in := range ptx.Inputs {
		if !in.IsSane() {
			return fmt.Errorf("partial input %d is not sane", i)
		}
		var prevout *transaction.TxOutput
		if in.WitnessUtxo != nil {
			prevout = in.WitnessUtxo
		} else {
			txinIndex := ptx.UnsignedTx.Inputs[i].Index
			prevout = in.NonWitnessUtxo.Outputs[txinIndex]
		}
		if checkInputKeys {
			script := hex.EncodeToString(prevout.Script)
			if _, ok := inBlindKeys[script]; !ok {
				return fmt.Errorf("missing blinding key for input %d", i)
			}
		}
	}

	checkOutputKeys := outBlindKeys != nil
	for i, out := range ptx.UnsignedTx.Outputs {
		if len(out.Script) > 0 && checkOutputKeys {
			script := hex.EncodeToString(out.Script)
			if _, ok := outBlindKeys[script]; !ok {
				return fmt.Errorf("missing blinding key for output %d", i)
			}
		}
	}

	return nil
}
