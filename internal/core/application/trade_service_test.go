package application_test

import (
	"encoding/base64"
	"encoding/hex"
	"math"
	"testing"
	"time"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	pbswap "github.com/tdex-network/tdex-protobuf/generated/go/swap"
)

var (
	tradeExpiryDuration = 120 * time.Second
	tradePriceSlippage  = decimal.NewFromFloat(0.1)

	tradeFeeOutpoints = []application.TxOutpoint{
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
	}
	tradeMktOutpoints = []application.TxOutpoint{
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
		{Hash: randomHex(32), Index: 0},
	}
)

func TestMarketTrading(t *testing.T) {
	repoManager, explorerSvc, bcListener := newServices() //

	t.Run("without fixed fees", func(t *testing.T) {
		tradeSvc, err := newTradeService(
			repoManager,
			explorerSvc,
			bcListener,
			false,
		)
		require.NoError(t, err)

		markets, err := tradeSvc.GetTradableMarkets(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, markets)

		market := markets[0].Market
		balances, err := tradeSvc.GetMarketBalance(ctx, market)
		require.NoError(t, err)
		require.NotNil(t, balances)
		require.True(t, balances.Balance.BaseAmount > 0)
		require.True(t, balances.Balance.QuoteAmount > 0)

		t.Run("buy LBTC fixed LBTC", func(t *testing.T) {
			t.Parallel()
			marketOrder(t, tradeSvc, market, application.TradeBuy, 0.1, marketBaseAsset)
		})
		t.Run("buy LBTC fixed USDT", func(t *testing.T) {
			t.Parallel()
			marketOrder(t, tradeSvc, market, application.TradeBuy, 900.0, marketQuoteAsset)
		})
		t.Run("sell LBTC fixed LBTC", func(t *testing.T) {
			t.Parallel()
			marketOrder(t, tradeSvc, market, application.TradeSell, 0.1, marketBaseAsset)
		})
		t.Run("sell LBTC fixed USDT", func(t *testing.T) {
			t.Parallel()
			marketOrder(t, tradeSvc, market, application.TradeSell, 900.0, marketQuoteAsset)
		})
	})

	t.Run("with fixed fees", func(t *testing.T) {
		tradeSvc, err := newTradeService(
			repoManager,
			explorerSvc,
			bcListener,
			false,
		)

		require.NoError(t, err)

		markets, err := tradeSvc.GetTradableMarkets(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, markets)

		market := markets[0].Market
		balances, err := tradeSvc.GetMarketBalance(ctx, market)
		require.NoError(t, err)
		require.NotNil(t, balances)
		require.True(t, balances.Balance.BaseAmount > 0)
		require.True(t, balances.Balance.QuoteAmount > 0)

		t.Run("buy LBTC fixed LBTC", func(t *testing.T) {
			marketOrder(t, tradeSvc, market, application.TradeBuy, 0.1, marketBaseAsset)
		})
		t.Run("buy LBTC fixed USDT", func(t *testing.T) {
			marketOrder(t, tradeSvc, market, application.TradeBuy, 900.0, marketQuoteAsset)
		})
		t.Run("sell LBTC fixed LBTC", func(t *testing.T) {
			marketOrder(t, tradeSvc, market, application.TradeSell, 0.1, marketBaseAsset)
		})
		t.Run("sell LBTC fixed USDT", func(t *testing.T) {
			marketOrder(t, tradeSvc, market, application.TradeSell, 900.0, marketQuoteAsset)
		})
	})
}

func newTradeService(
	repoManager ports.RepoManager,
	explorerSvc explorer.Service,
	bcListener application.BlockchainListener,
	withFixedFee bool,
) (application.TradeService, error) {

	v, err := repoManager.VaultRepository().GetOrCreateVault(
		ctx, mnemonic, passphrase, regtest,
	)
	if err != nil {
		return nil, err
	}

	unspents := make([]domain.Unspent, 0)
	for _, outpoint := range tradeFeeOutpoints {
		info, _ := v.DeriveNextExternalAddressForAccount(domain.FeeAccount)
		script, _ := hex.DecodeString(info.Script)
		unspents = append(unspents, domain.Unspent{
			TxID:            outpoint.Hash,
			VOut:            uint32(outpoint.Index),
			Value:           uint64(1000),
			AssetHash:       marketBaseAsset,
			ValueCommitment: randomValueCommitment(),
			AssetCommitment: randomAssetCommitment(),
			ValueBlinder:    randomBytes(32),
			AssetBlinder:    randomBytes(32),
			ScriptPubKey:    script,
			Nonce:           randomBytes(32),
			RangeProof:      make([]byte, 1),
			SurjectionProof: make([]byte, 1),
			Address:         info.Address,
			Confirmed:       true,
		})
	}

	oLen := len(tradeMktOutpoints)
	for i, outpoint := range tradeMktOutpoints {
		info, _ := v.DeriveNextExternalAddressForAccount(domain.MarketAccountStart)
		script, _ := hex.DecodeString(info.Script)
		assetHash := marketBaseAsset
		value := uint64(5000000)
		if i > (oLen/2 - 1) {
			assetHash = marketQuoteAsset
			value = 50000000000
		}

		unspents = append(unspents, domain.Unspent{
			TxID:            outpoint.Hash,
			VOut:            uint32(outpoint.Index),
			Value:           value,
			AssetHash:       assetHash,
			ValueCommitment: randomValueCommitment(),
			AssetCommitment: randomAssetCommitment(),
			ValueBlinder:    randomBytes(32),
			AssetBlinder:    randomBytes(32),
			ScriptPubKey:    script,
			Nonce:           randomBytes(32),
			RangeProof:      make([]byte, 1),
			SurjectionProof: make([]byte, 1),
			Address:         info.Address,
			Confirmed:       true,
		})
		explorerSvc.(*mockExplorer).On("GetTransaction", outpoint.Hash).
			Return(randomTxs(info.Address)[0], nil)
	}

	mkt, err := domain.NewMarket(
		domain.MarketAccountStart, marketBaseAsset, marketQuoteAsset, marketFee,
	)
	if err != nil {
		return nil, err
	}
	if _, err := repoManager.MarketRepository().GetOrCreateMarket(
		ctx, mkt,
	); err != nil {
		return nil, err
	}

	if err := repoManager.MarketRepository().UpdateMarket(ctx, mkt.AccountIndex, func(m *domain.Market) (*domain.Market, error) {
		if withFixedFee {
			m.ChangeFixedFee(600, 5000)
		}
		m.MakeTradable()
		return m, nil
	}); err != nil {
		return nil, err
	}

	if err := repoManager.VaultRepository().UpdateVault(ctx, func(_ *domain.Vault) (*domain.Vault, error) {
		return v, nil
	}); err != nil {
		return nil, err
	}
	if _, err := repoManager.UnspentRepository().AddUnspents(ctx, unspents); err != nil {
		return nil, err
	}

	mockedTradeManager := newMockedTradeManager()
	mockedTradeManager.
		On("FillProposal", mock.AnythingOfType("application.FillProposalOpts")).
		Return(&application.FillProposalResult{
			PsetBase64:         randomBase64(),
			SelectedUnspents:   randomSelection(unspents, mockedTradeManager.counter),
			InputBlindingKeys:  nil,
			OutputBlindingKeys: nil,
		}, nil)

	application.TradeManager = mockedTradeManager

	explorerSvc.(*mockExplorer).
		On("GetTransactionStatus", mock.AnythingOfType("string")).
		Return(mockTxStatus{
			"confirmed":    true,
			"block_time":   randomIntInRange(100000, 1000000),
			"block_height": randomIntInRange(100000, 1000000),
			"block_hash":   randomHex(32),
		}, nil)
	explorerSvc.(*mockExplorer).
		On("GetTransactionHex", mock.AnythingOfType("string")).
		Return(randomHex(1000), nil)
	explorerSvc.(*mockExplorer).
		On("BroadcastTransaction", mock.AnythingOfType("string")).
		Return(randomHex(32), nil)

	return application.NewTradeService(
		repoManager,
		explorerSvc,
		bcListener,
		marketBaseAsset,
		tradeExpiryDuration,
		tradePriceSlippage,
		regtest,
		feeBalanceThreshold,
	), nil
}

func marketOrder(
	t *testing.T,
	tradeSvc application.TradeService,
	market application.Market,
	tradeType int,
	btcAmount float64,
	asset string,
) {
	amount := uint64(btcAmount * math.Pow10(8))
	preview, err := tradeSvc.GetMarketPrice(ctx, market, tradeType, amount, asset)
	require.NoError(t, err)
	require.NotNil(t, preview)

	wallet, err := trade.NewRandomWallet(regtest)
	require.NoError(t, err)
	require.NotNil(t, wallet)
	_, script := wallet.Script()

	assetToSend := asset
	amountToSend := amount
	assetToReceive := preview.Asset
	amountToReceive := preview.Amount
	if tradeType == application.TradeSell && asset == marketQuoteAsset {
		assetToSend, assetToReceive = assetToReceive, assetToSend
		amountToSend, amountToReceive = amountToReceive, amountToSend
	}
	if tradeType == application.TradeBuy && asset == marketBaseAsset {
		assetToSend, assetToReceive = assetToReceive, assetToSend
		amountToSend, amountToReceive = amountToReceive, amountToSend
	}

	unspents := []explorer.Utxo{
		esplora.NewWitnessUtxo(
			randomHex(32),
			uint32(randomIntInRange(0, 15)),
			amountToSend,
			assetToSend,
			randomValueCommitment(),
			randomAssetCommitment(),
			randomBytes(32),
			randomBytes(32),
			script,
			randomBytes(32),
			randomBytes(100),
			randomBytes(100),
			true,
		),
	}

	psetBase64, _ := trade.NewSwapTx(
		unspents,
		assetToSend,
		amountToSend,
		assetToReceive,
		amountToReceive,
		script,
	)

	blindingKeyMap := map[string][]byte{
		hex.EncodeToString(script): wallet.BlindingKey(),
	}

	swapRequest := &pbswap.SwapRequest{
		Id:                randomId(),
		AssetP:            assetToSend,
		AmountP:           amountToSend,
		AssetR:            assetToReceive,
		AmountR:           amountToReceive,
		Transaction:       psetBase64,
		InputBlindingKey:  blindingKeyMap,
		OutputBlindingKey: blindingKeyMap,
	}

	swapAccept, swapFail, expiryTimestamp, err := tradeSvc.TradePropose(ctx, market, tradeType, swapRequest)
	require.NoError(t, err)
	require.Nil(t, swapFail)
	require.NotNil(t, swapAccept)
	require.True(t, time.Now().Before(time.Unix(int64(expiryTimestamp), 0)))

	swapComplete := &pbswap.SwapComplete{
		Id:          randomId(),
		AcceptId:    swapAccept.GetId(),
		Transaction: swapAccept.GetTransaction(),
	}

	time.Sleep(200 * time.Millisecond)
	_, _, err = tradeSvc.TradeComplete(ctx, swapComplete, nil)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
}

func randomBase64() string {
	return base64.StdEncoding.EncodeToString(randomBytes(100))
}

func randomSelection(list []domain.Unspent, counter int) []explorer.Utxo {
	selectedUtxos := make([]explorer.Utxo, 0, 3)
	for i := 0; i < 3; i++ {
		selectedIndex := counter*3 + i
		selectedUtxos = append(selectedUtxos, list[selectedIndex].ToUtxo())
	}
	return selectedUtxos
}
