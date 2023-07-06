package mapper

import (
	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1/v1-domain"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func (m *mapperService) FromV091TransactionsToV1Transactions(
	trades []*domain.Trade, deposits []*domain.Deposit,
	withdrawals []*domain.Withdrawal, accountsByLabel map[string]string,
) map[string]*v1domain.Transaction {
	txs := make(map[string]*v1domain.Transaction)
	for _, trade := range trades {
		if len(trade.TxId) <= 0 {
			continue
		}
		accountName := accountsByLabel[trade.MarketName]
		if t, ok := txs[trade.TxId]; ok {
			t.AddAccount(accountName)
			continue
		}

		txs[trade.TxId] = &v1domain.Transaction{
			TxID:  trade.TxId,
			TxHex: trade.TxHex,
			Accounts: map[string]struct{}{
				accountName: {},
			},
		}
	}
	for _, deposit := range deposits {
		accountName := accountsByLabel[deposit.AccountName]
		if t, ok := txs[deposit.TxID]; ok {
			t.AddAccount(accountName)
			continue
		}

		txs[deposit.TxID] = &v1domain.Transaction{
			TxID: deposit.TxID,
			Accounts: map[string]struct{}{
				accountName: {},
			},
		}
	}
	for _, withdrawal := range withdrawals {
		accountName := accountsByLabel[withdrawal.AccountName]
		if t, ok := txs[withdrawal.TxID]; ok {
			t.AddAccount(accountName)
			continue
		}

		txs[withdrawal.TxID] = &v1domain.Transaction{
			TxID: withdrawal.TxID,
			Accounts: map[string]struct{}{
				accountName: {},
			},
		}
	}
	return txs
}
