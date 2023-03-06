package swap

import (
	"encoding/hex"
	"fmt"

	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/psetv2"
	"github.com/vulpemventures/go-elements/transaction"
)

func isPsetV0(tx string) bool {
	_, err := pset.NewPsetFromBase64(tx)
	return err == nil
}

func isPsetV2(tx string) bool {
	_, err := psetv2.NewPsetFromBase64(tx)
	return err == nil
}

func validateSwapTxV0(
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

func validateSwapRequestTx(opts RequestOpts) error {
	ptx, _ := psetv2.NewPsetFromBase64(opts.Transaction)

	var amountP, amountR uint64
	for _, in := range opts.UnblindedInputs {
		if uint64(in.Index) >= ptx.Global.InputCount {
			return fmt.Errorf("unblinded input index %d out of range", in.Index)
		}
		// Sum cumulative amountP.
		if in.Asset == opts.AssetToSend {
			amountP += in.Amount
		}
	}

	for _, out := range ptx.Outputs {
		// Sum cumulative amountR.
		asset := elementsutil.TxIDFromBytes(out.Asset)
		if asset == opts.AssetToReceive {
			amountR += out.Value
		}
		// Subtract any change amount of assetP from amountP.
		if asset == opts.AssetToSend {
			amountP -= out.Value
		}
	}

	// Take out fees from calculated amount. Since it's not possible to know
	// whether the fees have been added or subtracted to the relative amount,
	// we try both ways and return an error only if both branches give negative
	// results.
	amounts := map[string]uint64{
		opts.AssetToSend:    amountP,
		opts.AssetToReceive: amountR,
	}
	amounts[opts.FeeAsset] += opts.FeeAmount
	if amounts[opts.AssetToSend] == opts.AmountToSend &&
		amounts[opts.AssetToReceive] == opts.AmountToReceive {
		return nil
	}

	amounts = map[string]uint64{
		opts.AssetToSend:    amountP,
		opts.AssetToReceive: amountR,
	}
	amounts[opts.FeeAsset] -= opts.FeeAmount
	if amounts[opts.AssetToSend] != opts.AmountToSend ||
		amounts[opts.AssetToReceive] != opts.AmountToReceive {
		return fmt.Errorf(
			"transaction in/out amounts do not match amount_p/amount_r",
		)
	}

	return nil
}

func validateSwapAcceptTx(tx string, ins []UnblindedInput) error {
	ptx, _ := psetv2.NewPsetFromBase64(tx)

	for _, in := range ins {
		if uint64(in.Index) >= ptx.Global.InputCount {
			return fmt.Errorf("unblinded input index %d out of range", in.Index)
		}
	}

	return nil
}
