package wallet

import (
	"bytes"
	"math"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil/base58"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-bip39"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/slip77"
	"github.com/vulpemventures/go-elements/transaction"
)

const (
	// MaxHardenedValue is the max value for hardened indexes of BIP32
	// derivation paths
	MaxHardenedValue = math.MaxUint32 - hdkeychain.HardenedKeyStart
)

func generateMnemonic(entropySize int) ([]string, error) {
	entropy, err := bip39.NewEntropy(entropySize)
	if err != nil {
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}
	return strings.Split(mnemonic, " "), nil
}

func generateSeedFromMnemonic(mnemonic []string) []byte {
	m := strings.Join(mnemonic, " ")
	return bip39.NewSeed(m, "")
}

func isMnemonicValid(mnemonic []string) bool {
	m := strings.Join(mnemonic, " ")
	return bip39.IsMnemonicValid(m)
}

func generateSigningMasterKey(
	seed []byte,
	path DerivationPath,
) ([]byte, error) {
	hdNode, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	for _, step := range path {
		hdNode, err = hdNode.Derive(step)
		if err != nil {
			return nil, err
		}
	}
	return base58.Decode(hdNode.String()), nil
}

func generateBlindingMasterKey(seed []byte) ([]byte, error) {
	slip77Node, err := slip77.FromSeed(seed)
	if err != nil {
		return nil, err
	}
	return slip77Node.MasterKey, nil
}

func anyOutputWithScript(outputs []*transaction.TxOutput, script []byte) bool {
	return outputIndexByScript(outputs, script) >= 0
}

func outputIndexByScript(outputs []*transaction.TxOutput, script []byte) int {
	for i, out := range outputs {
		if bytes.Equal(out.Script, script) {
			return i
		}
	}
	return -1
}

func getRemainingUnspents(unspents, unspentsToRemove []explorer.Utxo) []explorer.Utxo {
	isContained := func(unspent explorer.Utxo, unspents []explorer.Utxo) bool {
		for _, u := range unspents {
			if u.Hash() == unspent.Hash() && u.Index() == unspent.Index() {
				return true
			}
		}
		return false
	}
	remainingUnspents := make([]explorer.Utxo, 0)
	for _, unspent := range unspents {
		if !isContained(unspent, unspentsToRemove) {
			remainingUnspents = append(remainingUnspents, unspent)
		}
	}
	return remainingUnspents
}

func varIntSerializeSize(val uint64) int {
	// The value is small enough to be represented by itself, so it's
	// just 1 byte.
	if val < 0xfd {
		return 1
	}

	// Discriminant 1 byte plus 2 bytes for the uint16.
	if val <= math.MaxUint16 {
		return 3
	}

	// Discriminant 1 byte plus 4 bytes for the uint32.
	if val <= math.MaxUint32 {
		return 5
	}

	// Discriminant 1 byte plus 8 bytes for the uint64.
	return 9
}

func newTxOutput(asset string, value uint64, script []byte) (*transaction.TxOutput, error) {
	changeAsset, err := bufferutil.AssetHashToBytes(asset)
	if err != nil {
		return nil, err
	}
	changeValue, err := bufferutil.ValueToBytes(value)
	if err != nil {
		return nil, err
	}
	return transaction.NewTxOutput(changeAsset, changeValue, script), nil
}

func addInsAndOutsToPset(
	ptx *pset.Pset,
	inputsToAdd []explorer.Utxo,
	outputsToAdd []*transaction.TxOutput,
) (string, error) {
	updater, err := pset.NewUpdater(ptx)
	if err != nil {
		return "", err
	}

	for _, in := range inputsToAdd {
		input, witnessUtxo, _ := in.Parse()
		updater.AddInput(input)
		err := updater.AddInWitnessUtxo(witnessUtxo, len(ptx.Inputs)-1)
		if err != nil {
			return "", err
		}
	}

	for _, out := range outputsToAdd {
		updater.AddOutput(out)
	}

	return ptx.ToBase64()
}

func extractScriptTypesFromPset(ptx *pset.Pset) ([]int, []int, []int, []int, []int) {
	inScriptTypes := make([]int, 0, len(ptx.Inputs))
	inAuxiliaryRedeemScriptSize := make([]int, 0)
	inAuxiliaryWitnessSize := make([]int, 0)
	for i, in := range ptx.Inputs {
		var prevout *transaction.TxOutput
		if in.WitnessUtxo != nil {
			prevout = in.WitnessUtxo
		} else {
			prevoutIndex := ptx.UnsignedTx.Inputs[i].Index
			prevout = in.NonWitnessUtxo.Outputs[prevoutIndex]
		}

		sType := address.GetScriptType(prevout.Script)
		switch sType {
		case address.P2ShScript:
			if in.WitnessScript != nil {
				inScriptTypes = append(inScriptTypes, P2SH_P2WSH)
				// redeem script is treated as a multisig one. In case it's something
				// different, it is treated as a singlesig instead.
				m, _, _ := txscript.CalcMultiSigStats(in.RedeemScript)
				if m <= 0 {
					m = 1
				}
				scriptLen := len(in.RedeemScript)
				scriptSize := 1 + (1+72)*m + 1 + varIntSerializeSize(uint64(scriptLen)) + scriptLen
				inAuxiliaryWitnessSize = append(inAuxiliaryWitnessSize, scriptSize)
			} else if in.RedeemScript != nil {
				inScriptTypes[i] = P2SH_P2WPKH
			}
		case address.P2WpkhScript:
			inScriptTypes = append(inScriptTypes, P2WPKH)
		case address.P2WshScript:
			inScriptTypes = append(inScriptTypes, P2WSH)
			scriptSize := calcWitnessSizeFromRedeemScript(in.RedeemScript)
			inAuxiliaryWitnessSize = append(inAuxiliaryWitnessSize, scriptSize)
		case address.P2PkhScript:
			inScriptTypes = append(inScriptTypes, P2PKH)
		case address.P2MultiSigScript:
			inScriptTypes = append(inScriptTypes, P2MS)
			scriptSize := calcWitnessSizeFromRedeemScript(in.RedeemScript)
			inAuxiliaryRedeemScriptSize = append(inAuxiliaryRedeemScriptSize, scriptSize)
		}
	}

	outScriptTypes := make([]int, 0, len(ptx.Outputs))
	outAuxiliaryRedeemScriptSize := make([]int, 0)
	for _, out := range ptx.UnsignedTx.Outputs {
		sType := address.GetScriptType(out.Script)
		switch sType {
		case address.P2PkhScript:
			outScriptTypes = append(outScriptTypes, P2PKH)
		case address.P2ShScript:
			if len(out.Script) == 20 {
				outScriptTypes = append(outScriptTypes, P2SH_P2WPKH)
			} else {
				outScriptTypes = append(outScriptTypes, P2SH_P2WSH)
			}
		case address.P2MultiSigScript:
			outScriptTypes = append(outScriptTypes, P2MS)
			scriptLen := len(out.Script)
			scriptSize := varIntSerializeSize(uint64(scriptLen)) + scriptLen
			outAuxiliaryRedeemScriptSize = append(outAuxiliaryRedeemScriptSize, scriptSize)
		case address.P2WpkhScript:
			outScriptTypes = append(outScriptTypes, P2WPKH)
		case address.P2WshScript:
			outScriptTypes = append(outScriptTypes, P2WSH)
		}
	}

	return inScriptTypes, inAuxiliaryRedeemScriptSize, inAuxiliaryWitnessSize,
		outScriptTypes, outAuxiliaryRedeemScriptSize
}

func calcWitnessSizeFromRedeemScript(script []byte) int {
	// redeem script is treated as a multisig one. In case it's something
	// different, it is treated as a singlesig instead.
	m, _, _ := txscript.CalcMultiSigStats(script)
	if m <= 0 {
		m = 1
	}
	scriptLen := len(script)
	scriptSize := 1 + (1+72)*m + 1 + varIntSerializeSize(uint64(scriptLen)) + scriptLen
	return scriptSize
}

func calcFeeAmount(ptx *pset.Pset, nInputs, nOutputs, mSatsPerByte int) uint64 {
	inScriptTypes, inAuxiliaryRedeemScriptSize, inAuxiliaryWitnessSize,
		outScriptTypes, outAuxiliaryRedeemScriptSize := extractScriptTypesFromPset(ptx)
	// expect to add 1 input more to pay for network fees
	for i := 0; i < nInputs; i++ {
		inScriptTypes = append(inScriptTypes, P2WPKH)
	}
	for i := 0; i < nOutputs; i++ {
		outScriptTypes = append(outScriptTypes, P2WPKH)
	}

	txSize := EstimateTxSize(
		inScriptTypes, inAuxiliaryRedeemScriptSize, inAuxiliaryWitnessSize,
		outScriptTypes, outAuxiliaryRedeemScriptSize,
	)

	millisatsPerByte := float64(mSatsPerByte) / 1000
	return uint64(float64(txSize) * millisatsPerByte)
}
