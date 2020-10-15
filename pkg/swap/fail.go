package swap

import (
	"fmt"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	"github.com/thanhpk/randstr"
	"google.golang.org/protobuf/proto"
)

type ErrCode int

const (
	ErrCodeInvalidSwapRequest ErrCode = iota
	ErrCodeRejectedSwapRequest
	ErrCodeFailedToComplete
)

var errMsg = map[ErrCode]string{
	ErrCodeInvalidSwapRequest:  "invalid swap request",
	ErrCodeRejectedSwapRequest: "swap request not accepted",
	ErrCodeFailedToComplete:    "swap not completed",
}

type FailOpts struct {
	MessageID  string
	ErrCode    ErrCode
	ErrMessage string
}

func Fail(opts FailOpts) (string, []byte, error) {
	randomID := randstr.Hex(8)
	msgFail := &pb.SwapFail{
		Id:             randomID,
		MessageId:      opts.MessageID,
		FailureCode:    uint32(opts.ErrCode),
		FailureMessage: fmt.Sprintf("%s: %s", errMsg[opts.ErrCode], opts.ErrMessage),
	}

	msgFailSerialized, err := proto.Marshal(msgFail)
	if err != nil {
		return "", nil, err
	}
	return randomID, msgFailSerialized, nil
}
