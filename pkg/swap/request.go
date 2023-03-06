package swap

import (
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/thanhpk/randstr"
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
	FeeAmount          uint64
	FeeAsset           string
}

func (o RequestOpts) validate() error {
	if isPsetV0(o.Transaction) {
		return validateSwapTxV0(
			o.Transaction,
			o.InputBlindingKeys,
			o.OutputBlindingKeys,
		)
	}
	if isPsetV2(o.Transaction) {
		if len(o.UnblindedInputs) <= 0 {
			return fmt.Errorf("missing unblinded inputs")
		}
		return validateSwapRequestTx(o)
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
