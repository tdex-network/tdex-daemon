package trade

import (
	"errors"
	"time"

	pkgswap "github.com/tdex-network/tdex-daemon/pkg/swap"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/config"
	"google.golang.org/protobuf/proto"
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
	request  uint64
	accept   uint64
	complete uint64
	expiry   uint64
}

type swap struct {
	id      string
	message []byte
}

// Trade defines the Trade entity data structure for holding swap transactions
type Trade struct {
	id               uuid.UUID
	marketQuoteAsset string
	traderPubkey     []byte
	status           Status
	psetBase64       string
	txID             string
	price            float32
	timestamp        timestamp
	swapRequest      swap
	swapAccept       swap
	swapComplete     swap
}

// NewTrade returns an empty trade
func NewTrade() *Trade {
	return &Trade{id: uuid.New(), status: EmptyStatus}
}

// Propose returns a new trade proposal for the given trader and market
func (t *Trade) Propose(swapRequest *pb.SwapRequest, marketQuoteAsset string, traderPubkey []byte) error {
	if !t.IsEmpty() {
		return ErrMustBeEmpty
	}

	t.traderPubkey = traderPubkey
	t.marketQuoteAsset = marketQuoteAsset
	t.swapRequest.id = swapRequest.GetId()
	t.timestamp.request = uint64(time.Now().Unix())
	t.timestamp.expiry = t.timestamp.request + uint64(config.GetInt(config.TradeExpiryTimeKey))
	t.psetBase64 = swapRequest.GetTransaction()

	msg, err := pkgswap.ParseSwapRequest(swapRequest)
	if err != nil {
		t.status = ProposalRejectedStatus
		return err
	}
	// TODO: check price

	t.status = ProposalStatus
	t.swapRequest.message = msg
	return nil
}

// Accept attempts to accept a trade proposal. The trade must be in Proposal
// status to be accepted, otherwise an error is thrown
func (t *Trade) Accept(
	psetBase64 string,
	inputBlindingKeys,
	outputBlindingKeys map[string][]byte,
) error {
	if t.status != ProposalStatus {
		return ErrMustBeProposal
	}

	swapAcceptID, swapAcceptMsg, err := pkgswap.Accept(pkgswap.AcceptOpts{
		Message:            t.swapRequest.message,
		PsetBase64:         psetBase64,
		InputBlindingKeys:  inputBlindingKeys,
		OutputBlindingKeys: outputBlindingKeys,
	})
	if err != nil {
		t.status = ProposalRejectedStatus
		return err
	}

	t.status = AcceptedStatus
	t.swapAccept.id = swapAcceptID
	t.swapAccept.message = swapAcceptMsg
	t.timestamp.accept = uint64(time.Now().Unix())
	t.psetBase64 = psetBase64
	return nil
}

// Complete sets the status of the trade to Complete by adding the txID
// of the tx in the blockchain. The trade must be in Accepted or
// FailedToComplete status for being completed, otherwise an error is thrown
func (t *Trade) Complete(psetBase64 string, txID string) error {
	if !t.IsAccepted() {
		return ErrMustBeAccepted
	}

	swapCompleteID, swapCompleteMsg, err := pkgswap.Complete(pkgswap.CompleteOpts{
		Message:    t.swapAccept.message,
		PsetBase64: psetBase64,
	})
	if err != nil {
		t.status = FailedToCompleteStatus
		return err
	}

	t.status = CompletedStatus
	t.swapComplete.id = swapCompleteID
	t.swapComplete.message = swapCompleteMsg
	t.psetBase64 = psetBase64
	t.txID = txID
	return nil
}

// AddBlockTime sets the timestamp for a completed trade to the given blocktime.
// If the trade is not in Complete status, an error is thrown
func (t *Trade) AddBlocktime(blocktime uint64) error {
	if !t.IsCompleted() {
		return ErrMustBeCompleted
	}

	t.timestamp.complete = blocktime
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

// IsExpired returns whether the trade has reached the expiration date, ie if
// now is after the expriation date and the trade is not in Complete status
func (t *Trade) IsExpired() bool {
	now := uint64(time.Now().Unix())
	return now >= t.timestamp.expiry && !t.IsCompleted()
}

// ContainsSwap returns whether some swap identified by an id belongs to the
// current trade
func (t *Trade) ContainsSwap(swapID string) bool {
	return swapID == t.swapRequest.id || swapID == t.swapAccept.id || swapID == t.swapComplete.id
}

// ID returns the ids by witch a trade is univoquely identified.
// At the very least a trade is identified by swapRequestID and swapAcceptID.
// If present, also the swapCompleteID is added to the returned list of ids.
func (t *Trade) ID() uuid.UUID {
	return t.id
}

// MarketQuoteAsset returns the quote asset of the market that the trade is
// referring to
func (t *Trade) MarketQuoteAsset() string {
	return t.marketQuoteAsset
}

// TraderPubkey returns the EC pubkey (check BOTD#2) of the trader who proposed
// the current trade
func (t *Trade) TraderPubkey() []byte {
	return t.traderPubkey
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
	s := &pb.SwapRequest{}
	proto.Unmarshal(t.swapRequest.message, s)
	return s
}

// SwapAcceptMessage returns the swap accept made for the trade
func (t *Trade) SwapAcceptMessage() *pb.SwapAccept {
	if t.IsEmpty() || t.IsProposal() {
		return nil
	}

	s := &pb.SwapAccept{}
	proto.Unmarshal(t.swapAccept.message, s)
	return s
}

// SwapCompleteMessage returns the swap complete message for the trade if existing
func (t *Trade) SwapCompleteMessage() *pb.SwapComplete {
	if !t.IsCompleted() {
		return nil
	}

	s := &pb.SwapComplete{}
	proto.Unmarshal(t.swapComplete.message, s)
	return s
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
