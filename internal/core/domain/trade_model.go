package domain

import (
	"encoding/hex"
	"reflect"

	"github.com/google/uuid"
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"
	pkgswap "github.com/tdex-network/tdex-daemon/pkg/swap"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"google.golang.org/protobuf/proto"
)

var (
	// EmptyStatus represent the status of an empty trade.
	EmptyStatus = Status{
		Code: Empty,
	}
	// ProposalStatus represent the status of a trade presented by some trader to
	// the daemon and not yet processed.
	ProposalStatus = Status{
		Code: Proposal,
	}
	// ProposalRejectedStatus represents the status of a trade presented by some
	// trader to the daemon and rejected for some reason.
	ProposalRejectedStatus = Status{
		Code:   Proposal,
		Failed: true,
	}
	// AcceptedStatus represents the status of a trade proposal that has been
	// accepted by the daemon
	AcceptedStatus = Status{
		Code: Accepted,
	}
	// FailedToCompleteStatus represents the status of a trade that failed to be
	// be completed for some reason.
	FailedToCompleteStatus = Status{
		Code:   Accepted,
		Failed: true,
	}
	// CompletedStatus represents the status of a trade that has been completed
	// and accepted in mempool, waiting for being published on the blockchain.
	CompletedStatus = Status{
		Code: Completed,
	}
	// FailedToSettleStatus represents the status of a trade that failed to be
	// be settled for some reason.
	FailedToSettleStatus = Status{
		Code:   Completed,
		Failed: true,
	}
	// SettledStatus represents the status of a trade that has been settled,
	// meaning that has been included into the blockchain.
	SettledStatus = Status{
		Code: Settled,
	}
	// ExpiredStatus represents the status of a trade that has been expired,
	// meaning that it was at least accepted, but not settled within the
	// expiration time frame.
	ExpiredStatus = Status{
		Code:   Expired,
		Failed: true,
	}
)

// AcceptArgs ...
type AcceptArgs struct {
	RequestMessage     []byte
	Transaction        string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
}

// SwapError is the special error returned by the ISwapParser when serializing
// a swap message.
type SwapError struct {
	Err  error
	Code int
}

func (s *SwapError) Error() string {
	return s.Err.Error()
}

// SwapParser defines the required methods to override the default swap
// message parser, which is grpc-proto.
type SwapParser interface {
	SerializeRequest(r SwapRequest) ([]byte, *SwapError)
	SerializeAccept(args AcceptArgs) (string, []byte, *SwapError)
	SerializeComplete(accMsg []byte, tx string) (string, []byte, *SwapError)
	SerializeFail(id string, code int, msg string) (string, []byte)

	DeserializeRequest(msg []byte) (SwapRequest, error)
	DeserializeAccept(msg []byte) (SwapAccept, error)
	DeserializeComplete(msg []byte) (SwapComplete, error)
	DeserializeFail(msg []byte) (SwapFail, error)
}

type swapParser struct{}

func (p swapParser) SerializeRequest(r SwapRequest) ([]byte, *SwapError) {
	msg, err := pkgswap.Request(pkgswap.RequestOpts{
		Id:                 r.GetId(),
		AssetToSend:        r.GetAssetP(),
		AmountToSend:       r.GetAmountP(),
		AssetToReceive:     r.GetAssetR(),
		AmountToReceive:    r.GetAmountR(),
		PsetBase64:         r.GetTransaction(),
		InputBlindingKeys:  r.GetInputBlindingKey(),
		OutputBlindingKeys: r.GetOutputBlindingKey(),
	})
	if err != nil {
		return nil, &SwapError{err, int(pkgswap.ErrCodeInvalidSwapRequest)}
	}
	return msg, nil
}

func (p swapParser) SerializeAccept(args AcceptArgs) (string, []byte, *SwapError) {
	id, msg, err := pkgswap.Accept(pkgswap.AcceptOpts{
		Message:            args.RequestMessage,
		PsetBase64:         args.Transaction,
		InputBlindingKeys:  args.InputBlindingKeys,
		OutputBlindingKeys: args.OutputBlindingKeys,
	})
	if err != nil {
		return "", nil, &SwapError{err, int(pkgswap.ErrCodeRejectedSwapRequest)}
	}
	return id, msg, nil
}

func (p swapParser) SerializeComplete(accMsg []byte, tx string) (string, []byte, *SwapError) {
	id, msg, err := pkgswap.Complete(pkgswap.CompleteOpts{
		Message:     accMsg,
		Transaction: tx,
	})
	if err != nil {
		return "", nil, &SwapError{err, int(pkgswap.ErrCodeFailedToComplete)}
	}

	// If the tx is not in hex format, let's make sure that the pset can be
	// finalized  and the final raw transaction extracted.
	if _, err := hex.DecodeString(tx); err != nil {
		if _, _, err := wallet.FinalizeAndExtractTransaction(wallet.FinalizeAndExtractTransactionOpts{
			PsetBase64: tx,
		}); err != nil {
			return "", nil, &SwapError{err, int(pkgswap.ErrCodeFailedToComplete)}
		}
	}

	return id, msg, nil
}

func (p swapParser) SerializeFail(id string, errCode int, errMsg string) (string, []byte) {
	id, msg, _ := pkgswap.Fail(pkgswap.FailOpts{
		MessageID:  id,
		ErrMessage: errMsg,
		ErrCode:    pkgswap.ErrCode(errCode),
	})
	return id, msg
}

func (p swapParser) DeserializeRequest(msg []byte) (SwapRequest, error) {
	s := &tdexv1.SwapRequest{}
	if err := proto.Unmarshal(msg, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (p swapParser) DeserializeAccept(msg []byte) (SwapAccept, error) {
	s := &tdexv1.SwapAccept{}
	if err := proto.Unmarshal(msg, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (p swapParser) DeserializeComplete(msg []byte) (SwapComplete, error) {
	s := &tdexv1.SwapComplete{}
	if err := proto.Unmarshal(msg, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (p swapParser) DeserializeFail(msg []byte) (SwapFail, error) {
	s := &tdexv1.SwapFail{}
	if err := proto.Unmarshal(msg, s); err != nil {
		return nil, err
	}
	return s, nil
}

// PsetParser defines the required methods to override the extraction of the
// txid and of the final transaction in hex format from the PSET one.
// The default one comes from go-elements.
type PsetParser interface {
	GetTxID(psetBase64 string) (string, error)
	GetTxHex(psetBase64 string) (string, error)
}

type psetManager struct{}

func (m psetManager) GetTxID(psetBase64 string) (string, error) {
	return transactionutil.GetTxIdFromPset(psetBase64)
}

func (m psetManager) GetTxHex(psetBase64 string) (string, error) {
	txHex, _, err := wallet.FinalizeAndExtractTransaction(
		wallet.FinalizeAndExtractTransactionOpts{PsetBase64: psetBase64},
	)
	if err != nil {
		return "", err
	}
	return txHex, nil
}

var (
	// SwapParserManager ...
	SwapParserManager SwapParser
	// PsetParserManager ...
	PsetParserManager PsetParser
)

func init() {
	SwapParserManager = swapParser{}
	PsetParserManager = psetManager{}
}

// SwapRequest is the abstracted representation of a SwapRequest message.
type SwapRequest interface {
	GetId() string
	GetAssetP() string
	GetAmountP() uint64
	GetAssetR() string
	GetAmountR() uint64
	GetTransaction() string
	GetInputBlindingKey() map[string][]byte
	GetOutputBlindingKey() map[string][]byte
}

// SwapAccept is the abstracted representation of a SwapAccept message.
type SwapAccept interface {
	GetId() string
	GetRequestId() string
	GetTransaction() string
	GetInputBlindingKey() map[string][]byte
	GetOutputBlindingKey() map[string][]byte
}

// SwapComplete is the abstracted representation of a SwapComplete message.
type SwapComplete interface {
	GetId() string
	GetAcceptId() string
	GetTransaction() string
}

// SwapFail is the abstracted representation of a SwapFail message.
type SwapFail interface {
	GetId() string
	GetMessageId() string
	GetFailureCode() uint32
	GetFailureMessage() string
}

// Swap is the data structure that represents any of the above swaps.
type Swap struct {
	ID        string
	Message   []byte
	Timestamp uint64
}

// IsZero ...
func (s Swap) IsZero() bool {
	return reflect.DeepEqual(s, Swap{})
}

// Status represents the different statuses that a trade can assume.
type Status struct {
	Code   int
	Failed bool
}

// Trade is the data structure representing a trade entity.
type Trade struct {
	ID                  uuid.UUID
	MarketBaseAsset     string
	MarketQuoteAsset    string
	MarketPrice         Prices
	MarketFee           int64
	MarketFixedBaseFee  int64
	MarketFixedQuoteFee int64
	TraderPubkey        []byte
	Status              Status
	PsetBase64          string
	TxID                string
	TxHex               string
	ExpiryTime          uint64
	SettlementTime      uint64
	SwapRequest         Swap
	SwapAccept          Swap
	SwapComplete        Swap
	SwapFail            Swap
}

// NewTrade returns a trade with a new id and Empty status.
func NewTrade() *Trade {
	return &Trade{ID: uuid.New(), Status: EmptyStatus}
}
