package mapper

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	v0domain "github.com/tdex-network/tdex-daemon/old-v0"
)

func (m *mapperService) FromV091DepositsToV1Deposits(
	deposits []*v0domain.Deposit,
) ([]domain.Deposit, error) {
	res := make([]domain.Deposit, 0, len(deposits))
	depositsPerTxId := make(map[string][]v0domain.Deposit)
	for _, v := range deposits {
		if _, ok := depositsPerTxId[v.TxID]; !ok {
			depositsPerTxId[v.TxID] = make([]v0domain.Deposit, 0)
		}
		depositsPerTxId[v.TxID] = append(depositsPerTxId[v.TxID], *v)
	}

	for k, v := range depositsPerTxId {
		amountPerAsset := make(map[string]uint64)
		for _, vv := range v {
			amountPerAsset[vv.Asset] += vv.Value
		}

		label, err := m.getLabel(v[0].AccountIndex)
		if err != nil {
			return nil, err
		}

		res = append(res, domain.Deposit{
			AccountName:       label,
			TxID:              k,
			TotAmountPerAsset: amountPerAsset,
			Timestamp:         int64(v[0].Timestamp),
		})
	}

	return res, nil
}
