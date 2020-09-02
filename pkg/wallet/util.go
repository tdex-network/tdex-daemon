package wallet

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/base58"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
	"github.com/vulpemventures/go-elements/slip77"
)

func generateMnemonicSeedAndMasterKey(entropySize int) (
	mnemonic string,
	seed []byte,
	err error,
) {
	entropy, err := bip39.NewEntropy(entropySize)
	if err != nil {
		return
	}
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return
	}
	seed = bip39.NewSeed(mnemonic, "")
	return
}

func generateSeedFromMnemonic(mnemonic string) []byte {
	return bip39.NewSeed(mnemonic, "")
}

func isMnemonicValid(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
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
