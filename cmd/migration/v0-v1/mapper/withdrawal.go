package mapper

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	v0domain "github.com/tdex-network/tdex-daemon/old-v0"
	"github.com/vulpemventures/go-elements/network"
)

var lbtcByNetwork = map[string]string{
	network.Liquid.Name:  network.Liquid.AssetID,
	network.Testnet.Name: network.Testnet.AssetID,
	network.Regtest.Name: network.Regtest.AssetID,
}

func (m *mapperService) FromV0WithdrawalsToV1Withdrawals(
	withdrawals []*v0domain.Withdrawal, net string,
) ([]domain.Withdrawal, error) {
	res := make([]domain.Withdrawal, 0, len(withdrawals))
	for _, v := range withdrawals {
		withdrawal, err := m.fromV0WithdrawalToV1Withdrawal(v, net)
		if err != nil {
			return nil, err
		}
		if withdrawal != nil {
			res = append(res, *withdrawal)
		}
	}

	return res, nil
}

func (m *mapperService) fromV0WithdrawalToV1Withdrawal(
	withdrawal *v0domain.Withdrawal, net string,
) (*domain.Withdrawal, error) {
	// In v0, the daemon stores withdrawals only for fee and market accounts,
	// therefore we must take care of both specific cases.
	if withdrawal.AccountIndex == v0domain.FeeAccount {
		label, err := m.getLabel(withdrawal.AccountIndex)
		if err != nil {
			return nil, err
		}
		return &domain.Withdrawal{
			AccountName: label,
			TxID:        withdrawal.TxID,
			TotAmountPerAsset: map[string]uint64{
				lbtcByNetwork[net]: withdrawal.BaseAmount,
			},
			Timestamp: int64(withdrawal.Timestamp),
		}, nil
	}

	market, err := m.v0RepoManager.MarketRepository().GetMarketByAccount(
		withdrawal.AccountIndex,
	)
	if err != nil {
		return nil, err
	}
	// If the market is not stored it means it's been deleted at some point.
	// In this case we skip translating withdrawals.
	if market == nil {
		return nil, nil
	}

	return &domain.Withdrawal{
		AccountName: market.AccountName(),
		TxID:        withdrawal.TxID,
		TotAmountPerAsset: map[string]uint64{
			market.BaseAsset:  withdrawal.BaseAmount,
			market.QuoteAsset: withdrawal.QuoteAmount,
		},
		Timestamp: int64(withdrawal.Timestamp),
	}, nil
}
