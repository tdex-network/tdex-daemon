package domain

import (
	"github.com/google/uuid"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

var (
	// EmptyStatus represent the status of an empty trade
	EmptyStatus = Status{
		Code: pb.SwapStatus_UNDEFINED,
	}
	// ProposalStatus represent the status of a trade presented by some trader to
	// the daemon and not yet processed
	ProposalStatus = Status{
		Code: pb.SwapStatus_REQUEST,
	}
	// ProposalRejectedStatus represents the status of a trade presented by some
	// trader to the daemon and rejected for some reason
	ProposalRejectedStatus = Status{
		Code:   pb.SwapStatus_REQUEST,
		Failed: true,
	}
	// AcceptedStatus represents the status of a trade proposal that has been
	// accepted by the daemon
	AcceptedStatus = Status{
		Code: pb.SwapStatus_ACCEPT,
	}
	// FailedToCompleteStatus represents the status of a trade that failed to be
	// be completed for some reason
	FailedToCompleteStatus = Status{
		Code:   pb.SwapStatus_ACCEPT,
		Failed: true,
	}
	// CompletedStatus represents the status of a trade that has been completed
	// and accepted in mempool, waiting for being published on the blockchain
	CompletedStatus = Status{
		Code: pb.SwapStatus_COMPLETE,
	}
)

type Timestamp struct {
	Request  uint64
	Accept   uint64
	Complete uint64
	Expiry   uint64
}

type Swap struct {
	ID      string
	Message []byte
}

// Status represents the different statuses that a trade between a trader and
// a provider can assume
type Status struct {
	Code   pb.SwapStatus
	Failed bool
}

// Trade defines the Trade entity data structure for holding swap transactions
type Trade struct {
	ID               uuid.UUID
	MarketQuoteAsset string
	TraderPubkey     []byte
	Status           Status
	PsetBase64       string
	TxID             string
	Price            float32
	MarketFee        int64
	Timestamp        Timestamp
	SwapRequest      Swap
	SwapAccept       Swap
	SwapComplete     Swap
	SwapFail         Swap
}

// NewTrade returns an empty trade
func NewTrade() *Trade {
	return &Trade{ID: uuid.New(), Status: EmptyStatus}
}
