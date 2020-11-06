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
		addresses []string,
		assetHash string,
	) (uint64, error)
	GetAvailableUnspents(ctx context.Context) ([]Unspent, error)
	GetAllUnspentsForAddresses(
		ctx context.Context,
		addresses []string,
	) ([]Unspent, error)
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
		addresses []string,
		assetHash string,
	) (uint64, error)
	SpendUnspents(
		ctx context.Context,
		unspentKeys []UnspentKey,
	) error
	ConfirmUnspents(
		ctx context.Context,
		unspentKeys []UnspentKey,
	) error
	LockUnspents(
		ctx context.Context,
		unspentKeys []UnspentKey,
		tradeID uuid.UUID,
	) error
	UnlockUnspents(ctx context.Context, unspentKeys []UnspentKey) error
	GetUnspentForKey(
		ctx context.Context,
		unspentKey UnspentKey,
	) (*Unspent, error)
}
