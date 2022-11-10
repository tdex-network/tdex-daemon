package domain_test

import (
	"github.com/stretchr/testify/mock"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

/*
 * SwapParser
 */
type mockSwapParser struct {
	mock.Mock
}

func (m mockSwapParser) SerializeRequest(
	req domain.SwapRequest,
) ([]byte, int) {
	args := m.Called(req)

	var res []byte
	if a := args.Get(0); a != nil {
		res = a.([]byte)
	}

	var err int
	if a := args.Get(1); a != nil {
		err = a.(int)
	}
	return res, err
}

func (m mockSwapParser) SerializeAccept(
	reqMsg []byte, tx string, unblindedIns []domain.UnblindedInput,
) (string, []byte, int) {
	args := m.Called(reqMsg, tx, unblindedIns)

	var sres string
	if a := args.Get(0); a != nil {
		sres = a.(string)
	}

	var bres []byte
	if a := args.Get(1); a != nil {
		bres = a.([]byte)
	}

	var err int
	if args.Get(2) != nil {
		err = args.Get(2).(int)
	}

	return sres, bres, err
}

func (m mockSwapParser) SerializeComplete(
	accMsg []byte, tx string,
) (string, []byte, int) {
	args := m.Called(accMsg, tx)

	var sres string
	if a := args.Get(0); a != nil {
		sres = a.(string)
	}

	var bres []byte
	if a := args.Get(1); a != nil {
		bres = a.([]byte)
	}

	var err int
	if args.Get(2) != nil {
		err = args.Get(2).(int)
	}

	return sres, bres, err
}

func (m mockSwapParser) SerializeFail(id string, errCode int) (string, []byte) {
	args := m.Called(id, errCode)

	var sres string
	if a := args.Get(0); a != nil {
		sres = a.(string)
	}

	var bres []byte
	if a := args.Get(1); a != nil {
		bres = a.([]byte)
	}

	return sres, bres
}

func (m mockSwapParser) DeserializeRequest(msg []byte) *domain.SwapRequest {
	args := m.Called(msg)
	return args.Get(0).(*domain.SwapRequest)
}

func (m mockSwapParser) DeserializeAccept(msg []byte) *domain.SwapAccept {
	args := m.Called(msg)
	return args.Get(0).(*domain.SwapAccept)
}

func (m mockSwapParser) DeserializeComplete(msg []byte) *domain.SwapComplete {
	args := m.Called(msg)
	return args.Get(0).(*domain.SwapComplete)
}

func (m mockSwapParser) DeserializeFail(msg []byte) *domain.SwapFail {
	args := m.Called(msg)
	return args.Get(0).(*domain.SwapFail)
}

func (m mockSwapParser) ParseSwapTransaction(
	tx string,
) (*domain.SwapTransactionDetails, int) {
	args := m.Called(tx)
	var res *domain.SwapTransactionDetails
	if a := args.Get(0); a != nil {
		res = a.(*domain.SwapTransactionDetails)
	}
	var err int
	if a := args.Get(1); a != nil {
		err = a.(int)
	}
	return res, err
}
