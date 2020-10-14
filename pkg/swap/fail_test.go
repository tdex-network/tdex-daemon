package swap

import (
	"testing"
)

func TestSwap_Fail(t *testing.T) {
	_, got, err := Fail(FailOpts{
		MessageID:  "a546bb0c",
		ErrCode:    ErrCodeInvalidSwapRequest,
		ErrMessage: "cumulative utxos count is not enough to cover SwapRequest.amount_p",
	})
	if err != nil {
		t.Errorf("Swap.Fail() error = %v ", err)
		return
	}
	want := make([]byte, 118)
	if len(got) != len(want) {
		t.Errorf("Swap.Fail() = %v, want %v", len(got), len(want))
	}
}
