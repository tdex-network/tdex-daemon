package domain

import (
	"context"
	"github.com/google/uuid"
)

type UnspentRepository interface {
	AddUnspents(ctx context.Context, unspents []Unspent) error
	GetAllUnspents(ctx context.Context) []Unspent
	GetBalance(
		ctx context.Context,
		address string,
		assetHash string,
	) (uint64, error)
	GetAvailableUnspents(ctx context.Context) ([]Unspent, error)
	GetUnspentsForAddresses(
		ctx context.Context,
		addresses []string,
	) ([]Unspent, error)
	GetAvailableUnspentsForAddresses(
		ctx context.Context,
		addresses []string,
	) ([]Unspent, error)
	GetUnlockedBalance(
		ctx context.Context,
		address string,
		assetHash string,
	) (uint64, error)
	LockUnspents(
		ctx context.Context,
		unspentKeys []UnspentKey,
		tradeID uuid.UUID,
	) error
	UnlockUnspents(ctx context.Context, unspentKeys []UnspentKey) error
	UpdateUnspent(
		ctx context.Context,
		unspentKey UnspentKey,
		updateFn func(m *Unspent) (*Unspent, error),
	) error
	GetUnspentForKey(
		ctx context.Context,
		unspentKey UnspentKey,
	) (*Unspent, error)
	//Begin() (uow.Tx, error)
	//ContextKey() interface{}
}
