package esplora

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestGetBlockHeight(t *testing.T) {
	explorerSvc, err := NewService(explorerURL)
	if err != nil {
		t.Fatal(err)
	}

	height, err := explorerSvc.GetBlockHeight()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, height > 0)
}
