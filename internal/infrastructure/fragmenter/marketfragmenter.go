package fragmenter

import (
	"fmt"
	"sort"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
)

type marketFragmenter struct {
	ephWallet *trade.Wallet
}

func NewMarketFragmenter(ephWallet *trade.Wallet) ports.Fragmenter {
	return &marketFragmenter{ephWallet}
}

func (m *marketFragmenter) Address() string {
	return m.ephWallet.Address()
}

func (m *marketFragmenter) Keys() ([]byte, []byte) {
	return m.ephWallet.PrivateKey(), m.ephWallet.BlindingKey()
}

func (m *marketFragmenter) FragmentAmount(
	args ...interface{},
) (baseFragments, quoteFragments []uint64, err error) {
	assetValuePair, fragmentationMap, err := marketParseArgs(args)
	if err != nil {
		return nil, nil, err
	}

	baseSum := uint64(0)
	quoteSum := uint64(0)
	for numOfUtxo, percentage := range fragmentationMap {
		for ; numOfUtxo > 0; numOfUtxo-- {
			if assetValuePair.BaseValue() > 0 {
				baseAssetPart := percent(int(assetValuePair.BaseValue()), percentage)
				baseSum += baseAssetPart
				baseFragments = append(baseFragments, baseAssetPart)
			}

			if assetValuePair.QuoteValue() > 0 {
				quoteAssetPart := percent(int(assetValuePair.QuoteValue()), percentage)
				quoteSum += quoteAssetPart
				quoteFragments = append(quoteFragments, quoteAssetPart)
			}
		}
	}

	sort.Slice(baseFragments, func(i, j int) bool {
		return baseFragments[i] < baseFragments[j]
	})

	sort.Slice(quoteFragments, func(i, j int) bool {
		return quoteFragments[i] < quoteFragments[j]
	})

	// if there is rest, created when calculating percentage,
	// add it to last fragment
	if baseSum != assetValuePair.BaseValue() {
		baseRest := assetValuePair.BaseValue() - baseSum
		if baseRest > 0 {
			baseFragments[len(baseFragments)-1] =
				baseFragments[len(baseFragments)-1] + baseRest
		}
	}

	// if there is rest, created when calculating percentage,
	// add it to last fragment
	if quoteSum != assetValuePair.QuoteValue() {
		quoteRest := assetValuePair.QuoteValue() - quoteSum
		if quoteRest > 0 {
			quoteFragments[len(quoteFragments)-1] =
				quoteFragments[len(quoteFragments)-1] + quoteRest
		}
	}

	return
}

func (m *marketFragmenter) CraftTransaction(
	ins []explorer.Utxo, outs []ports.TxOut, feeAmount uint64, lbtc string,
) (string, error) {
	return buildFinalizedTx(m.ephWallet, ins, outs, feeAmount, lbtc)
}

func marketParseArgs(args []interface{}) (
	assetValuePair ports.AssetValuePair, fragmentationMap map[int]int, err error,
) {
	if len(args) != 2 {
		err = fmt.Errorf("invalid number of args, expected 2 got %d", len(args))
		return
	}
	avp, ok := args[0].(ports.AssetValuePair)
	if !ok {
		err = fmt.Errorf("arg 0 must be of type AssetValuePair")
		return
	}
	fm, ok := args[1].(map[int]int)
	if !ok {
		err = fmt.Errorf("args 1 must be of type map[int]int")
		return
	}

	assetValuePair, fragmentationMap = avp, fm
	return
}
