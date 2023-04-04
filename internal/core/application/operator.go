package application

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/application/operator"
	"github.com/tdex-network/tdex-daemon/internal/core/application/pubsub"
	"github.com/tdex-network/tdex-daemon/internal/core/application/wallet"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type OperatorService interface {
	// Fee account
	DeriveFeeAddresses(ctx context.Context, num int) ([]string, error)
	ListFeeExternalAddresses(ctx context.Context) ([]string, error)
	GetFeeBalance(ctx context.Context) (ports.Balance, error)
	WithdrawFeeFunds(
		ctx context.Context,
		password string, outputs []ports.TxOutput, millisatsPerByte uint64,
	) (string, error)

	// Market account
	NewMarket(
		ctx context.Context,
		market ports.Market, marketName string,
		basePercentageFee, quotePercentageFee uint64,
		baseAssetPrecision, quoteAssetPrecision uint,
	) (ports.MarketInfo, error)
	GetMarketInfo(
		ctx context.Context, market ports.Market,
	) (ports.MarketInfo, error)
	DeriveMarketAddresses(
		ctx context.Context, market ports.Market, num int,
	) ([]string, error)
	ListMarketExternalAddresses(
		ctx context.Context, market ports.Market,
	) ([]string, error)
	GetMarketReport(
		ctx context.Context, market ports.Market,
		timeRange ports.TimeRange, groupByTimeFrame int,
	) (ports.MarketReport, error)
	OpenMarket(ctx context.Context, market ports.Market) error
	CloseMarket(ctx context.Context, market ports.Market) error
	DropMarket(ctx context.Context, market ports.Market) error
	WithdrawMarketFunds(
		ctx context.Context,
		password string, market ports.Market, outputs []ports.TxOutput, millisatsPerByte uint64,
	) (string, error)
	UpdateMarketPercentageFee(
		ctx context.Context,
		market ports.Market, baseFee, quoteFee int64,
	) (ports.MarketInfo, error)
	UpdateMarketFixedFee(
		ctx context.Context,
		market ports.Market, baseFixedFee, quoteFixedFee int64,
	) (ports.MarketInfo, error)
	UpdateMarketAssetsPrecision(
		ctx context.Context,
		market ports.Market, baseAssetPrecision, quoteAssetPrecision int,
	) error
	UpdateMarketPrice(
		ctx context.Context,
		market ports.Market, basePrice, quotePrice decimal.Decimal,
	) error
	UpdateMarketStrategy(
		ctx context.Context, market ports.Market, strategyType int,
	) error

	// Fee Fragmenter account
	DeriveFeeFragmenterAddresses(ctx context.Context, num int) ([]string, error)
	ListFeeFragmenterExternalAddresses(ctx context.Context) ([]string, error)
	GetFeeFragmenterBalance(
		ctx context.Context,
	) (map[string]ports.Balance, error)
	FeeFragmenterSplitFunds(
		ctx context.Context, maxFragments uint32, millisatsPerByte uint64,
		chRes chan ports.FragmenterReply,
	)
	WithdrawFeeFragmenterFunds(
		ctx context.Context,
		password string, outputs []ports.TxOutput, millisatsPerByte uint64,
	) (string, error)

	// Market fragmenter account
	DeriveMarketFragmenterAddresses(
		ctx context.Context, numOfAddresses int,
	) ([]string, error)
	ListMarketFragmenterExternalAddresses(ctx context.Context) ([]string, error)
	GetMarketFragmenterBalance(
		ctx context.Context,
	) (map[string]ports.Balance, error)
	MarketFragmenterSplitFunds(
		ctx context.Context, market ports.Market, millisatsPerByte uint64,
		chRes chan ports.FragmenterReply,
	)
	WithdrawMarketFragmenterFunds(
		ctx context.Context,
		password string, outputs []ports.TxOutput, millisatsPerByte uint64,
	) (string, error)

	// List methods
	ListMarkets(ctx context.Context) ([]ports.MarketInfo, error)
	ListTradesForMarket(
		ctx context.Context, market ports.Market, page ports.Page,
	) ([]ports.Trade, error)
	ListUtxos(
		ctx context.Context, accountName string, page ports.Page,
	) ([]ports.Utxo, []ports.Utxo, error)
	ListDeposits(
		ctx context.Context, accountName string, page ports.Page,
	) ([]ports.Deposit, error)
	ListWithdrawals(
		ctx context.Context, accountName string, page ports.Page,
	) ([]ports.Withdrawal, error)

	// Webhook
	AddWebhook(ctx context.Context, hook ports.Webhook) (string, error)
	RemoveWebhook(ctx context.Context, id string) error
	ListWebhooks(ctx context.Context, actionType int) ([]ports.WebhookInfo, error)
}

func NewOperatorService(
	walletSvc WalletService, pubsubSvc PubSubService,
	repoManager ports.RepoManager, feeAccountBalanceThreshold uint64,
) (OperatorService, error) {
	w := walletSvc.(*wallet.Service)
	p := pubsubSvc.(*pubsub.Service)
	return operator.NewService(
		w, p, repoManager, feeAccountBalanceThreshold,
	)
}
