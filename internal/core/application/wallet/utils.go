package wallet

import "github.com/tdex-network/tdex-daemon/internal/core/ports"

func totOutputAmountPerAsset(outs []ports.TxOutput) map[string]uint64 {
	tot := make(map[string]uint64)
	for _, out := range outs {
		tot[out.GetAsset()] += out.GetAmount()
	}
	return tot
}
