package operator

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

func (s *service) DeriveFeeAddresses(
	ctx context.Context, num int,
) ([]string, error) {
	if !s.accountExists(ctx, domain.FeeAccount) {
		if _, err := s.wallet.Account().CreateAccount(
			ctx, domain.FeeAccount, false,
		); err != nil {
			return nil, err
		}
	}
	return s.wallet.Account().DeriveAddresses(ctx, domain.FeeAccount, num)
}

func (s *service) ListFeeExternalAddresses(
	ctx context.Context,
) ([]string, error) {
	return s.wallet.Account().ListAddresses(ctx, domain.FeeAccount)
}

func (s *service) GetFeeBalance(ctx context.Context) (ports.Balance, error) {
	balance, err := s.wallet.Account().GetBalance(ctx, domain.FeeAccount)
	if err != nil {
		return nil, err
	}
	if len(balance) <= 0 {
		return nil, nil
	}

	return balance[s.wallet.NativeAsset()], nil
}

func (s *service) WithdrawFeeFunds(
	ctx context.Context, outs []ports.TxOutput, millisatsPerByte uint64,
) (string, error) {
	txHex, err := s.wallet.Transaction().Transfer(
		ctx, domain.FeeAccount, outs, millisatsPerByte,
	)
	if err != nil {
		return "", err
	}
	return s.wallet.Transaction().BroadcastTransaction(ctx, txHex)
}

func (s *service) accountExists(ctx context.Context, account string) bool {
	info, err := s.wallet.Wallet().Info(ctx)
	if err != nil {
		return false
	}
	for _, i := range info.GetAccounts() {
		if i.GetName() == account {
			return true
		}
	}
	return false
}
