package swap

import (
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	"github.com/thanhpk/randstr"
	"google.golang.org/protobuf/proto"
)

const (
	ErrCodeInvalidSwapRequest = iota
	ErrCodeRejectedSwapRequest
	ErrCodeFailedToComplete
	ErrCodeInvalidTransaction
	ErrCodeBadPricingSwapRequest
	ErrCodeAborted
	ErrCodeFailedToBroadcast
)

var errMsg = map[int]string{
	ErrCodeInvalidSwapRequest:    "invalid swap request",
	ErrCodeRejectedSwapRequest:   "swap request not accepted",
	ErrCodeFailedToComplete:      "swap not completed",
	ErrCodeInvalidTransaction:    "invalid transaction format",
	ErrCodeBadPricingSwapRequest: "swap request price not accepted",
	ErrCodeAborted:               "aborted by counter-party ",
	ErrCodeFailedToBroadcast:     "swap completed but didn't get included in blockchain ",
}

type FailOpts struct {
	MessageID string
	ErrCode   int
}

func Fail(opts FailOpts) (string, []byte, error) {
	randomID := randstr.Hex(8)
	msgFail := &tdexv1.SwapFail{
		Id:             randomID,
		MessageId:      opts.MessageID,
		FailureCode:    uint32(opts.ErrCode),
		FailureMessage: errMsg[opts.ErrCode],
	}

	msgFailSerialized, err := proto.Marshal(msgFail)
	if err != nil {
		return "", nil, err
	}
	return randomID, msgFailSerialized, nil
}
