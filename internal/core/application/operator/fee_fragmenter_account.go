package operator

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/vulpemventures/go-elements/address"
)

const (
	maxNumFragments     = 20
	defaultNumFragments = maxNumFragments
	feeFragmentAmount   = uint64(5000)
)

func (s *service) DeriveFeeFragmenterAddresses(
	ctx context.Context, num int,
) ([]string, error) {
	if !s.accountExists(ctx, domain.FeeFragmenterAccount) {
		if _, err := s.wallet.Account().CreateAccount(
			ctx, domain.FeeFragmenterAccount, false,
		); err != nil {
			return nil, err
		}
	}
	return s.wallet.Account().DeriveAddresses(
		ctx, domain.FeeFragmenterAccount, num,
	)
}

func (s *service) ListFeeFragmenterExternalAddresses(
	ctx context.Context,
) ([]string, error) {
	return s.wallet.Account().ListAddresses(ctx, domain.FeeFragmenterAccount)
}

func (s *service) GetFeeFragmenterBalance(
	ctx context.Context,
) (map[string]ports.Balance, error) {
	return s.wallet.Account().GetBalance(ctx, domain.FeeFragmenterAccount)
}

func (s *service) FeeFragmenterSplitFunds(
	ctx context.Context, maxFragments uint32, millisatsPerByte uint64,
	chRes chan ports.FragmenterReply,
) {
	defer close(chRes)

	chRes <- fragmenterReply{"fetching fee fragmenter account LBTC balance", nil}

	balanceInfo, err := s.wallet.Account().GetBalance(ctx, domain.FeeFragmenterAccount)
	if err != nil {
		chRes <- fragmenterReply{"", err}
		return
	}
	if len(balanceInfo) <= 0 {
		chRes <- fragmenterReply{
			"", fmt.Errorf("fee fragmenter account has 0 LBTC balance"),
		}
		return
	}
	lbtc := s.wallet.NativeAsset()
	lbtcBalance, ok := balanceInfo[lbtc]
	if !ok || lbtcBalance.GetConfirmedBalance() == 0 {
		chRes <- fragmenterReply{
			"", fmt.Errorf("fee fragmenter account has 0 LBTC balance"),
		}
		return
	}

	balance := lbtcBalance.GetConfirmedBalance()
	if balance < feeFragmentAmount {
		chRes <- fragmenterReply{
			"", fmt.Errorf(
				"not enough LBTC balance  (%d) to fragment (required ar least %d)",
				balance, feeFragmentAmount,
			),
		}
		return
	}
	numFragments := defaultNumFragments
	if maxFragments > 0 {
		numFragments = int(maxFragments)
	}
	if feeFragmentAmount*uint64(numFragments)-1 >= balance {
		chRes <- fragmenterReply{
			"", fmt.Errorf(
				"not enough lbtc balance (%d) for requested number of fragments "+
					"(%d - %d sats each)",
				lbtcBalance.GetConfirmedBalance(), numFragments, feeFragmentAmount,
			)}
		return
	}

	chRes <- fragmenterReply{
		fmt.Sprintf(
			"fetched funds with total LBTC balance of %d, to be split into %d "+
				"fragments of %d sats each", balance, numFragments, feeFragmentAmount,
		), nil,
	}

	// make sure the fee account is created.
	if _, err := s.wallet.Account().CreateAccount(
		ctx, domain.FeeAccount, false,
	); err != nil {
		if !strings.Contains(err.Error(), "already exist") {
			chRes <- fragmenterReply{
				"", fmt.Errorf("failed to create fee fragmenter account: %s", err),
			}
			return
		}
	}

	chRes <- fragmenterReply{
		"crafting transaction to deposit funds to fee account", nil,
	}

	addresses, err := s.wallet.Account().DeriveAddresses(
		ctx, domain.FeeAccount, numFragments,
	)
	if err != nil {
		chRes <- fragmenterReply{
			"", fmt.Errorf("failed to derive addresses from fee account: %s", err),
		}
		return
	}

	outputs := make([]ports.TxOutput, 0, numFragments)
	for i := 0; i < numFragments; i++ {
		amount := feeFragmentAmount
		if i == numFragments-1 {
			amount = balance
		}
		outputs = append(outputs, output{lbtc, amount, addresses[i]})
		balance -= feeFragmentAmount
	}

	txHex, err := s.wallet.Transaction().Transfer(
		ctx, domain.FeeFragmenterAccount, outputs, 100,
	)
	if err != nil {
		chRes <- fragmenterReply{"", err}
		return
	}

	chRes <- fragmenterReply{"broadcasting transaction", nil}

	txid, err := s.wallet.Transaction().BroadcastTransaction(ctx, txHex)
	if err != nil {
		chRes <- fragmenterReply{"", err}
		return
	}

	chRes <- fragmenterReply{
		fmt.Sprintf("fee account funding transaction: %s", txid), nil,
	}

	chRes <- fragmenterReply{"fragmentation succeeded", nil}
}

func (s *service) WithdrawFeeFragmenterFunds(
	ctx context.Context, outs []ports.TxOutput, millisatsPerByte uint64,
) (string, error) {
	return s.wallet.Transaction().Transfer(
		ctx, domain.FeeFragmenterAccount, outs, millisatsPerByte,
	)
}

type output struct {
	asset   string
	amount  uint64
	address string
}

func (o output) GetAsset() string {
	return o.asset
}

func (o output) GetAmount() uint64 {
	return o.amount
}

func (o output) GetScript() string {
	script, _ := address.ToOutputScript(o.address)
	return hex.EncodeToString(script)
}

func (o output) GetBlindingKey() string {
	info, _ := address.FromConfidential(o.address)
	return hex.EncodeToString(info.BlindingKey)
}

type fragmenterReply struct {
	msg string
	err error
}

func (r fragmenterReply) GetMessage() string {
	return r.msg
}

func (r fragmenterReply) GetError() error {
	return r.err
}
