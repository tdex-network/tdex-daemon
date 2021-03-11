package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func TestSpendUnspent(t *testing.T) {
	t.Parallel()

	u := domain.Unspent{}
	require.False(t, u.IsSpent())

	u.Spend()
	require.True(t, u.IsSpent())
}

func TestConfirmUnspent(t *testing.T) {
	t.Parallel()

	u := domain.Unspent{}
	require.False(t, u.IsConfirmed())

	u.Confirm()
	require.True(t, u.IsConfirmed())
}

func TestLockUnlockUnspent(t *testing.T) {
	t.Parallel()

	u := domain.Unspent{}
	require.False(t, u.IsLocked())

	tradeID := uuid.New()
	err := u.Lock(&tradeID)
	require.NoError(t, err)
	require.True(t, u.IsLocked())

	u.Unlock()
	require.False(t, u.IsLocked())
}

func TestFailingLockUnspent(t *testing.T) {
	t.Parallel()

	u := domain.Unspent{}
	require.False(t, u.IsLocked())

	tradeID := uuid.New()
	err := u.Lock(&tradeID)
	require.NoError(t, err)
	require.True(t, u.IsLocked())

	err = u.Lock(&tradeID)
	require.NoError(t, err)

	otherTradeID := uuid.New()
	err = u.Lock(&otherTradeID)
	require.EqualError(t, err, domain.ErrUnspentAlreadyLocked.Error())
}
