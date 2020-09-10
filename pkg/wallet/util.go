package wallet

import (
	"bytes"
	"math"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/base58"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tyler-smith/go-bip39"
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
		hdNode, err = hdNode.Child(step)
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

func estimateTxSize(numInputs, numOutputs int, withChange bool, satsPerBytes int) uint64 {
	baseSize := calcTxSize(
		false, withChange,
		numInputs, numOutputs,
	)
	totalSize := calcTxSize(
		true, withChange,
		numInputs, numOutputs,
	)
	weight := baseSize*3 + totalSize
	vsize := (weight + 3) / 4
	return uint64(vsize * satsPerBytes)
}

func calcTxSize(withWitness, withChange bool, numInputs, numOutputs int) int {
	inputsSize := calcInputsSize(withWitness, numInputs)
	outputsSize := calcOutputsSize(withWitness, withChange, numOutputs)

	return 9 +
		varIntSerializeSize(uint64(numInputs)) +
		varIntSerializeSize(uint64(numOutputs)) +
		inputsSize +
		outputsSize
}

func calcInputsSize(withWitness bool, numInputs int) int {
	// prevout hash & index for each input
	size := (32 + 8) * numInputs
	if withWitness {
		// scriptsig + pubkey per each witness of all inputs
		size += numInputs * (72 + 33)
	}
	return size
}

func calcOutputsSize(withWitness, withChange bool, numOutputs int) int {
	// assetcommitment, valuecommitment, nonce
	baseOutputSize := 33 + 33 + 33
	size := baseOutputSize * numOutputs
	if withWitness {
		// rangeproof & surjection proof
		size += (4174 + 67) * numOutputs
	}

	if withChange {
		size += baseOutputSize
		if withWitness {
			size += 4174 + 67
		}
	}

	// fee asset, fee nonce, fee amount
	size += 33 + 1 + 9
	return size
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
