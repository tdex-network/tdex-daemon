package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLockUnlockUnSpents(t *testing.T) {
	u := Unspent{
		Spent:  false,
		Locked: false,
	}
	tradeID := uuid.New()

	u.Spend()
	assert.Equal(t, true, u.IsSpent())
	u.Unlock()
	assert.Equal(t, false, u.IsLocked())
	u.Lock(&tradeID)
	assert.Equal(t, true, u.IsLocked())
}
