package fragmenter

import (
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
)

type feeFragmenter struct {
	ephWallet *trade.Wallet
}

func NewFeeFragmenter(ephWallet *trade.Wallet) ports.Fragmenter {
	return &feeFragmenter{ephWallet}
}

func (f *feeFragmenter) Address() string {
	return f.ephWallet.Address()
}

func (f *feeFragmenter) Keys() ([]byte, []byte) {
	return f.ephWallet.PrivateKey(), f.ephWallet.BlindingKey()
}

func (f *feeFragmenter) FragmentAmount(
	args ...interface{},
) (baseFragments, _ []uint64, err error) {
	valueToBeFragmented, fragmentValue, maxNumOfFragments, err := feeParseArgs(args)
	if err != nil {
		return nil, nil, err
	}

	res := make([]uint64, 0)
	for i := uint32(0); valueToBeFragmented >= fragmentValue && i < maxNumOfFragments; i++ {
		res = append(res, fragmentValue)
		valueToBeFragmented -= fragmentValue
	}
	if valueToBeFragmented > 0 {
		if len(res) > 0 {
			res[len(res)-1] += valueToBeFragmented
		} else {
			res = append(res, valueToBeFragmented)
		}
	}

	baseFragments = res
	return
}

func (f *feeFragmenter) CraftTransaction(
	ins []explorer.Utxo, outs []ports.TxOut, feeAmount uint64, lbtc string,
) (string, error) {
	return buildFinalizedTx(f.ephWallet, ins, outs, feeAmount, lbtc)
}

func feeParseArgs(args []interface{}) (
	valueToBeFragmented, fragmentValue uint64, maxNumOfFragments uint32, err error,
) {
	if len(args) != 3 {
		err = fmt.Errorf(
			"invalid number of args, expected 3 got %d", len(args),
		)
		return
	}
	vtf, ok := args[0].(uint64)
	if !ok {
		err = fmt.Errorf("arg 0 must be of type uint64")
		return
	}
	fv, ok := args[1].(uint64)
	if !ok {
		err = fmt.Errorf("arg 1 must be of type uint64")
		return
	}
	mnf, ok := args[2].(uint32)
	if !ok {
		err = fmt.Errorf("arg 2 must be of type uint64")
		return
	}

	valueToBeFragmented, fragmentValue, maxNumOfFragments = vtf, fv, mnf
	return
}
