package swap

import (
	"errors"

	"github.com/novalagung/gubrak/v2"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/pset"
)

//Core defines the initial parameter
type Core struct {
	Verbose bool
}

func (*Core) compareMessagesAndTransaction(request *pb.SwapRequest, accept *pb.SwapAccept) error {
	decodedFromRequest, err := pset.NewPsetFromBase64(request.GetTransaction())
	if err != nil {
		return err
	}

	totalP, err := countCumulativeAmount(decodedFromRequest.Inputs, request.GetAssetP())
	if err != nil {
		return err
	}
	if totalP < request.GetAmountP() {
		return errors.New("Cumulative utxos count is not enough to cover SwapRequest.amount_p")
	}

	return nil
}

func countCumulativeAmount(utxos []pset.PInput, asset string) (uint64, error) {
	result, err := gubrak.From(utxos).
		Filter(func(each pset.PInput) bool {
			// TODO check if nonWitnessUtxo is given
			return assetHashFromBytes(each.WitnessUtxo.Asset) == asset
		}).
		Map(func(each pset.PInput) uint64 {
			var elementsValue [9]byte
			copy(elementsValue[:], each.WitnessUtxo.Value[0:9])
			value, _ := confidential.ElementsToSatoshiValue(elementsValue)
			return value
		}).
		Reduce(func(accumulator, value uint64) uint64 {
			return accumulator + value
		}, uint64(0)).ResultAndError()

	if err != nil {
		return 0, err
	}

	return result.(uint64), nil
}
