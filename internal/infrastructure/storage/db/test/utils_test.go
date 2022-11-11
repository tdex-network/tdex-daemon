package db_test

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func makeRandomMarket() *domain.Market {
	market, _ := domain.NewMarket(
		randomHex(32), randomHex(32), uint32(randomIntInRange(0, 9999)),
	)
	return market
}

func makeRandomTrade() *domain.Trade {
	trade := domain.NewTrade()
	trade.MarketBaseAsset = randomHex(32)
	trade.MarketQuoteAsset = randomHex(32)
	trade.MarketName = randomHex(20)
	return trade
}

func makeRandomDeposits(num int) []domain.Deposit {
	deposits := make([]domain.Deposit, 0, num)
	for i := 0; i < num; i++ {
		deposits = append(deposits, domain.Deposit{
			AccountName: randomHex(20),
			TxID:        randomHex(32),
			Timestamp:   randomTimestamp(),
		})
	}
	return deposits
}

func makeRandomWithdrawals(num int) []domain.Withdrawal {
	withdrawals := make([]domain.Withdrawal, 0, num)
	for i := 0; i < num; i++ {
		withdrawals = append(withdrawals, domain.Withdrawal{
			AccountName: randomHex(20),
			TxID:        randomHex(32),
			Timestamp:   randomTimestamp(),
		})
	}
	return withdrawals
}

func randomTimestamp() int64 {
	return int64(randomIntInRange(1000000000, 1662688000))
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(32))
}

func randomId() string {
	return uuid.New().String()
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	//nolint
	rand.Read(b)
	return b
}

func randomIntInRange(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(int(n.Int64())) + min
}

type page struct {
	number int64
	size   int64
}

func (p page) GetSize() int64 {
	return p.size
}

func (p page) GetNumber() int64 {
	return p.number
}
