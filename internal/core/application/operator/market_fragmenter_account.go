package operator

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

var (
	minMarketFragmentAmount = uint64(100000)
	fragmentationMap        = fragmenterMap{
		1: 0.3,
		2: 0.15,
		3: 0.1,
		5: 0.02,
	}
)

func (s *service) DeriveMarketFragmenterAddresses(
	ctx context.Context, num int,
) ([]string, error) {
	if !s.accountExists(ctx, domain.MarketFragmenterAccount) {
		if _, err := s.wallet.Account().CreateAccount(
			ctx, domain.MarketFragmenterAccount,
		); err != nil {
			return nil, err
		}
	}
	return s.wallet.Account().DeriveAddresses(
		ctx, domain.MarketFragmenterAccount, num,
	)
}

func (s *service) ListMarketFragmenterExternalAddresses(
	ctx context.Context,
) ([]string, error) {
	return s.wallet.Account().ListAddresses(ctx, domain.MarketFragmenterAccount)
}

func (s *service) GetMarketFragmenterBalance(
	ctx context.Context,
) (map[string]ports.Balance, error) {
	return s.wallet.Account().GetBalance(ctx, domain.MarketFragmenterAccount)
}

func (s *service) MarketFragmenterSplitFunds(
	ctx context.Context, mkt ports.Market, millisatsPerByte uint64,
	chRes chan ports.FragmenterReply,
) {
	defer close(chRes)

	chRes <- fragmenterReply{"fetching market", nil}

	market, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, mkt.GetBaseAsset(), mkt.GetQuoteAsset(),
	)
	if err != nil {
		chRes <- fragmenterReply{"", err}
		return
	}

	marketBalanceInfo, _ := s.wallet.Account().GetBalance(
		ctx, market.Name,
	)
	isZeroMarketBalance := func() bool {
		if marketBalanceInfo != nil {
			if b := marketBalanceInfo[market.BaseAsset]; b != nil &&
				b.GetConfirmedBalance() > 0 {
				return false
			}
			if b := marketBalanceInfo[market.QuoteAsset]; b != nil &&
				b.GetConfirmedBalance() > 0 {
				return false
			}
		}
		return true
	}

	chRes <- fragmenterReply{"fetching market fragmenter account balance", nil}

	balanceInfo, err := s.wallet.Account().GetBalance(
		ctx, domain.MarketFragmenterAccount,
	)
	if err != nil {
		chRes <- fragmenterReply{"", err}
		return
	}

	if len(balanceInfo) <= 0 {
		chRes <- fragmenterReply{
			"", fmt.Errorf("market fragmenter account has 0 balance"),
		}
		return
	}

	bb, bOk := balanceInfo[market.BaseAsset]
	qb, qOk := balanceInfo[market.QuoteAsset]
	if (!bOk && !qOk) || (bb.GetConfirmedBalance() == 0 &&
		qb.GetConfirmedBalance() == 0) {
		chRes <- fragmenterReply{
			"", fmt.Errorf("market fragmenter account has 0 balance"),
		}
		return
	}
	baseBalance := bb.GetConfirmedBalance()
	quoteBalance := qb.GetConfirmedBalance()
	if baseBalance < minMarketFragmentAmount &&
		quoteBalance < minMarketFragmentAmount {
		chRes <- fragmenterReply{
			"", fmt.Errorf(
				"not enough balance (base %d, quote %d) to fragment "+
					"(required at least %d)", baseBalance, quoteBalance,
				minMarketFragmentAmount,
			)}
		return
	}

	// If market has balanced strategy and zero balance, fragmenter must have
	// both base and quote asset non-zero balance. In all other cases it's ok
	// to fragment only base or quote balance.
	if market.IsStrategyBalanced() && isZeroMarketBalance() {
		if baseBalance < minMarketFragmentAmount {
			chRes <- fragmenterReply{
				"", fmt.Errorf(
					"since market has 0 balance, market fragmenter account must have "+
						"enough base balance (%d) to fragment (required at least %d)",
					baseBalance, minMarketFragmentAmount,
				),
			}
			return
		}
		if quoteBalance < minMarketFragmentAmount {
			chRes <- fragmenterReply{
				"", fmt.Errorf(
					"since market has 0 balance, market fragmenter account must have "+
						"enough quote balance (%d) to fragment (required at least %d)",
					quoteBalance, minMarketFragmentAmount,
				),
			}
			return
		}
	}

	numFragments := fragmentationMap.numFragments()
	numOfAddresses := numFragments
	if quoteBalance > minMarketFragmentAmount &&
		baseBalance > minMarketFragmentAmount {
		numOfAddresses *= 2
	}
	addresses, err := s.wallet.Account().DeriveAddresses(
		ctx, market.Name, numOfAddresses,
	)
	if err != nil {
		chRes <- fragmenterReply{
			"", fmt.Errorf("failed to derive addresses for market account: %s", err),
		}
		return
	}

	outputs := make([]ports.TxOutput, 0, numOfAddresses)
	if baseBalance > minMarketFragmentAmount {
		chRes <- fragmenterReply{
			fmt.Sprintf(
				"fetched funds with total base balance of %d, to be split into %d "+
					"fragments", baseBalance, numFragments,
			), nil,
		}

		for n, percentage := range fragmentationMap {
			amount := percentageAmount(baseBalance, percentage)
			for i := 0; i < n; i++ {
				addr := addresses[len(outputs)+i]
				outputs = append(outputs, output{market.BaseAsset, amount, addr})
			}
		}
	}

	if quoteBalance > minMarketFragmentAmount {
		chRes <- fragmenterReply{
			fmt.Sprintf(
				"fetched funds with total quote balance of %d, to be split into %d "+
					"fragments", quoteBalance, numFragments,
			), nil,
		}

		offset := (len(outputs) - 1)
		for n, percentage := range fragmentationMap {
			amount := percentageAmount(quoteBalance, percentage)
			for i := 0; i < n; i++ {
				addr := addresses[offset+i]
				outputs = append(outputs, output{market.QuoteAsset, amount, addr})
			}
		}
	}

	chRes <- fragmenterReply{
		"crafting and broadcasting transaciton to send funds to market account",
		nil,
	}
	txid, err := s.wallet.SendToMany(domain.MarketFragmenterAccount, outputs, 100)
	if err != nil {
		chRes <- fragmenterReply{
			"", fmt.Errorf("failed to send funds to market account: %s", err),
		}
		return
	}

	chRes <- fragmenterReply{
		fmt.Sprintf("market account funding transaction: %s", txid), nil,
	}

	chRes <- fragmenterReply{"fragmentation succeeded", nil}
}

func (s *service) WithdrawMarketFragmenterFunds(
	ctx context.Context, outputs []ports.TxOutput, millisatsPerByte uint64,
) (string, error) {
	return s.wallet.SendToMany(
		domain.MarketFragmenterAccount, outputs, millisatsPerByte,
	)
}

type fragmenterMap map[int]float64

func (m fragmenterMap) numFragments() int {
	tot := 0
	for num := range m {
		tot += num
	}
	return tot
}

func percentageAmount(amount uint64, percentage float64) uint64 {
	return decimal.NewFromInt(int64(amount)).Mul(
		decimal.NewFromFloat(percentage),
	).BigInt().Uint64()
}
