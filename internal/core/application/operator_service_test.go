package application

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

var (
	marketQuoteAsset = randomHex(32)
	feeOutpoints     = []TxOutpoint{
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
	}
	mktOutpoints = []TxOutpoint{
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
	}
)

func TestAccountManagement(t *testing.T) {
	operatorSvc, err := newOperatorService()
	require.NoError(t, err)

	feeAddressesAndKeys, err := operatorSvc.GetFeeAddress(ctx, 2)
	require.NoError(t, err)

	mockedBlinderManager := &mockBlinderManager{}
	for _, f := range feeAddressesAndKeys {
		key, _ := hex.DecodeString(f.BlindingKey)
		mockedBlinderManager.
			On("UnblindOutput", mock.AnythingOfType("*transaction.TxOutput"), key).
			Return(UnblindedResult(&transactionutil.UnblindedResult{
				AssetHash:    regtest.AssetID,
				Value:        randomValue(),
				AssetBlinder: randomBytes(32),
				ValueBlinder: randomBytes(32),
			}), true)
	}

	time.Sleep(50 * time.Millisecond)

	mkt := Market{
		BaseAsset:  marketBaseAsset,
		QuoteAsset: marketQuoteAsset,
	}

	err = operatorSvc.NewMarket(ctx, mkt)
	require.NoError(t, err)

	mktAddressesAndKeys, err := operatorSvc.GetMarketAddress(ctx, mkt, 2)
	require.NoError(t, err)

	for i, m := range mktAddressesAndKeys {
		asset := marketBaseAsset
		if i == 0 {
			asset = marketQuoteAsset
		}
		key, _ := hex.DecodeString(m.BlindingKey)
		mockedBlinderManager.
			On("UnblindOutput", mock.Anything, key).
			Return(UnblindedResult(&transactionutil.UnblindedResult{
				AssetHash:    asset,
				Value:        randomValue(),
				AssetBlinder: randomBytes(32),
				ValueBlinder: randomBytes(32),
			}), true)
	}

	time.Sleep(50 * time.Millisecond)

	BlinderManager = mockedBlinderManager

	err = operatorSvc.ClaimFeeDeposits(ctx, feeOutpoints)
	require.NoError(t, err)

	_, feeBalance, err := operatorSvc.GetFeeBalance(ctx)
	require.NoError(t, err)
	require.Greater(t, feeBalance, int64(0))

	err = operatorSvc.ClaimMarketDeposits(ctx, mkt, mktOutpoints)
	require.NoError(t, err)

	markets, err := operatorSvc.ListMarkets(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(markets), 1)
	require.False(t, markets[0].Tradable)

	err = operatorSvc.OpenMarket(ctx, mkt)
	require.NoError(t, err)

	markets, err = operatorSvc.ListMarkets(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(markets), 1)
	require.True(t, markets[0].Tradable)

	err = operatorSvc.CloseMarket(ctx, mkt)
	require.NoError(t, err)

	markets, err = operatorSvc.ListMarkets(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(markets), 1)
	require.False(t, markets[0].Tradable)

	// TODO: uncomment the line belows after the following issue is fixed:
	// https://github.com/tdex-network/tdex-daemon/issues/482
	//
	// To drop the market it's required to withdraw all the funds first.
	// The builder/blinder/signer used by the WithdrawMarketFunds should be
	// detached to be mocked here, but it's currently empbedded and therefore the
	// market cannot be actually dropped.

	// err = operatorSvc.DropMarket(ctx, markets[0].Market)
	// require.NoError(t, err)

	// markets, err = operatorSvc.ListMarkets(ctx)
	// require.NoError(t, err)
	// require.Len(t, markets, 0)
}

// newOperatorService returns a new service with brand new and unlocked wallet.
func newOperatorService() (OperatorService, error) {
	repoManager, explorerSvc, bcListener := newServices()

	if _, err := repoManager.VaultRepository().GetOrCreateVault(
		ctx, mnemonic, passphrase, regtest,
	); err != nil {
		return nil, err
	}

	w, _ := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})

	accounts := []int{domain.FeeAccount, domain.MarketAccountStart}
	for _, accountIndex := range accounts {
		for i := 0; i < 2; i++ {
			addr, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
				DerivationPath: fmt.Sprintf("%d'/0/%d", accountIndex, i),
				Network:        regtest,
			})
			txid := feeOutpoints[i].Hash
			if accountIndex == domain.MarketAccountStart {
				txid = mktOutpoints[i].Hash
			}
			explorerSvc.(*mockExplorer).On("GetTransaction", txid).
				Return(randomTxs(addr)[0], nil)
		}
	}

	explorerSvc.(*mockExplorer).
		On("IsTransactionConfirmed", mock.AnythingOfType("string")).
		Return(true, nil)

	return NewOperatorService(
		repoManager,
		explorerSvc,
		bcListener,
		marketBaseAsset,
		"",
		marketFee,
		regtest,
		feeBalanceThreshold,
	), nil
}

func TestInitGroupedVolume(t *testing.T) {
	start := time.Now()
	type args struct {
		start            time.Time
		end              time.Time
		groupByTimeFrame int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "1",
			args: args{
				start:            start,
				end:              start.Add(12 * time.Hour),
				groupByTimeFrame: 4,
			},
			want: 3,
		},
		{
			name: "2",
			args: args{
				start:            start,
				end:              start.Add(12 * time.Hour),
				groupByTimeFrame: 1,
			},
			want: 12,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initGroupedVolume(tt.args.start, tt.args.end, tt.args.groupByTimeFrame); len(got) != tt.want {
				t.Errorf("initGroupedVolume() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestOperatorServiceGetMarketReport(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		repoManager ports.RepoManager
	}
	type args struct {
		ctx              context.Context
		market           Market
		timeRange        TimeRange
		groupByTimeFrame int
	}
	tests := []struct {
		name                  string
		fields                fields
		args                  args
		want                  func(report *MarketReport, notNilAt []int, wantGroupedVolumeLen int) error
		wantBaseQuoteValuesAt []int
		wantGroupedVolumeLen  int
		wantErr               bool
	}{
		{
			name: "1",
			fields: fields{
				repoManager: mockRepoManager(
					ctx,
					Market{
						BaseAsset:  "b",
						QuoteAsset: "q",
					},
					trades1(),
				),
			},
			args: args{
				ctx: ctx,
				market: Market{
					BaseAsset:  "b",
					QuoteAsset: "q",
				},
				timeRange: TimeRange{
					CustomPeriod: &CustomPeriod{
						StartDate: "2022-03-16T15:00:05Z",
						EndDate:   "2022-03-17T15:00:05Z",
					},
				},
				groupByTimeFrame: 1,
			},
			want: func(report *MarketReport, notNilAt []int, wantGroupedVolumeLen int) error {
				for _, v := range notNilAt {
					if report.GroupedVolume[v].BaseVolume == 0 && report.GroupedVolume[v].QuoteVolume == 0 {
						return errors.New(fmt.Sprintf("not expected to found volume with BaseVolume/QuoteVolume=0 at index: %v", v))
					}
				}

				if len(report.GroupedVolume) != wantGroupedVolumeLen {
					return errors.New(fmt.Sprintf("expected grouped volume len: %v, got: %v", wantGroupedVolumeLen, len(report.GroupedVolume)))
				}

				return nil
			},
			wantBaseQuoteValuesAt: []int{0},
			wantGroupedVolumeLen:  24,
			wantErr:               false,
		},
		{
			name: "2",
			fields: fields{
				repoManager: mockRepoManager(
					ctx,
					Market{
						BaseAsset:  "b",
						QuoteAsset: "q",
					},
					trades2(),
				),
			},
			args: args{
				ctx: ctx,
				market: Market{
					BaseAsset:  "b",
					QuoteAsset: "q",
				},
				timeRange: TimeRange{
					CustomPeriod: &CustomPeriod{
						StartDate: "2022-03-16T15:00:05Z",
						EndDate:   "2022-03-17T15:00:05Z",
					},
				},
				groupByTimeFrame: 4,
			},
			want: func(report *MarketReport, notNilAt []int, wantGroupedVolumeLen int) error {
				for _, v := range notNilAt {
					if report.GroupedVolume[v].BaseVolume == 0 && report.GroupedVolume[v].QuoteVolume == 0 {
						return errors.New(fmt.Sprintf("not expected to found volume with BaseVolume/QuoteVolume=0 at index: %v", v))
					}
				}

				if len(report.GroupedVolume) != wantGroupedVolumeLen {
					return errors.New(fmt.Sprintf("expected grouped volume len: %v, got: %v", wantGroupedVolumeLen, len(report.GroupedVolume)))
				}

				return nil
			},
			wantBaseQuoteValuesAt: []int{0, 5},
			wantGroupedVolumeLen:  6,
			wantErr:               false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &operatorService{
				repoManager: tt.fields.repoManager,
			}
			got, err := o.GetMarketReport(tt.args.ctx, tt.args.market, tt.args.timeRange, tt.args.groupByTimeFrame)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMarketReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := tt.want(got, tt.wantBaseQuoteValuesAt, tt.wantGroupedVolumeLen); err != nil {
				t.Error(err)
				return
			}
		})
	}
}

func mockRepoManager(ctx context.Context, market Market, trades []*domain.Trade) ports.RepoManager {
	repoManagerMock := new(ports.MockRepoManager)

	markerRepositoryMock := new(domain.MockMarketRepository)
	markerRepositoryMock.On(
		"GetMarketByAssets",
		ctx,
		market.BaseAsset,
		market.QuoteAsset,
	).Return(
		&domain.Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
		0,
		nil,
	)

	tradeRepositoryMock := new(domain.MockTradeRepository)
	tradeRepositoryMock.On("GetCompletedTradesByMarket", ctx, market.QuoteAsset).Return(trades, nil)

	repoManagerMock.On("MarketRepository").Return(markerRepositoryMock)
	repoManagerMock.On("TradeRepository").Return(tradeRepositoryMock)

	return repoManagerMock
}

func trades1() []*domain.Trade {
	trades := make([]*domain.Trade, 0)

	timePoint, _ := time.Parse(time.RFC3339, "2022-03-17T14:30:00Z")

	for i := 0; i < 5; i++ {
		trades = append(trades, &domain.Trade{
			MarketBaseAsset:     "b",
			MarketQuoteAsset:    "a",
			MarketFee:           20,
			MarketFixedBaseFee:  30,
			MarketFixedQuoteFee: 40,
			SwapRequest: domain.Swap{
				ID:        "",
				Timestamp: uint64(timePoint.Unix()),
			},
		})
	}

	return trades
}

func trades2() []*domain.Trade {
	trades := make([]*domain.Trade, 0)

	timePoint1, _ := time.Parse(time.RFC3339, "2022-03-17T14:30:00Z")

	for i := 0; i < 5; i++ {
		trades = append(trades, &domain.Trade{
			MarketBaseAsset:     "b",
			MarketQuoteAsset:    "a",
			MarketFee:           20,
			MarketFixedBaseFee:  30,
			MarketFixedQuoteFee: 40,
			SwapRequest: domain.Swap{
				ID:        "",
				Timestamp: uint64(timePoint1.Unix()),
			},
		})
	}

	timePoint2, _ := time.Parse(time.RFC3339, "2022-03-16T15:30:00Z")

	for i := 0; i < 5; i++ {
		trades = append(trades, &domain.Trade{
			MarketBaseAsset:     "b",
			MarketQuoteAsset:    "a",
			MarketFee:           20,
			MarketFixedBaseFee:  30,
			MarketFixedQuoteFee: 40,
			SwapRequest: domain.Swap{
				ID:        "",
				Timestamp: uint64(timePoint2.Unix()),
			},
		})
	}

	return trades
}
