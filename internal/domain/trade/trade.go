package trade

import (
	"errors"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/swap"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
)

var (
	// ErrMustBeEmpty ...
	ErrMustBeEmpty = errors.New(
		"trade must be empty for parsing a proposal",
	)
	// ErrMustBeProposal ...
	ErrMustBeProposal = errors.New(
		"trade must be in proposal state for being accepted",
	)
	// ErrMustBeAccepted ...
	ErrMustBeAccepted = errors.New(
		"trade must be in accepted state for being completed",
	)
	// ErrMustBeCompleted ...
	ErrMustBeCompleted = errors.New(
		"trade must be in completed state to add txid",
	)
	// ErrExpirationDateNotReached ...
	ErrExpirationDateNotReached = errors.New(
		"trade did not reached expiration date yet and cannot be set expired",
	)
)

type timestamp struct {
	request    uint64
	accept     uint64
	complete   uint64
	expiry uint64
}

// Trade defines the Trade entity data structure for holding swap transactions
type Trade struct {
	id          id
	marketIndex int
	traderID    []byte
	status      Status
	psetBase64  string
	txID        string
	price       float32
	timestamp   timestamp
	swapMessage swapMessage
}

// NewTrade returns an empty trade
func NewTrade() *Trade {
	return &Trade{}
}

// Proposal returns a new trade proposal for the given trader and market
func (t *Trade) Proposal(swapRequest *pb.SwapRequest, marketIndex int, traderID []byte) error {
	if !t.IsEmpty() {
		return ErrMustBeEmpty
	}

	t.traderID = traderID
	t.marketIndex = marketIndex
	t.id.request = swapRequest.GetId()
	t.timestamp.request = uint64(time.Now().Unix())
	t.psetBase64 = swapRequest.GetTransaction()
	// TODO: set the expiration date to t.timestamp.request + expiry_time for
	// for keeping consistency of data presented to traders and operators

	msg, err := swap.ParseSwapRequest(swapRequest)
	if err != nil {
		t.status = ProposalRejectedStatus
		return err
	}
	// TODO: check price

	t.status = ProposalStatus
	t.swapMessage.request = msg
	return nil
}

// Accept attempts to accept a trade proposal. The trade must be in Proposal
// status to be accepted, otherwise an error is thrown
func (t *Trade) Accept(psetBase64 string) error {
	if t.status != ProposalStatus {
		return ErrMustBeProposal
	}

	swapAcceptID, swapAcceptMsg, err := swap.Accept(swap.AcceptOpts{
		Message:    t.swapMessage.request,
		PsetBase64: psetBase64,
	})
	if err != nil {
		t.status = ProposalRejectedStatus
		return err
	}

	t.status = AcceptedStatus
	t.id.accept = swapAcceptID
	t.swapMessage.accept = swapAcceptMsg
	t.psetBase64 = psetBase64
	return nil
}

// Complete checks the signatures and finalizes the trade. The trade must be in
// Accepted or FailedToComplete status for being completed, otherwise an
// error is thrown
func (t *Trade) Complete(psetBase64 string) error {
	if !t.IsAccepted() {
		return ErrMustBeAccepted
	}

	swapCompleteID, swapCompleteMsg, err := swap.Complete(swap.CompleteOpts{
		Message:    t.swapMessage.accept,
		PsetBase64: psetBase64,
	})
	if err != nil {
		t.status = FailedToCompleteStatus
		return err
	}

	t.status = CompletedStatus
	t.id.complete = swapCompleteID
	t.swapMessage.complete = swapCompleteMsg
	t.psetBase64 = psetBase64
	return nil
}

// Published adds the txId of the swap transaction included in the blockchain
// to the trade. The trade must be in Completed status to set it as Published,
// otherwise an error is thrown
func (t *Trade) Published(txID string) error {
	if !t.IsCompleted() {
		return ErrMustBeCompleted
	}

	t.txID = txID
	return nil
}

// Expired sets the status of the trade to AcceptedAndExpired. The trade must
// be in Accepted or FailedToComplete status for expiring, otherwise an error
// is thrown
func (t *Trade) Expired() error {
	if t.status != AcceptedStatus && t.status == FailedToCompleteStatus {
		return ErrMustBeAccepted
	}

	// TODO: check that now is actually after expiration date.
	t.status = AcceptedAndExpiredStatus
	return nil
}

// IsEmpty returns whether the Trade is empty
func (t *Trade) IsEmpty() bool {
	return t.status == EmptyStatus
}

// IsProposal returns whether the trade is in Proposal status
func (t *Trade) IsProposal() bool {
	return t.status == ProposalStatus || t.status == ProposalRejectedStatus
}

// IsAccepted returns whether the trade is in Accepted status
func (t *Trade) IsAccepted() bool {
	return t.status == AcceptedStatus || t.status == FailedToCompleteStatus
}

// IsCompleted returns whether the trade is in Completed status
func (t *Trade) IsCompleted() bool {
	return t.status == CompletedStatus
}

// IsExpired returns whether the trade has reached the expiration date
func (t *Trade) IsExpired() bool {
	return t.status == AcceptedAndExpiredStatus
}

// ID returns the ids by witch a trade is univoquely identified.
// At the very list a trade is identified by swapRequestID and swapAcceptID.
// If present, also the swapCompleteID is added to the returned list of ids.
func (t *Trade) ID() []string {
	return t.id.id()
}

// MarketIndex returns the market that the trade is referring to
func (t *Trade) MarketIndex() int {
	return t.marketIndex
}

// TraderID returns the id of the trader who proposed the current trade
func (t *Trade) TraderID() []byte {
	return t.traderID
}

// Status returns the current status of the trade
func (t *Trade) Status() Status {
	return t.status
}

// SwapRequestMessage returns the swap request made for the trade
func (t *Trade) SwapRequestMessage() *pb.SwapRequest {
	if t.IsEmpty() {
		return nil
	}
	return t.swapMessage.Request()
}

// SwapAcceptMessage returns the swap accept made for the trade
func (t *Trade) SwapAcceptMessage() *pb.SwapAccept {
	if t.IsEmpty() || t.IsProposal() {
		return nil
	}
	return t.swapMessage.Accept()
}

// SwapCompleteMessage returns the swap complete message for the trade if existing
func (t *Trade) SwapCompleteMessage() *pb.SwapComplete {
	if !t.IsCompleted() {
		return nil
	}
	return t.swapMessage.Complete()
}

// SwapRequestTime returns the timestamp of the proposal of the current trade
func (t *Trade) SwapRequestTime() uint64 {
	return t.timestamp.request
}

// SwapAcceptTime returns the timestamp of when the trade proposal has been
// accepted
func (t *Trade) SwapAcceptTime() uint64 {
	return t.timestamp.accept
}

// SwapCompleteTime returns the timestamp of when the accepted trade has been
// completed
func (t *Trade) SwapCompleteTime() uint64 {
	return t.timestamp.complete
}

// SwapExpiryTime returns the timestamp of when the current trade will expire
func (t *Trade) SwapExpiryTime() uint64 {
	return t.timestamp.expiry
}
