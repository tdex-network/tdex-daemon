package mapper

import (
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func (m *mapperService) FromV091DepositsToV1Deposits(
	deposits []*v091domain.Deposit,
) ([]*domain.Deposit, error) {
	res := make([]*domain.Deposit, 0, len(deposits))
	depositsPerTxId := make(map[string][]v091domain.Deposit)
	for _, v := range deposits {
		if _, ok := depositsPerTxId[v.TxID]; !ok {
			depositsPerTxId[v.TxID] = make([]v091domain.Deposit, 0)
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

		res = append(
			res,
			&domain.Deposit{
				AccountName:       label,
				TxID:              k,
				TotAmountPerAsset: amountPerAsset,
				Timestamp:         int64(v[0].Timestamp),
			},
		)
	}

	return res, nil
}
