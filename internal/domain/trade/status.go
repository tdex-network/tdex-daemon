package trade

import (
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

var (
	// EmptyStatus represent the status of an empty trade
	EmptyStatus = Status{
		code: pb.SwapStatus_UNDEFINED,
	}
	// ProposalStatus represent the status of a trade presented by some trader to
	// the daemon and not yet processed
	ProposalStatus = Status{
		code: pb.SwapStatus_REQUEST,
	}
	// ProposalRejectedStatus represents the status of a trade presented by some
	// trader to the daemon and rejected for some reason
	ProposalRejectedStatus = Status{
		code:   pb.SwapStatus_REQUEST,
		failed: true,
	}
	// AcceptedStatus represents the status of a trade proposal that has been
	// accepted by the daemon
	AcceptedStatus = Status{
		code: pb.SwapStatus_ACCEPT,
	}
	// FailedToCompleteStatus represents the status of a trade that failed to be
	// be completed for some reason
	FailedToCompleteStatus = Status{
		code:   pb.SwapStatus_ACCEPT,
		failed: true,
	}
	// CompletedStatus represents the status of a trade that has been completed
	// and accepted in mempool, waiting for being published on the blockchain
	CompletedStatus = Status{
		code: pb.SwapStatus_COMPLETE,
	}
)

// Status represents the different statuses that a trade between a trader and
// a provider can assume
type Status struct {
	code    pb.SwapStatus
	failed  bool
}

// Code returns the SwapStatus code of the current Status entity
func (s Status) Code() pb.SwapStatus {
	return s.code
}
