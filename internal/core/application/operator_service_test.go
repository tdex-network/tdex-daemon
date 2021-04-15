package application_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

var (
	marketQuoteAsset = randomHex(32)
	feeOutpoints     = []application.TxOutpoint{
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
	}
	mktOutpoints = []application.TxOutpoint{
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
	}
)

func TestAccountManagement(t *testing.T) {
	operatorSvc, err := newOperatorService()
	require.NoError(t, err)

	feeAddressesAndKeys, err := operatorSvc.DepositFeeAccount(ctx, 2)
	require.NoError(t, err)

	mockedBlinderManager := &mockBlinderManager{}
	for _, f := range feeAddressesAndKeys {
		key, _ := hex.DecodeString(f.BlindingKey)
		mockedBlinderManager.
			On("UnblindOutput", mock.AnythingOfType("*transaction.TxOutput"), key).
			Return(application.UnblindedResult(&transactionutil.UnblindedResult{
				AssetHash:    regtest.AssetID,
				Value:        randomValue(),
				AssetBlinder: randomBytes(32),
				ValueBlinder: randomBytes(32),
			}), true)
	}

	time.Sleep(50 * time.Millisecond)

	mktAddressesAndKeys, err := operatorSvc.DepositMarket(ctx, "", "", 2)
	require.NoError(t, err)

	for i, m := range mktAddressesAndKeys {
		asset := marketBaseAsset
		if i == 0 {
			asset = marketQuoteAsset
		}
		key, _ := hex.DecodeString(m.BlindingKey)
		mockedBlinderManager.
			On("UnblindOutput", mock.Anything, key).
			Return(application.UnblindedResult(&transactionutil.UnblindedResult{
				AssetHash:    asset,
				Value:        randomValue(),
				AssetBlinder: randomBytes(32),
				ValueBlinder: randomBytes(32),
			}), true)
	}

	time.Sleep(50 * time.Millisecond)

	application.BlinderManager = mockedBlinderManager

	err = operatorSvc.ClaimFeeDeposit(ctx, feeOutpoints)
	require.NoError(t, err)

	feeBalance, err := operatorSvc.FeeAccountBalance(ctx)
	require.NoError(t, err)
	require.Greater(t, feeBalance, int64(0))

	mkt := application.Market{
		BaseAsset:  marketBaseAsset,
		QuoteAsset: marketQuoteAsset,
	}
	err = operatorSvc.ClaimMarketDeposit(ctx, mkt, mktOutpoints)
	require.NoError(t, err)

	markets, err := operatorSvc.ListMarket(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(markets), 1)
	require.False(t, markets[0].Tradable)

	err = operatorSvc.OpenMarket(ctx, marketBaseAsset, marketQuoteAsset)
	require.NoError(t, err)

	markets, err = operatorSvc.ListMarket(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(markets), 1)
	require.True(t, markets[0].Tradable)

	err = operatorSvc.CloseMarket(ctx, marketBaseAsset, marketQuoteAsset)
	require.NoError(t, err)

	markets, err = operatorSvc.ListMarket(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(markets), 1)
	require.False(t, markets[0].Tradable)

	err = operatorSvc.DropMarket(ctx, int(markets[0].AccountIndex))
	require.NoError(t, err)

	markets, err = operatorSvc.ListMarket(ctx)
	require.NoError(t, err)
	require.Len(t, markets, 0)
}

// newOperatorService returns a new service with brand new and unlocked wallet.
func newOperatorService() (application.OperatorService, error) {
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

	return application.NewOperatorService(
		repoManager,
		explorerSvc,
		bcListener,
		marketBaseAsset,
		marketFee,
		regtest,
		feeBalanceThreshold,
	), nil
}
