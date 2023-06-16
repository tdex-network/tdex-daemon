package mapper

import (
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func (m *mapperService) FromV091WithdrawalsToV1Withdrawals(
	withdrawals []*v091domain.Withdrawal,
) ([]*domain.Withdrawal, error) {
	res := make([]*domain.Withdrawal, 0, len(withdrawals))
	for _, v := range withdrawals {
		withdrawal, err := m.fromV091WithdrawalToV1Withdrawal(v)
		if err != nil {
			return nil, err
		}
		res = append(res, withdrawal)
	}

	return res, nil
}

func (m *mapperService) fromV091WithdrawalToV1Withdrawal(
	withdrawal *v091domain.Withdrawal,
) (*domain.Withdrawal, error) {
	market, err := m.v091RepoManager.MarketRepository().GetMarketByAccount(
		withdrawal.AccountIndex,
	)
	if err != nil {
		return nil, err
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
