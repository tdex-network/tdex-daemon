package trade

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/txscript"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/psetv2"
)

func NewSwapTx(
	utxos []explorer.Utxo, inAsset, outAsset string,
	inAmount, outAmount uint64, outScript, outBlindingKey []byte,
) (string, error) {
	ptx, err := psetv2.New(nil, nil, nil)
	if err != nil {
		return "", err
	}

	selectedUtxos, change, err := explorer.SelectUnspents(
		utxos,
		inAmount,
		inAsset,
	)
	if err != nil {
		return "", err
	}

	updater, _ := psetv2.NewUpdater(ptx)

	for _, u := range selectedUtxos {
		_, prevout, _ := u.Parse()
		if err := updater.AddInputs([]psetv2.InputArgs{
			{
				Txid:    u.Hash(),
				TxIndex: u.Index(),
			},
		}); err != nil {
			return "", err
		}

		index := int(ptx.Global.InputCount) - 1
		if err := updater.AddInWitnessUtxo(index, prevout); err != nil {
			return "", err
		}
		if err := updater.AddInUtxoRangeProof(
			index, prevout.RangeProof,
		); err != nil {
			return "", err
		}
	}

	outputs := []psetv2.OutputArgs{
		{
			Asset:        outAsset,
			Amount:       outAmount,
			Script:       outScript,
			BlindingKey:  outBlindingKey,
			BlinderIndex: uint32(ptx.Global.InputCount),
		},
	}
	if change > 0 {
		outputs = append(outputs, psetv2.OutputArgs{
			Asset:       inAsset,
			Amount:      change,
			Script:      outScript,
			BlindingKey: outBlindingKey,
		})
	}

	if err := updater.AddOutputs(outputs); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

type Wallet struct {
	privateKey         *btcec.PrivateKey
	blindingPrivateKey *btcec.PrivateKey
	network            *network.Network
}

func NewRandomWallet(net *network.Network) (*Wallet, error) {
	prvkey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	blindPrvkey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}

	return &Wallet{prvkey, blindPrvkey, net}, nil
}

func NewWalletFromKey(privateKey, blindingKey []byte, net *network.Network) *Wallet {
	prvkey, _ := btcec.PrivKeyFromBytes(privateKey)
	blindPrvkey, _ := btcec.PrivKeyFromBytes(blindingKey)

	return &Wallet{prvkey, blindPrvkey, net}
}

func (w *Wallet) Address() string {
	p2wpkh := payment.FromPublicKey(w.privateKey.PubKey(), w.network, w.blindingPrivateKey.PubKey())
	ctAddress, _ := p2wpkh.ConfidentialWitnessPubKeyHash()
	return ctAddress
}

func (w *Wallet) Script() ([]byte, []byte) {
	p2wpkh := payment.FromPublicKey(w.privateKey.PubKey(), w.network, w.blindingPrivateKey.PubKey())
	return p2wpkh.Script, p2wpkh.WitnessScript
}

func (w *Wallet) Sign(psetBase64 string) (string, error) {
	ptx, err := psetv2.NewPsetFromBase64(psetBase64)
	if err != nil {
		return "", err
	}
	signer, err := psetv2.NewSigner(ptx)
	if err != nil {
		return "", err
	}

	for i, in := range ptx.Inputs {
		script, witnessScript := w.Script()
		if bytes.Equal(in.WitnessUtxo.Script, witnessScript) {
			tx, err := ptx.UnsignedTx()
			if err != nil {
				return "", err
			}
			hashForSignature := tx.HashForWitnessV0(
				i,
				script,
				in.WitnessUtxo.Value,
				txscript.SigHashAll,
			)

			sig := ecdsa.Sign(w.privateKey, hashForSignature[:])
			if !sig.Verify(hashForSignature[:], w.privateKey.PubKey()) {
				return "", fmt.Errorf(
					"signature verification failed for input %d",
					i,
				)
			}

			sigWithSigHashType := append(sig.Serialize(), byte(txscript.SigHashAll))
			if err := signer.SignInput(
				i, sigWithSigHashType, w.privateKey.PubKey().SerializeCompressed(),
				nil, nil,
			); err != nil {
				return "", err
			}
		}
	}

	return ptx.ToBase64()
}

func (w *Wallet) PrivateKey() []byte {
	return w.privateKey.Serialize()
}

func (w *Wallet) BlindingKey() []byte {
	return w.blindingPrivateKey.Serialize()
}
