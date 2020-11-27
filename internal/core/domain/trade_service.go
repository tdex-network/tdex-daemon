package domain

import (
	"github.com/vulpemventures/go-elements/pset"
	"time"

	"github.com/tdex-network/tdex-daemon/config"
	pkgswap "github.com/tdex-network/tdex-daemon/pkg/swap"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	"google.golang.org/protobuf/proto"
)

// Propose returns a new trade proposal for the given trader and market
func (t *Trade) Propose(swapRequest *pb.SwapRequest, marketQuoteAsset string, traderPubkey []byte) (bool, error) {
	if !t.IsEmpty() {
		return false, ErrMustBeEmpty
	}

	t.TraderPubkey = traderPubkey
	t.MarketQuoteAsset = marketQuoteAsset
	t.SwapRequest.ID = swapRequest.GetId()
	t.Timestamp.Request = uint64(time.Now().Unix())
	t.Timestamp.Expiry = t.Timestamp.Request + uint64(config.GetInt(config.
		TradeExpiryTimeKey))
	t.PsetBase64 = swapRequest.GetTransaction()

	msg, err := pkgswap.ParseSwapRequest(swapRequest)
	if err != nil {
		t.Fail(
			swapRequest.GetId(),
			ProposalRejectedStatus,
			pkgswap.ErrCodeInvalidSwapRequest,
			err.Error(),
		)
		return false, nil
	}

	t.Status = ProposalStatus
	t.SwapRequest.Message = msg
	return true, nil
}

// Accept attempts to accept a trade proposal. The trade must be in Proposal
// status to be accepted, otherwise an error is thrown
func (t *Trade) Accept(
	psetBase64 string,
	inputBlindingKeys,
	outputBlindingKeys map[string][]byte,
) (bool, error) {
	if t.Status != ProposalStatus {
		return false, ErrMustBeProposal
	}

	swapAcceptID, swapAcceptMsg, err := pkgswap.Accept(pkgswap.AcceptOpts{
		Message:            t.SwapRequest.Message,
		PsetBase64:         psetBase64,
		InputBlindingKeys:  inputBlindingKeys,
		OutputBlindingKeys: outputBlindingKeys,
	})
	if err != nil {
		t.Fail(
			t.SwapRequest.ID,
			ProposalRejectedStatus,
			pkgswap.ErrCodeRejectedSwapRequest,
			err.Error(),
		)
		return false, nil
	}

	t.Status = AcceptedStatus
	t.SwapAccept.ID = swapAcceptID
	t.SwapAccept.Message = swapAcceptMsg
	t.Timestamp.Accept = uint64(time.Now().Unix())
	t.PsetBase64 = psetBase64

	p, err := pset.NewPsetFromBase64(psetBase64)
	if err != nil {
		return false, err
	}
	t.TxID = p.UnsignedTx.TxHash().String()

	return true, nil
}

// CompleteResult is return type of Complete method
type CompleteResult struct {
	OK    bool
	TxHex string
	TxID  string
}

// Complete sets the status of the trade to Complete by adding the txID
// of the tx in the blockchain. The trade must be in Accepted or
// FailedToComplete status for being completed, otherwise an error is thrown
func (t *Trade) Complete(psetBase64 string) (*CompleteResult, error) {
	if t.IsCompleted() {
		return &CompleteResult{OK: true, TxHex: t.TxHex, TxID: t.TxID}, nil
	}

	if !t.IsAccepted() {
		return nil, ErrMustBeAccepted
	}

	if err := pkgswap.ValidateCompletePset(pkgswap.ValidateCompletePsetOpts{
		PsetBase64:         psetBase64,
		InputBlindingKeys:  t.SwapAcceptMessage().GetInputBlindingKey(),
		OutputBlindingKeys: t.SwapAcceptMessage().GetOutputBlindingKey(),
		SwapRequest:        t.SwapRequestMessage(),
	}); err != nil {
		t.Fail(
			t.SwapAccept.ID,
			FailedToCompleteStatus,
			pkgswap.ErrCodeFailedToComplete,
			err.Error(),
		)
		return &CompleteResult{OK: false}, nil
	}

	opts := wallet.FinalizeAndExtractTransactionOpts{
		PsetBase64: psetBase64,
	}
	txHex, txHash, err := wallet.FinalizeAndExtractTransaction(opts)
	if err != nil {
		t.Fail(
			t.SwapAccept.ID,
			FailedToCompleteStatus,
			pkgswap.ErrCodeFailedToComplete,
			err.Error(),
		)
		return &CompleteResult{OK: false}, nil
	}

	swapCompleteID, swapCompleteMsg, err := pkgswap.Complete(pkgswap.CompleteOpts{
		Message:    t.SwapAccept.Message,
		PsetBase64: psetBase64,
	})
	if err != nil {
		t.Fail(
			t.SwapAccept.ID,
			FailedToCompleteStatus,
			pkgswap.ErrCodeFailedToComplete,
			err.Error(),
		)
		return &CompleteResult{OK: false}, nil
	}

	t.SwapComplete.ID = swapCompleteID
	t.SwapComplete.Message = swapCompleteMsg
	t.PsetBase64 = psetBase64
	t.TxID = txHash
	t.TxHex = txHex
	return &CompleteResult{OK: true, TxHex: txHex, TxID: txHash}, nil
}

func (t *Trade) Settle(settlementTime uint64) error {
	t.Status = CompletedStatus
	return t.AddBlocktime(settlementTime)
}

// Fail sets the status of the trade to the provided status and creates the
// serialized SwapFail message and id for the given swap, errCode and errMsg
func (t *Trade) Fail(swapID string, tradeStatus Status, errCode pkgswap.ErrCode, errMsg string) {
	swapFailID, swapFailMsg, _ := pkgswap.Fail(pkgswap.FailOpts{
		MessageID:  swapID,
		ErrCode:    errCode,
		ErrMessage: errMsg,
	})
	t.SwapFail.ID = swapFailID
	t.SwapFail.Message = swapFailMsg
	t.Status = tradeStatus
}

// AddBlocktime sets the timestamp for a completed trade to the given blocktime.
// If the trade is not in Complete status, an error is thrown
func (t *Trade) AddBlocktime(blocktime uint64) error {
	if !t.IsCompleted() {
		return ErrMustBeCompleted
	}

	t.Timestamp.Complete = blocktime
	return nil
}

// IsEmpty returns whether the Trade is empty
func (t *Trade) IsEmpty() bool {
	return t.Status == EmptyStatus
}

// IsProposal returns whether the trade is in Proposal status
func (t *Trade) IsProposal() bool {
	return t.Status == ProposalStatus || t.Status == ProposalRejectedStatus
}

// IsAccepted returns whether the trade is in Accepted status
func (t *Trade) IsAccepted() bool {
	return t.Status == AcceptedStatus || t.Status == FailedToCompleteStatus
}

// IsCompleted returns whether the trade is in Completed status
func (t *Trade) IsCompleted() bool {
	return t.Status == CompletedStatus
}

// IsRejected returns whether the trade is in ProposalRejected status
func (t *Trade) IsRejected() bool {
	return t.Status == ProposalRejectedStatus
}

// IsExpired returns whether the trade has reached the expiration date, ie if
// now is after the expriation date and the trade is not in Complete status
func (t *Trade) IsExpired() bool {
	now := uint64(time.Now().Unix())
	return now >= t.Timestamp.Expiry && !t.IsCompleted()
}

// ContainsSwap returns whether some swap identified by an id belongs to the
// current trade
func (t *Trade) ContainsSwap(swapID string) bool {
	return swapID == t.SwapRequest.ID || swapID == t.SwapAccept.
		ID || swapID == t.SwapComplete.ID
}

// SwapRequestMessage returns the swap request made for the trade
func (t *Trade) SwapRequestMessage() *pb.SwapRequest {
	if t.IsEmpty() {
		return nil
	}
	s := &pb.SwapRequest{}
	proto.Unmarshal(t.SwapRequest.Message, s)
	return s
}

// SwapAcceptMessage returns the swap accept made for the trade
func (t *Trade) SwapAcceptMessage() *pb.SwapAccept {
	if t.IsEmpty() || t.IsProposal() {
		return nil
	}

	s := &pb.SwapAccept{}
	proto.Unmarshal(t.SwapAccept.Message, s)
	return s
}

// SwapCompleteMessage returns the swap complete message for the trade if existing
func (t *Trade) SwapCompleteMessage() *pb.SwapComplete {
	if !t.IsCompleted() {
		return nil
	}

	s := &pb.SwapComplete{}
	proto.Unmarshal(t.SwapComplete.Message, s)
	return s
}

// SwapFailMessage returns the swap fail message for the trade if existing
func (t *Trade) SwapFailMessage() *pb.SwapFail {
	if !t.IsRejected() {
		return nil
	}

	s := &pb.SwapFail{}
	proto.Unmarshal(t.SwapFail.Message, s)
	return s
}

// SwapRequestTime returns the timestamp of the proposal of the current trade
func (t *Trade) SwapRequestTime() uint64 {
	return t.Timestamp.Request
}

// SwapAcceptTime returns the timestamp of when the trade proposal has been
// accepted
func (t *Trade) SwapAcceptTime() uint64 {
	return t.Timestamp.Accept
}

// SwapCompleteTime returns the timestamp of when the accepted trade has been
// completed
func (t *Trade) SwapCompleteTime() uint64 {
	return t.Timestamp.Complete
}

// SwapExpiryTime returns the timestamp of when the current trade will expire
func (t *Trade) SwapExpiryTime() uint64 {
	return t.Timestamp.Expiry
}
