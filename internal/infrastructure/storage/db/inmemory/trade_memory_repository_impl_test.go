package inmemory

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func TestGetOrCreateTrade(t *testing.T) {
	db := newMockDb()
	tradeRepository := NewTradeRepositoryImpl(db)

	tradeID, _ := uuid.Parse("5440a53e-58d2-4e3d-8380-20410e687589")
	trade, err := tradeRepository.GetOrCreateTrade(ctx, &tradeID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "mqa2", trade.MarketQuoteAsset)
}

func TestGetAllTrades(t *testing.T) {
	db := newMockDb()
	tradeRepository := NewTradeRepositoryImpl(db)

	// try to get unknow trade
	tradeID := uuid.New()
	_, err := tradeRepository.GetOrCreateTrade(ctx, &tradeID)
	assert.NotNil(t, err)

	// create a trade by passing a nil trade id
	if _, err := tradeRepository.GetOrCreateTrade(ctx, nil); err != nil {
		t.Fatal(err)
	}

	trades, err := tradeRepository.GetAllTrades(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3+1, len(trades))
}

func TestGetAllTradesByMarket(t *testing.T) {
	db := newMockDb()
	tradeRepository := NewTradeRepositoryImpl(db)

	trades, err := tradeRepository.GetAllTradesByMarket(ctx, "mqa2")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(trades))
}

func TestGetTradeBySwapAcceptID(t *testing.T) {
	db := newMockDb()
	tradeRepository := NewTradeRepositoryImpl(db)

	trade, err := tradeRepository.GetTradeBySwapAcceptID(ctx, "22")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "424", trade.TxID)
}

func TestUpdateTrade(t *testing.T) {
	db := newMockDb()
	tradeRepository := NewTradeRepositoryImpl(db)

	tradeID, err := uuid.Parse("5440a53e-58d2-4e3d-8380-20410e687589")
	if err != nil {
		t.Fatal(err)
	}

	err = tradeRepository.UpdateTrade(
		ctx,
		&tradeID,
		func(t *domain.Trade) (*domain.Trade, error) {
			t.Price = 100
			return t, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	trade, err := tradeRepository.GetOrCreateTrade(ctx, &tradeID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, float32(100), trade.Price)
}
