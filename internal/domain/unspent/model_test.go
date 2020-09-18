package unspent

import (
	"testing"

	"github.com/google/uuid"
	"github.com/magiconair/properties/assert"
)

func TestLockUnlockUnSpents(t *testing.T) {
	u := Unspent{
		spent:  false,
		locked: false,
	}
	tradeID := uuid.New()

	u.Spend()
	assert.Equal(t, u.spent, true)
	u.UnLock()
	assert.Equal(t, u.locked, false)
	u.Lock(&tradeID)
	assert.Equal(t, u.locked, true)
}
