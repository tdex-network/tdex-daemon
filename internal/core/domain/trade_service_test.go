package domain_test

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func TestTradePropose(t *testing.T) {
	domain.SwapParser = newMockedSwapParser(false)

	swapRequest := newMockedSwapRequest()
	marketAsset := swapRequest.GetAssetR()
	marketFee := int64(25)
	expiryDuration := uint64(600)
	traderPubkey := []byte{}

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

			ok, err := tt.trade.Propose(swapRequest, marketAsset, marketFee, expiryDuration, traderPubkey)
			require.NoError(t, err)
			require.True(t, ok)

			require.GreaterOrEqual(t, tt.trade.Status.Code, domain.Proposal)
			require.NotEmpty(t, tt.trade.SwapRequest.ID)
			require.NotNil(t, tt.trade.SwapRequest.Message)
			require.NotEmpty(t, tt.trade.SwapRequest.Timestamp)
		})
	}
}

func TestFailingTradePropose(t *testing.T) {
	swapRequest := newMockedSwapRequest()
	marketAsset := swapRequest.GetAssetR()
	marketFee := int64(25)
	expiryDuration := uint64(600)
	traderPubkey := []byte{}

	t.Run("failing_because_invalid_request", func(t *testing.T) {
		trade := newTradeEmpty()
		domain.SwapParser = newMockedSwapParser(true)

		ok, err := trade.Propose(swapRequest, marketAsset, marketFee, expiryDuration, traderPubkey)
		require.NoError(t, err)
		require.False(t, ok)
		require.True(t, trade.IsProposal())
		require.True(t, trade.IsRejected())
	})
}

func TestTradeAccept(t *testing.T) {
	domain.SwapParser = newMockedSwapParser(false)

	tx := randomBase64(100)
	inBlindKeys := map[string][]byte{
		randomHex(20): randomBytes(32),
	}
	outBlindKeys := map[string][]byte{
		randomHex(20): randomBytes(32),
		randomHex(20): randomBytes(32),
	}

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

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ok, err := tt.trade.Accept(tx, inBlindKeys, outBlindKeys)
			require.NoError(t, err)
			require.True(t, ok)
			require.GreaterOrEqual(t, tt.trade.Status.Code, domain.Accepted)
		})
	}
}

func TestFailingTradeAccept(t *testing.T) {
	tx := randomBase64(100)
	inBlindKeys := map[string][]byte{
		randomHex(20): randomBytes(32),
	}
	outBlindKeys := map[string][]byte{
		randomHex(20): randomBytes(32),
		randomHex(20): randomBytes(32),
	}

	t.Run("failing_because_invalid_request", func(t *testing.T) {
		trade := newTradeProposal()
		domain.SwapParser = newMockedSwapParser(true)

		ok, err := trade.Accept(tx, inBlindKeys, outBlindKeys)
		require.NoError(t, err)
		require.False(t, ok)
		require.False(t, trade.IsAccepted())
		require.True(t, trade.IsRejected())
	})

	t.Run("failing_because_expired", func(t *testing.T) {
		trade := newTradeProposal()
		trade.ExpiryTime = uint64(time.Now().AddDate(0, 0, -1).Unix())
		require.True(t, trade.IsExpired())

		domain.SwapParser = newMockedSwapParser(true)

		ok, err := trade.Accept(tx, inBlindKeys, outBlindKeys)
		require.EqualError(t, err, domain.ErrTradeExpired.Error())
		require.False(t, ok)
		require.False(t, trade.IsAccepted())
	})

	t.Run("failing_because_invalid_status", func(t *testing.T) {
		trade := newTradeEmpty()
		domain.SwapParser = newMockedSwapParser(false)

		ok, err := trade.Accept(tx, inBlindKeys, outBlindKeys)
		require.EqualError(t, err, domain.ErrTradeMustBeProposal.Error())
		require.False(t, ok)
		require.False(t, trade.IsAccepted())
	})
}

func TestTradeComplete(t *testing.T) {
	domain.SwapParser = newMockedSwapParser(false)
	domain.PsetManager = newMockedPsetManager()

	tx := randomBase64(100)
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

			res, err := tt.trade.Complete(tx)
			require.NoError(t, err)
			require.NotNil(t, res)
			require.True(t, res.OK)
			require.GreaterOrEqual(t, tt.trade.Status.Code, domain.Completed)
		})
	}
}

func TestFailingTradeComplete(t *testing.T) {
	tx := randomBase64(100)

	t.Run("failing_because_invalid_request", func(t *testing.T) {
		trade := newTradeAccepted()
		domain.SwapParser = newMockedSwapParser(true)

		res, err := trade.Complete(tx)
		require.NoError(t, err)
		require.False(t, res.OK)
		require.Empty(t, res.TxID)
		require.Empty(t, res.TxHex)
		require.False(t, trade.IsCompleted())
		require.True(t, trade.IsRejected())
	})

	t.Run("failing_because_expired", func(t *testing.T) {
		trade := newTradeAccepted()
		trade.ExpiryTime = uint64(time.Now().AddDate(0, 0, -1).Unix())
		require.True(t, trade.IsExpired())

		domain.SwapParser = newMockedSwapParser(true)

		res, err := trade.Complete(tx)
		require.EqualError(t, err, domain.ErrTradeExpired.Error())
		require.Nil(t, res)
		require.False(t, trade.IsCompleted())
		require.True(t, trade.IsRejected())
	})

	t.Run("failing_because_invalid_status", func(t *testing.T) {
		domain.SwapParser = newMockedSwapParser(false)

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
				require.Nil(t, res)
				require.False(t, tt.trade.IsCompleted())
			})
		}

	})
}

func TestTradeSettle(t *testing.T) {
	now := uint64(time.Now().Unix())

	tests := []struct {
		name  string
		trade *domain.Trade
	}{
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
	now := uint64(time.Now().Unix())

	t.Run("failing_because_expired", func(t *testing.T) {
		trade := newTradeCompleted()
		trade.ExpiryTime = uint64(time.Now().AddDate(0, 0, -1).Unix())
		require.True(t, trade.IsExpired())

		ok, err := trade.Settle(now)
		require.EqualError(t, err, domain.ErrTradeExpired.Error())
		require.False(t, ok)
		require.False(t, trade.IsSettled())
		require.True(t, trade.IsRejected())
	})

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
				name:  "with_trade_accepted",
				trade: newTradeAccepted(),
			},
		}

		for i := range tests {
			tt := tests[i]

			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				ok, err := tt.trade.Settle(now)
				require.EqualError(t, err, domain.ErrTradeMustBeCompleted.Error())
				require.False(t, ok)
				require.False(t, tt.trade.IsSettled())
			})
		}

	})
}

func newTradeEmpty() *domain.Trade {
	return domain.NewTrade()
}

func newTradeProposal() *domain.Trade {
	req := newMockedSwapRequest()
	now := time.Now()
	return &domain.Trade{
		ID:               uuid.New(),
		MarketQuoteAsset: req.GetAssetR(),
		MarketPrice: domain.Prices{
			BasePrice:  decimal.NewFromInt(int64(req.GetAmountR())).Div(decimal.NewFromInt(int64(req.GetAmountP()))),
			QuotePrice: decimal.NewFromInt(int64(req.GetAmountP())).Div(decimal.NewFromInt(int64(req.GetAmountR()))),
		},
		Status: domain.ProposalStatus,
		SwapRequest: domain.Swap{
			ID:        req.GetId(),
			Message:   randomBytes(100),
			Timestamp: uint64(now.Unix()),
		},
		ExpiryTime: uint64(now.Add(5 * time.Minute).Unix()),
	}
}

func newTradeAccepted() *domain.Trade {
	trade := newTradeProposal()
	acc := newMockedSwapAccept()

	trade.Status = domain.AcceptedStatus
	trade.SwapAccept = domain.Swap{
		ID:        acc.GetId(),
		Message:   randomBytes(100),
		Timestamp: uint64(time.Now().Unix()),
	}
	trade.TxID = randomHex(32)
	return trade
}

func newTradeCompleted() *domain.Trade {
	trade := newTradeAccepted()
	com := newMockedSwapComplete()

	trade.Status = domain.CompletedStatus
	trade.SwapComplete = domain.Swap{
		ID:        com.GetId(),
		Message:   randomBytes(100),
		Timestamp: uint64(time.Now().Unix()),
	}
	trade.TxHex = randomHex(100)
	return trade
}

func newTradeSettled() *domain.Trade {
	trade := newTradeCompleted()
	trade.ExpiryTime = 0
	trade.SettlementTime = uint64(time.Now().Unix())
	trade.Status = domain.SettledStatus
	return trade
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(len))
}

func randomBase64(len int) string {
	return base64.StdEncoding.EncodeToString(randomBytes(len))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	rand.Read(b)
	return b
}
