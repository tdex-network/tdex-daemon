package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/storageutil/uow"
)

type UnspentRepository interface {
	AddUnspents(ctx context.Context, unspents []Unspent) error
	GetAllUnspents(ctx context.Context) []Unspent
	GetBalance(ctx context.Context, address string, assetHash string) uint64
	GetAvailableUnspents(ctx context.Context) []Unspent
	GetAvailableUnspentsForAddresses(ctx context.Context, addresses []string) []Unspent
	GetUnlockedBalance(ctx context.Context, address string, assetHash string) uint64
	LockUnspents(ctx context.Context, unspentKeys []UnspentKey, tradeID uuid.UUID) error
	UnlockUnspents(ctx context.Context, unspentKey []UnspentKey) error
	Begin() (uow.Tx, error)
	ContextKey() interface{}
	GetBalanceInfoForAsset(unspents []Unspent) map[string]BalanceInfo
}
