package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/pset"
)

const (
	// MaxBlindingAttempts is the max number of times the blinding of a pset
	// can be repeated in case it fails to generate valid proofs.
	MaxBlindingAttempts = 8
	// DefaultBlindingAttempts is the default number of times the blinding of a
	// pset is retried if it fails to generate valid proofs.
	DefaultBlindingAttempts = 4
)

// BlindTransactionWithKeysOpts is the struct given to BlindTransactionWithKeys method
type BlindTransactionWithKeysOpts struct {
	PsetBase64         string
	OutputBlindingKeys [][]byte
	Attempts           int
}

func (o BlindTransactionWithKeysOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}
	ptx, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}
	for _, in := range ptx.Inputs {
		if in.WitnessUtxo == nil {
			return ErrNullInputWitnessUtxo
		}
	}

	if len(o.OutputBlindingKeys) != len(ptx.Outputs) {
		return ErrInvalidOutputBlindingKeysLen
	}

	if o.Attempts < 0 || o.Attempts > MaxBlindingAttempts {
		return ErrInvalidAttempts
	}
	return nil
}

func (o BlindTransactionWithKeysOpts) maxAttempts() int {
	if o.Attempts == 0 {
		return DefaultBlindingAttempts
	}
	return o.Attempts
}

// BlindTransactionWithKeys blinds the outputs of the provided partial transaction
// by deriving the blinding keys from the output scripts following SLIP-77 spec
func (w *Wallet) BlindTransactionWithKeys(opts BlindTransactionWithKeysOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	inKeysLen := len(ptx.Inputs)
	inBlindingKeys := make([]pset.BlindingDataLike, inKeysLen, inKeysLen)
	for i, in := range ptx.Inputs {
		blindingPrvkey, _, _ := w.DeriveBlindingKeyPair(DeriveBlindingKeyPairOpts{
			Script: in.WitnessUtxo.Script,
		})
		inBlindingKeys[i] = pset.PrivateBlindingKey(blindingPrvkey.Serialize())
	}

	outBlindingKeys := make(map[int][]byte)
	for i, k := range opts.OutputBlindingKeys {
		outBlindingKeys[i] = k
	}

	if err := w.blindTransaction(
		ptx,
		inBlindingKeys,
		outBlindingKeys,
		opts.maxAttempts(),
	); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

// BlindSwapTransactionWithKeysOpts is the struct given to BlindSwapTransactionWithKeys method
type BlindSwapTransactionWithKeysOpts struct {
	PsetBase64         string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
	Attempts           int
}

func (o BlindSwapTransactionWithKeysOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}
	ptx, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}

	for i, in := range ptx.Inputs {
		script := hex.EncodeToString(in.WitnessUtxo.Script)
		if _, ok := o.InputBlindingKeys[script]; !ok {
			return fmt.Errorf(
				"missing blinding key for input %d with script '%s'", i, script,
			)
		}
	}
	for i, out := range ptx.UnsignedTx.Outputs {
		script := hex.EncodeToString(out.Script)
		if _, ok := o.OutputBlindingKeys[script]; !ok {
			return fmt.Errorf(
				"missing blinding key for output %d with script '%s'", i, script,
			)
		}
	}

	if o.Attempts < 0 || o.Attempts > MaxBlindingAttempts {
		return ErrInvalidAttempts
	}

	return nil
}

func (o BlindSwapTransactionWithKeysOpts) maxAttempts() int {
	if o.Attempts == 0 {
		return DefaultBlindingAttempts
	}
	return o.Attempts
}

// BlindSwapTransactionWithKeys blinds the outputs of a swap transaction. Since this
// type of transaciton is composed of inputs and outputs owned by 2 different
// parties, the blinding keys for inputs and outputs are provided through maps
// outputScript -> blinding key. Note that all the blinding keys provided must
// be private, thus for the outputs this function will use the provided
// blinding keys to get the list of all public keys. This of course also means
// that no blinding keys are derived internally, but these are all provided as
// function arguments.
func (w *Wallet) BlindSwapTransactionWithKeys(opts BlindSwapTransactionWithKeysOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	inKeysLen := len(opts.InputBlindingKeys)
	inBlindingKeys := make([]pset.BlindingDataLike, inKeysLen, inKeysLen)
	for i, in := range ptx.Inputs {
		script := hex.EncodeToString(in.WitnessUtxo.Script)
		inBlindingKeys[i] = pset.PrivateBlindingKey(opts.InputBlindingKeys[script])
	}

	outBlindingKeys := make(map[int][]byte)
	for i, out := range ptx.UnsignedTx.Outputs {
		script := hex.EncodeToString(out.Script)
		_, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), opts.OutputBlindingKeys[script])
		outBlindingKeys[i] = pubkey.SerializeCompressed()
	}

	if err := w.blindTransaction(
		ptx,
		inBlindingKeys,
		outBlindingKeys,
		opts.maxAttempts(),
	); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

type BlindingData struct {
	Asset         string
	Amount        uint64
	AssetBlinder  []byte
	AmountBlinder []byte
}

func (b BlindingData) validate() error {
	asset, err := hex.DecodeString(b.Asset)
	if err != nil || len(asset) != 32 {
		return ErrInvalidInputAsset
	}
	if len(b.AssetBlinder) != 32 {
		return ErrInvalidInputAssetBlinder
	}
	if len(b.AmountBlinder) != 32 {
		return ErrInvalidInputAmountBlinder
	}
	return nil
}

func (b BlindingData) ToBlindingData() pset.BlindingData {
	asset, _ := hex.DecodeString(b.Asset)
	return pset.BlindingData{
		Asset:               elementsutil.ReverseBytes(asset),
		Value:               b.Amount,
		AssetBlindingFactor: b.AssetBlinder,
		ValueBlindingFactor: b.AmountBlinder,
	}
}

// BlindTransactionWithDataOpts is the struct given to BlindTransactionWithData method
type BlindTransactionWithDataOpts struct {
	PsetBase64         string
	InputBlindingData  map[int]BlindingData
	OutputBlindingKeys [][]byte
	Attempts           int
}

func (o BlindTransactionWithDataOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}
	ptx, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}
	for _, in := range ptx.Inputs {
		if in.WitnessUtxo == nil {
			return ErrNullInputWitnessUtxo
		}
	}

	if o.InputBlindingData == nil {
		return ErrNullInputBlindingData
	}
	for i, b := range o.InputBlindingData {
		if i < 0 || i >= len(ptx.Inputs) {
			return ErrInvalidInputIndex
		}
		if err := b.validate(); err != nil {
			return err
		}
	}

	if len(o.OutputBlindingKeys) != len(ptx.Outputs) {
		return ErrInvalidOutputBlindingKeysLen
	}

	if o.Attempts < 0 || o.Attempts > MaxBlindingAttempts {
		return ErrInvalidAttempts
	}
	return nil
}

func (o BlindTransactionWithDataOpts) maxAttempts() int {
	if o.Attempts == 0 {
		return DefaultBlindingAttempts
	}
	return o.Attempts
}

// BlindTransactionWithData blinds the outputs of the provided partial transaction
// by using the provided input blinding data.
func (w *Wallet) BlindTransactionWithData(opts BlindTransactionWithDataOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	dataLen := len(opts.InputBlindingData)
	inBlindingData := make([]pset.BlindingDataLike, dataLen, dataLen)
	for i, b := range opts.InputBlindingData {
		inBlindingData[i] = b.ToBlindingData()
	}

	outBlindingKeys := make(map[int][]byte)
	for i, k := range opts.OutputBlindingKeys {
		outBlindingKeys[i] = k
	}

	if err := w.blindTransaction(
		ptx,
		inBlindingData,
		outBlindingKeys,
		opts.maxAttempts(),
	); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

// BlindSwapTransactionWithDataOpts is the struct given to BlindSwapTransactionWithKeys method
type BlindSwapTransactionWithDataOpts struct {
	PsetBase64         string
	InputBlindingData  map[string]BlindingData
	OutputBlindingKeys map[string][]byte
	Attempts           int
}

func (o BlindSwapTransactionWithDataOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}
	ptx, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}

	for i, in := range ptx.Inputs {
		script := hex.EncodeToString(in.WitnessUtxo.Script)
		if _, ok := o.InputBlindingData[script]; !ok {
			return fmt.Errorf(
				"missing blinding data for input %d with script '%s'", i, script,
			)
		}
	}
	for i, out := range ptx.UnsignedTx.Outputs {
		script := hex.EncodeToString(out.Script)
		if _, ok := o.OutputBlindingKeys[script]; !ok {
			return fmt.Errorf(
				"missing blinding data for output %d with script '%s'", i, script,
			)
		}
	}

	if o.Attempts < 0 || o.Attempts > MaxBlindingAttempts {
		return ErrInvalidAttempts
	}

	return nil
}

func (o BlindSwapTransactionWithDataOpts) maxAttempts() int {
	if o.Attempts == 0 {
		return DefaultBlindingAttempts
	}
	return o.Attempts
}

// BlindSwapTransactionWithData blinds the outputs of a swap transaction.
// Instead of unblinding the input proofs with keys, blinding data
// (asset, value and respective blinders) are provided as a map
// script -> blinding_data.
func (w *Wallet) BlindSwapTransactionWithData(opts BlindSwapTransactionWithDataOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	dataLen := len(opts.InputBlindingData)
	inBlindingData := make([]pset.BlindingDataLike, dataLen, dataLen)
	for i, in := range ptx.Inputs {
		script := hex.EncodeToString(in.WitnessUtxo.Script)
		inBlindingData[i] = opts.InputBlindingData[script].ToBlindingData()
	}

	outBlindingKeys := make(map[int][]byte)
	for i, out := range ptx.UnsignedTx.Outputs {
		script := hex.EncodeToString(out.Script)
		_, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), opts.OutputBlindingKeys[script])
		outBlindingKeys[i] = pubkey.SerializeCompressed()
	}

	if err := w.blindTransaction(
		ptx,
		inBlindingData,
		outBlindingKeys,
		opts.maxAttempts(),
	); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

func (w *Wallet) blindTransaction(
	ptx *pset.Pset,
	inBlindingData []pset.BlindingDataLike,
	outBlindingKeys map[int][]byte,
	maxAttempts int,
) error {
	blinder, err := pset.NewBlinder(
		ptx,
		inBlindingData,
		outBlindingKeys,
		nil,
		nil,
	)
	if err != nil {
		return err
	}

	retryCount := 0
	for {
		if retryCount >= maxAttempts {
			return ErrReachedMaxBlindingAttempts
		}

		if err := blinder.Blind(); err != nil {
			if err == pset.ErrGenerateSurjectionProof {
				retryCount++
				continue
			}
			return err
		}
		return nil
	}
}

// ExtractBlindingDataFromTx unblinds the confidential inputs of the given tx
// (in pset's base64 format) with the provided blinding keys.
// The revealed data are returned mapped by output script.
func ExtractBlindingDataFromTx(
	psetBase64 string,
	inBlindingKeys map[string][]byte,
) (map[string]BlindingData, error) {
	ptx, err := pset.NewPsetFromBase64(psetBase64)
	if err != nil {
		return nil, err
	}

	blindingData := make(map[string]BlindingData)
	for i, in := range ptx.Inputs {
		prevout := in.WitnessUtxo
		if in.WitnessUtxo == nil {
			prevoutIndex := ptx.UnsignedTx.Inputs[i].Index
			prevout = in.NonWitnessUtxo.Outputs[prevoutIndex]
		}
		script := hex.EncodeToString(prevout.Script)
		var blindKey []byte

		if prevout.IsConfidential() {
			var ok bool
			blindKey, ok = inBlindingKeys[script]
			if !ok {
				return nil, ErrMissingInBlindingKey
			}
		}
		res, ok := transactionutil.UnblindOutput(prevout, blindKey)
		if !ok {
			return nil, ErrInvalidInBlindingKey
		}

		blindingData[script] = BlindingData{
			Asset:         res.AssetHash,
			Amount:        res.Value,
			AssetBlinder:  res.AssetBlinder,
			AmountBlinder: res.ValueBlinder,
		}
	}

	return blindingData, nil
}
