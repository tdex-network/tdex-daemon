package domain_test

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/domain/mocks"
	pkgswap "github.com/tdex-network/tdex-daemon/pkg/swap"
)

func TestTradePropose(t *testing.T) {
	mktName := "test"
	swapRequest := newSwapRequest()
	mktBaseAsset := swapRequest.GetAssetP()
	mktQuoteAsset := swapRequest.GetAssetR()
	mktPercentageFee := domain.MarketFee{BaseAsset: 25, QuoteAsset: 25}
	mktFixedFee := domain.MarketFee{}
	traderPubkey := []byte{}
	mockedSwapParser := mocks.NewMockSwapParser(t)
	mockedSwapParser.On("SerializeRequest", swapRequest).Return(randomBytes(100), -1)
	domain.SwapParserManager = mockedSwapParser
	tradeType := domain.TradeBuy

	tests := []struct {
		name  string
		trade *domain.Trade
	}{
		{
			name:  "with_trade_empty",
			trade: newTradeEmpty(),
		},
		{
			name:  "with_trade_proposal",
			trade: newTradeProposal(),
		},
		{
			name:  "with_trade_accepted",
			trade: newTradeAccepted(),
		},
		{
			name:  "with_trade_completed",
			trade: newTradeCompleted(),
		},
		{
			name:  "with_trade_settled",
			trade: newTradeSettled(),
		},
	}

	for i := range tests {
		tt := tests[i]

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ok, err := tt.trade.Propose(
				tradeType, swapRequest, mktName, mktBaseAsset, mktQuoteAsset,
				mktPercentageFee, mktFixedFee, traderPubkey,
			)
			require.NoError(t, err)
			require.True(t, ok)

			require.GreaterOrEqual(t, tt.trade.Status.Code, domain.TradeStatusCodeProposal)
			require.NotEmpty(t, tt.trade.SwapRequest.Id)
			require.NotNil(t, tt.trade.SwapRequest.Message)
			require.NotEmpty(t, tt.trade.SwapRequest.Timestamp)
		})
	}
}

func TestFailingTradePropose(t *testing.T) {
	mktName := "test"
	swapRequest := newSwapRequest()
	mktBaseAsset := swapRequest.GetAssetP()
	mktQuoteAsset := swapRequest.GetAssetR()
	mktPercentageFee := domain.MarketFee{BaseAsset: 25, QuoteAsset: 25}
	mktFixedFee := domain.MarketFee{}
	traderPubkey := []byte{}
	mockedSwapParser := mocks.NewMockSwapParser(t)
	mockedSwapParser.On("SerializeRequest", swapRequest).Return(nil, 1)
	mockedSwapParser.On(
		"SerializeFail", swapRequest.GetId(), mock.Anything,
	).Return(randomId(), randomBytes(100))
	domain.SwapParserManager = mockedSwapParser
	tradeType := domain.TradeBuy

	t.Run("invalid_request", func(t *testing.T) {
		trade := newTradeEmpty()

		ok, err := trade.Propose(
			tradeType, swapRequest, mktName, mktBaseAsset, mktQuoteAsset,
			mktPercentageFee, mktFixedFee, traderPubkey,
		)
		require.NoError(t, err)
		require.False(t, ok)
		require.True(t, trade.IsProposal())
		require.True(t, trade.IsRejected())
	})
}

func TestTradeAccept(t *testing.T) {
	tx := randomBase64(100)
	expiryDuration := time.Now().Add(1 * time.Minute).Unix()

	tests := []struct {
		name  string
		trade *domain.Trade
	}{
		{
			name:  "with_trade_proposal",
			trade: newTradeProposal(),
		},
		{
			name:  "with_trade_accepted",
			trade: newTradeAccepted(),
		},
		{
			name:  "with_trade_completed",
			trade: newTradeCompleted(),
		},
		{
			name:  "with_trade_settled",
			trade: newTradeSettled(),
		},
	}

	for i := range tests {
		tt := tests[i]

		mockedSwapParser := mocks.NewMockSwapParser(t)
		mockedSwapParser.On(
			"SerializeAccept", mock.Anything, mock.Anything, mock.Anything,
		).Return(randomId(), randomBytes(100), -1).Maybe()
		domain.SwapParserManager = mockedSwapParser

		t.Run(tt.name, func(t *testing.T) {
			ok, err := tt.trade.Accept(tx, nil, expiryDuration)
			require.NoError(t, err)
			require.True(t, ok)
			require.GreaterOrEqual(t, tt.trade.Status.Code, domain.TradeStatusCodeAccepted)
		})
	}
}

func TestFailingTradeAccept(t *testing.T) {
	tx := randomBase64(100)
	expiryDuration := time.Now().Add(1 * time.Minute).Unix()

	t.Run("invalid_request", func(t *testing.T) {
		trade := newTradeProposal()
		mockedSwapParser := mocks.NewMockSwapParser(t)
		mockedSwapParser.On(
			"SerializeAccept", mock.Anything, mock.Anything, mock.Anything,
		).Return("", nil, 2)
		mockedSwapParser.On(
			"SerializeFail", trade.SwapRequest.Id, mock.Anything,
		).Return(randomId(), randomBytes(100)).Maybe()
		domain.SwapParserManager = mockedSwapParser

		ok, err := trade.Accept(tx, nil, expiryDuration)
		require.NoError(t, err)
		require.False(t, ok)
		require.False(t, trade.IsAccepted())
		require.True(t, trade.IsRejected())
	})

	t.Run("invalid_status", func(t *testing.T) {
		trade := newTradeEmpty()

		ok, err := trade.Accept(tx, nil, expiryDuration)
		require.EqualError(t, err, domain.ErrTradeMustBeProposal.Error())
		require.False(t, ok)
		require.False(t, trade.IsAccepted())
	})
}

func TestTradeComplete(t *testing.T) {
	tests := []struct {
		name  string
		trade *domain.Trade
		tx    string
	}{
		{
			name:  "with_trade_accepted_psetBase64",
			trade: newTradeAccepted(),
			tx:    randomBase64(100),
		},
		{
			name:  "with_trade_accepted_txHex",
			trade: newTradeAccepted(),
			tx:    randomHex(100),
		},
		{
			name:  "with_trade_completed",
			trade: newTradeCompleted(),
			tx:    randomBase64(100),
		},
		{
			name:  "with_trade_settled",
			trade: newTradeSettled(),
			tx:    randomBase64(100),
		},
	}

	for i := range tests {
		tt := tests[i]

		mockedSwapParser := mocks.NewMockSwapParser(t)
		mockedSwapParser.On(
			"SerializeComplete", tt.trade.SwapAccept.Message, tt.tx,
		).Return(randomId(), randomBytes(100), -1).Maybe()
		mockedSwapParser.On(
			"ParseSwapTransaction", mock.Anything, mock.Anything,
		).Return(&domain.SwapTransactionDetails{
			PsetBase64: randomBase64(100),
			TxHex:      randomHex(100),
			Txid:       randomHex(32),
		}, -1).Maybe()
		domain.SwapParserManager = mockedSwapParser

		t.Run(tt.name, func(t *testing.T) {
			ok, err := tt.trade.Complete(tt.tx)
			require.NoError(t, err)
			require.True(t, ok)
			require.GreaterOrEqual(t, tt.trade.Status.Code, domain.TradeStatusCodeCompleted)
			require.NotEmpty(t, tt.trade.TxHex)
			require.NotEmpty(t, tt.trade.TxId)
		})
	}
}

func TestFailingTradeComplete(t *testing.T) {
	tx := randomBase64(100)

	t.Run("invalid_request", func(t *testing.T) {
		trade := newTradeAccepted()
		mockedSwapParser := mocks.NewMockSwapParser(t)
		mockedSwapParser.On(
			"ParseSwapTransaction", mock.Anything, mock.Anything,
		).Return(&domain.SwapTransactionDetails{
			PsetBase64: randomBase64(100),
			TxHex:      randomHex(100),
			Txid:       randomHex(32),
		}, -1)
		mockedSwapParser.On(
			"SerializeComplete",
			trade.SwapAccept.Message,
			tx,
		).Return("", nil, 3)
		mockedSwapParser.On(
			"SerializeFail", trade.SwapAccept.Id, mock.Anything,
		).Return(randomId(), randomBytes(100))
		domain.SwapParserManager = mockedSwapParser

		ok, err := trade.Complete(tx)
		require.NoError(t, err)
		require.False(t, ok)
		require.Empty(t, trade.TxId)
		require.Empty(t, trade.TxHex)
		require.False(t, trade.IsCompleted())
		require.True(t, trade.IsRejected())
	})

	t.Run("expired", func(t *testing.T) {
		trade := newTradeAccepted()
		trade.ExpiryTime = time.Now().AddDate(0, 0, -1).Unix()
		require.True(t, trade.IsExpired())

		res, err := trade.Complete(tx)
		require.EqualError(t, err, domain.ErrTradeExpired.Error())
		require.False(t, res)
		require.False(t, trade.IsCompleted())
		require.True(t, trade.IsRejected())
	})

	t.Run("invalid_status", func(t *testing.T) {
		tests := []struct {
			name  string
			trade *domain.Trade
		}{
			{
				name:  "with_trade_empty",
				trade: newTradeEmpty(),
			},
			{
				name:  "with_trade_proposal",
				trade: newTradeProposal(),
			},
		}

		for i := range tests {
			tt := tests[i]

			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				res, err := tt.trade.Complete(tx)
				require.EqualError(t, err, domain.ErrTradeMustBeAccepted.Error())
				require.False(t, res)
				require.False(t, tt.trade.IsCompleted())
			})
		}
	})

	t.Run("invalid_transaction", func(t *testing.T) {
		trade := newTradeAccepted()
		require.Empty(t, trade.TxId)
		require.Empty(t, trade.TxHex)

		mockedSwapParser := mocks.NewMockSwapParser(t)
		mockedSwapParser.On(
			"SerializeFail", trade.SwapAccept.Id, mock.Anything,
		).Return(randomId(), randomBytes(100))
		mockedSwapParser.On(
			"ParseSwapTransaction", mock.Anything,
		).Return(nil, 4)
		domain.SwapParserManager = mockedSwapParser

		ok, err := trade.Complete(tx)
		require.NoError(t, err)
		require.False(t, ok)
		require.Empty(t, trade.TxId)
		require.Empty(t, trade.TxHex)
		require.False(t, trade.IsCompleted())
		require.True(t, trade.IsRejected())
	})
}

func TestTradeSettle(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name  string
		trade *domain.Trade
	}{
		{
			name:  "with_trade_accepted",
			trade: newTradeAccepted(),
		},
		{
			name:  "with_trade_completed",
			trade: newTradeCompleted(),
		},
		{
			name:  "with_trade_settled",
			trade: newTradeSettled(),
		},
	}

	for i := range tests {
		tt := tests[i]

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ok, err := tt.trade.Settle(now)
			require.NoError(t, err)
			require.True(t, ok)
			require.True(t, tt.trade.IsSettled())
		})
	}
}

func TestFailingTradeSettle(t *testing.T) {
	now := time.Now().Unix()
	mockedSwapParser := mocks.NewMockSwapParser(t)
	mockedSwapParser.On(
		"SerializeFail", mock.Anything, mock.Anything,
	).Return(randomId(), randomBytes(100))
	domain.SwapParserManager = mockedSwapParser

	t.Run("failing_because_invalid_status", func(t *testing.T) {
		tests := []struct {
			name  string
			trade *domain.Trade
		}{
			{
				name:  "with_trade_empty",
				trade: newTradeEmpty(),
			},
			{
				name:  "with_trade_proposal",
				trade: newTradeProposal(),
			},
			{
				name:  "with_trade_failed",
				trade: newTradeFailed(),
			},
		}

		for i := range tests {
			tt := tests[i]

			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				ok, err := tt.trade.Settle(now)
				require.EqualError(t, err, domain.ErrTradeMustBeCompletedOrAccepted.Error())
				require.False(t, ok)
				require.False(t, tt.trade.IsSettled())
			})
		}

	})
}

func TestTradeExpire(t *testing.T) {
	oneDayAgo := time.Now().AddDate(0, 0, -1).Unix()

	tests := []struct {
		name  string
		trade *domain.Trade
	}{
		{
			name:  "with_trade_accepted",
			trade: newTradeAccepted(),
		},
		{
			name:  "with_trade_completed",
			trade: newTradeCompleted(),
		},
	}

	for i := range tests {
		tt := tests[i]
		tt.trade.ExpiryTime = oneDayAgo

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ok, err := tt.trade.Expire()
			require.NoError(t, err)
			require.True(t, ok)
			require.True(t, tt.trade.IsExpired())
		})
	}
}

func TestFailingTradeExpire(t *testing.T) {
	t.Run("failing_because_invalid_status", func(t *testing.T) {
		tests := []struct {
			name  string
			trade *domain.Trade
		}{
			{
				name:  "with_trade_empty",
				trade: newTradeEmpty(),
			},
			{
				name:  "with_trade_proposal",
				trade: newTradeProposal(),
			},
			{
				name:  "with_trade_settled",
				trade: newTradeSettled(),
			},
		}

		for i := range tests {
			tt := tests[i]

			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				ok, err := tt.trade.Expire()
				require.EqualError(t, err, domain.ErrTradeNullExpiryTime.Error())
				require.False(t, ok)
				require.False(t, tt.trade.IsExpired())
			})
		}

	})
}

func newTradeEmpty() *domain.Trade {
	return domain.NewTrade()
}

func newTradeProposal() *domain.Trade {
	req := newSwapRequest()
	now := time.Now()
	return &domain.Trade{
		Id:               randomId(),
		MarketQuoteAsset: req.GetAssetR(),
		MarketPrice: domain.MarketPrice{
			BasePrice:  decimal.NewFromInt(int64(req.GetAmountR())).Div(decimal.NewFromInt(int64(req.GetAmountP()))).String(),
			QuotePrice: decimal.NewFromInt(int64(req.GetAmountP())).Div(decimal.NewFromInt(int64(req.GetAmountR()))).String(),
		},
		Status: domain.TradeStatus{
			Code: domain.TradeStatusCodeProposal,
		},
		SwapRequest: &domain.Swap{
			Id:        req.GetId(),
			Message:   randomBytes(100),
			Timestamp: now.Unix(),
		},
	}
}

func newTradeAccepted() *domain.Trade {
	trade := newTradeProposal()
	acc := newSwapAccept()

	trade.Status = domain.TradeStatus{
		Code: domain.TradeStatusCodeAccepted,
	}
	trade.SwapAccept = &domain.Swap{
		Id:        acc.GetId(),
		Message:   randomBytes(100),
		Timestamp: time.Now().Unix(),
	}
	trade.ExpiryTime = time.Now().Add(5 * time.Minute).Unix()
	return trade
}

func newTradeCompleted() *domain.Trade {
	trade := newTradeAccepted()
	com := newSwapComplete()

	trade.Status = domain.TradeStatus{
		Code: domain.TradeStatusCodeCompleted,
	}
	trade.SwapComplete = &domain.Swap{
		Id:        com.GetId(),
		Message:   randomBytes(100),
		Timestamp: time.Now().Unix(),
	}
	trade.TxHex = randomHex(100)
	trade.TxId = randomHex(32)
	return trade
}

func newTradeSettled() *domain.Trade {
	trade := newTradeCompleted()
	trade.ExpiryTime = 0
	trade.SettlementTime = time.Now().Unix()
	trade.Status = domain.TradeStatus{
		Code: domain.TradeStatusCodeSettled,
	}
	return trade
}

func newTradeFailed() *domain.Trade {
	trade := newTradeProposal()
	trade.Fail(
		trade.SwapRequest.Id, pkgswap.ErrCodeRejectedSwapRequest,
	)
	return trade
}

func newSwapRequest() domain.SwapRequest {
	return domain.SwapRequest{
		Id:          randomId(),
		AssetP:      randomHex(32),
		AmountP:     10000,
		AssetR:      randomHex(32),
		AmountR:     2000000,
		Transaction: randomBase64(100),
	}
}

func newSwapAccept() domain.SwapAccept {
	return domain.SwapAccept{
		Id:          randomId(),
		RequestId:   randomId(),
		Transaction: randomBase64(100),
	}
}

func newSwapComplete() domain.SwapComplete {
	return domain.SwapComplete{
		Id:          randomId(),
		AcceptId:    randomId(),
		Transaction: randomBase64(100),
	}
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(len))
}

func randomBase64(len int) string {
	return base64.StdEncoding.EncodeToString(randomBytes(len))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	//nolint
	rand.Read(b)
	return b
}

func randomId() string {
	return uuid.New().String()
}
