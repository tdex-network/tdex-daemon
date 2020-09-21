package unspent

import (
	"context"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/storageutil/uow"
)

type Repository interface {
	AddUnspents(ctx context.Context, unspents []Unspent) error
	GetAllUnspents(ctx context.Context) []Unspent
	GetBalance(ctx context.Context, address string, assetHash string) uint64
	GetAvailableUnspents(ctx context.Context) []Unspent
	GetUnlockedBalance(ctx context.Context, address string, assetHash string) uint64
	LockUnspents(ctx context.Context, unspentKeys []UnspentKey, tradeID uuid.UUID) error
	UnlockUnspents(ctx context.Context, unspentKey []UnspentKey) error
	Begin() (uow.Tx, error)
	ContextKey() interface{}
}
