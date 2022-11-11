package swap_parser

import (
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/swap"
	"github.com/vulpemventures/go-elements/psetv2"
	"github.com/vulpemventures/go-elements/transaction"
	"google.golang.org/protobuf/proto"
)

type service struct{}

func NewService() domain.SwapParser {
	return service{}
}

func (s service) SerializeRequest(
	r domain.SwapRequest,
) ([]byte, int) {
	unblindedIns := make([]swap.UnblindedInput, 0, len(r.GetUnblindedInputs()))
	for _, in := range r.GetUnblindedInputs() {
		unblindedIns = append(unblindedIns, swap.UnblindedInput(in))
	}
	msg, err := swap.Request(swap.RequestOpts{
		Id:              r.GetId(),
		AssetToSend:     r.GetAssetP(),
		AmountToSend:    r.GetAmountP(),
		AssetToReceive:  r.GetAssetR(),
		AmountToReceive: r.GetAmountR(),
		PsetBase64:      r.GetTransaction(),
		UnblindedInputs: unblindedIns,
	})
	if err != nil {
		return nil, swap.ErrCodeInvalidSwapRequest
	}
	return msg, -1
}

func (s service) SerializeAccept(
	reqMsg []byte, tx string, unblindedInputs []domain.UnblindedInput,
) (string, []byte, int) {
	unblindedIns := make([]swap.UnblindedInput, 0, len(unblindedInputs))
	for _, in := range unblindedInputs {
		unblindedIns = append(unblindedIns, swap.UnblindedInput(in))
	}
	id, msg, err := swap.Accept(swap.AcceptOpts{
		Message:         reqMsg,
		PsetBase64:      tx,
		UnblindedInputs: unblindedIns,
	})
	if err != nil {
		return "", nil, swap.ErrCodeRejectedSwapRequest
	}
	return id, msg, -1
}

func (s service) SerializeComplete(
	accMsg []byte, tx string,
) (string, []byte, int) {
	id, msg, err := swap.Complete(swap.CompleteOpts{
		Message:     accMsg,
		Transaction: tx,
	})
	if err != nil {
		return "", nil, swap.ErrCodeFailedToComplete
	}
	return id, msg, -1
}

func (s service) SerializeFail(id string, errCode int) (string, []byte) {
	id, msg, _ := swap.Fail(swap.FailOpts{
		MessageID: id,
		ErrCode:   errCode,
	})
	return id, msg
}

func (s service) DeserializeRequest(msg []byte) *domain.SwapRequest {
	swap := &tdexv1.SwapRequest{}
	//nolint
	proto.Unmarshal(msg, swap)
	unblindedIns := make([]domain.UnblindedInput, 0, len(swap.GetUnblindedInputs()))
	for _, in := range swap.GetUnblindedInputs() {
		unblindedIns = append(unblindedIns, domain.UnblindedInput{
			Index:         in.GetIndex(),
			Asset:         in.GetAsset(),
			Amount:        in.GetAmount(),
			AmountBlinder: in.GetAmountBlinder(),
			AssetBlinder:  in.GetAssetBlinder(),
		})
	}
	return &domain.SwapRequest{
		Id:              swap.GetId(),
		AssetP:          swap.GetAssetP(),
		AmountP:         swap.GetAmountP(),
		AssetR:          swap.GetAssetR(),
		AmountR:         swap.GetAmountR(),
		Transaction:     swap.GetTransaction(),
		UnblindedInputs: unblindedIns,
	}
}

func (s service) DeserializeAccept(msg []byte) *domain.SwapAccept {
	swap := &tdexv1.SwapAccept{}
	//nolint
	proto.Unmarshal(msg, swap)
	unblindedIns := make([]domain.UnblindedInput, 0, len(swap.GetUnblindedInputs()))
	for _, in := range swap.GetUnblindedInputs() {
		unblindedIns = append(unblindedIns, domain.UnblindedInput{
			Index:         in.GetIndex(),
			Asset:         in.GetAsset(),
			Amount:        in.GetAmount(),
			AmountBlinder: in.GetAmountBlinder(),
			AssetBlinder:  in.GetAssetBlinder(),
		})
	}
	return &domain.SwapAccept{
		Id:              swap.GetId(),
		RequestId:       swap.GetRequestId(),
		Transaction:     swap.GetTransaction(),
		UnblindedInputs: unblindedIns,
	}
}

func (s service) DeserializeComplete(msg []byte) *domain.SwapComplete {
	swap := &tdexv1.SwapComplete{}
	//nolint
	proto.Unmarshal(msg, swap)
	return &domain.SwapComplete{
		Id:          swap.GetId(),
		AcceptId:    swap.GetAcceptId(),
		Transaction: swap.GetTransaction(),
	}
}

func (s service) DeserializeFail(msg []byte) *domain.SwapFail {
	swap := &tdexv1.SwapFail{}
	//nolint
	proto.Unmarshal(msg, swap)
	return &domain.SwapFail{
		Id:             swap.GetId(),
		MessageId:      swap.GetMessageId(),
		FailureCode:    swap.GetFailureCode(),
		FailureMessage: swap.GetFailureMessage(),
	}
}

func (s service) ParseSwapTransaction(tx string) (*domain.SwapTransactionDetails, int) {
	if isPset(tx) {
		ptx, _ := psetv2.NewPsetFromBase64(tx)
		//nolint
		psetv2.FinalizeAll(ptx)
		t, _ := psetv2.Extract(ptx)
		txhex, _ := t.ToHex()
		return &domain.SwapTransactionDetails{
			PsetBase64: tx,
			TxHex:      txhex,
			Txid:       t.TxHash().String(),
		}, -1
	}
	if isHex(tx) {
		t, _ := transaction.NewTxFromHex(tx)
		return &domain.SwapTransactionDetails{
			TxHex: tx,
			Txid:  t.TxHash().String(),
		}, -1
	}
	return nil, swap.ErrCodeInvalidTransaction
}

func isPset(tx string) bool {
	_, err := psetv2.NewPsetFromBase64(tx)
	return err == nil
}

func isHex(tx string) bool {
	_, err := transaction.NewTxFromHex(tx)
	return err == nil
}
