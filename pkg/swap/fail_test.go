package swap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSwap_Fail(t *testing.T) {
	id, message, err := Fail(FailOpts{
		MessageID: "a546bb0c",
		ErrCode:   ErrCodeInvalidSwapRequest,
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NotEmpty(t, message)
}
