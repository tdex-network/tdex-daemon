package unspent

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestLockUnlockUnSpents(t *testing.T) {
	u := Unspent{
		spent:  false,
		locked: false,
	}

	u.Spend()
	assert.Equal(t, u.spent, true)
	u.UnLock()
	assert.Equal(t, u.locked, false)
	u.Lock()
	assert.Equal(t, u.locked, true)
}
