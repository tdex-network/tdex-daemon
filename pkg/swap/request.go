package swap

import (
	"encoding/hex"
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/thanhpk/randstr"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/psetv2"
	"github.com/vulpemventures/go-elements/transaction"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type UnblindedInput struct {
	Index         uint32
	Asset         string
	Amount        uint64
	AssetBlinder  string
	AmountBlinder string
}

// RequestOpts is the struct to be given to the Request method
type RequestOpts struct {
	Id                 string
	AssetToSend        string
	AmountToSend       uint64
	AssetToReceive     string
	AmountToReceive    uint64
	Transaction        string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
	UnblindedInputs    []UnblindedInput
}

func (o RequestOpts) validate() error {
	if isPsetV0(o.Transaction) {
		return checkTxAndBlindKeys(
			o.Transaction,
			o.InputBlindingKeys,
			o.OutputBlindingKeys,
		)
	}
	if isPsetV2(o.Transaction) {
		return checkTxAndUnblindedIns(o.Transaction, o.UnblindedInputs)
	}

	return fmt.Errorf("invalid swap transaction format")
}

func (o RequestOpts) forV1() bool {
	return isPsetV0(o.Transaction)
}

func (o RequestOpts) forV2() bool {
	return isPsetV2(o.Transaction)
}

func (o RequestOpts) id() string {
	if o.Id != "" {
		return o.Id
	}
	return randstr.Hex(8)
}

func (o RequestOpts) unblindedIns() []*tdexv2.UnblindedInput {
	if len(o.UnblindedInputs) <= 0 {
		return nil
	}
	list := make([]*tdexv2.UnblindedInput, 0, len(o.UnblindedInputs))
	for _, in := range o.UnblindedInputs {
		list = append(list, &tdexv2.UnblindedInput{
			Index:         in.Index,
			Asset:         in.Asset,
			Amount:        in.Amount,
			AssetBlinder:  in.AssetBlinder,
			AmountBlinder: in.AmountBlinder,
		})
	}
	return list
}

// Request takes a RequestOpts struct and returns a serialized protobuf message.
func Request(opts RequestOpts) ([]byte, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	id := opts.id()
	var message protoreflect.ProtoMessage

	switch {
	case opts.forV1():
		message = &tdexv1.SwapRequest{
			Id: id,
			// Proposer
			AssetP:  opts.AssetToSend,
			AmountP: opts.AmountToSend,
			// Receiver
			AssetR:  opts.AssetToReceive,
			AmountR: opts.AmountToReceive,
			// PSETv0
			Transaction: opts.Transaction,
			// Blinding keys
			InputBlindingKey:  opts.InputBlindingKeys,
			OutputBlindingKey: opts.InputBlindingKeys,
		}
	case opts.forV2():
		fallthrough
	default:
		message = &tdexv2.SwapRequest{
			Id: id,
			// Proposer
			AssetP:  opts.AssetToSend,
			AmountP: opts.AmountToSend,
			// Receiver
			AssetR:  opts.AssetToReceive,
			AmountR: opts.AmountToReceive,
			// PSETv2
			Transaction: opts.Transaction,
			// Unblinded inputs
			UnblindedInputs: opts.unblindedIns(),
		}
	}

	return proto.Marshal(message)
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

func checkTxAndUnblindedIns(
	psetBase64 string, unblindedIns []UnblindedInput,
) error {
	ptx, _ := psetv2.NewPsetFromBase64(psetBase64)

	if len(unblindedIns) <= 0 {
		return fmt.Errorf("missing unblinded inputs")
	}
	for _, in := range unblindedIns {
		if uint64(in.Index) >= ptx.Global.InputCount {
			return fmt.Errorf("unblinded input index %d out of range", in.Index)
		}
	}

	return nil
}

func isPsetV0(tx string) bool {
	_, err := pset.NewPsetFromBase64(tx)
	return err == nil
}

func isPsetV2(tx string) bool {
	_, err := psetv2.NewPsetFromBase64(tx)
	return err == nil
}
