package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	// SwapParserManager ...
	SwapParserManager SwapParser
)

// Status represents the different statuses that a trade can assume.
type TradeStatus struct {
	Code   int
	Failed bool
}

// Trade is the data structure representing a trade entity.
type Trade struct {
	Id                  string
	MarketName          string
	MarketBaseAsset     string
	MarketQuoteAsset    string
	MarketPrice         MarketPrice
	MarketPercentageFee uint32
	MarketFixedBaseFee  uint64
	MarketFixedQuoteFee uint64
	TraderPubkey        []byte
	Status              TradeStatus
	PsetBase64          string
	TxId                string
	TxHex               string
	ExpiryTime          int64
	SettlementTime      int64
	SwapRequest         *Swap
	SwapAccept          *Swap
	SwapComplete        *Swap
	SwapFail            *Swap
}

// NewTrade returns a trade with a new id and Empty status.
func NewTrade() *Trade {
	return &Trade{Id: uuid.New().String(), Status: TradeStatus{Code: TradeStatusCodeUndefined}}
}

// Propose brings an Empty trade to the Propose status by first validating the
// provided arguments.
func (t *Trade) Propose(
	swapRequest SwapRequest,
	mktName, mktBaseAsset, mktQuoteAsset string,
	mktPercentageFee uint32, mktFixedBaseFee, mktFixedQuoteFee uint64,
	traderPubkey []byte,
) (bool, error) {
	if t.Status.Code >= TradeStatusCodeProposal {
		return true, nil
	}

	t.MarketName = mktName
	t.MarketBaseAsset = mktBaseAsset
	t.MarketQuoteAsset = mktQuoteAsset
	t.TraderPubkey = traderPubkey
	t.SwapRequest = &Swap{
		Id:        swapRequest.GetId(),
		Timestamp: time.Now().Unix(),
	}
	t.PsetBase64 = swapRequest.GetTransaction()
	t.Status.Code = TradeStatusCodeProposal

	msg, errCode := SwapParserManager.SerializeRequest(swapRequest)
	if errCode >= 0 {
		t.Fail(swapRequest.GetId(), errCode)
		return false, nil
	}

	price := calculateMarketMarketPrice(swapRequest, mktQuoteAsset)

	t.SwapRequest.Message = msg
	t.MarketPrice = price
	t.MarketPercentageFee = mktPercentageFee
	t.MarketFixedBaseFee = mktFixedBaseFee
	t.MarketFixedQuoteFee = mktFixedQuoteFee
	return true, nil
}

// Accept brings a trade from the Proposal to the Accepted status by validating
// the provided argument against the SwapRequest message and sets its
// expiration time.
func (t *Trade) Accept(
	psetBase64 string, unblindedIns []UnblindedInput, expiryTime int64,
) (bool, error) {
	if t.Status.Code >= TradeStatusCodeAccepted {
		return true, nil
	}

	if t.Status.Code != TradeStatusCodeProposal {
		return false, ErrTradeMustBeProposal
	}

	if !time.Unix(expiryTime, 0).After(time.Unix(t.SwapRequest.Timestamp, 0)) {
		return false, ErrTradeInvalidExpiryTime
	}

	swapAcceptID, swapAcceptMsg, errCode := SwapParserManager.SerializeAccept(
		t.SwapRequest.Message, psetBase64, unblindedIns,
	)
	if errCode >= 0 {
		t.Fail(t.SwapRequest.Id, errCode)
		return false, nil
	}

	t.ExpiryTime = expiryTime
	t.Status.Code = TradeStatusCodeAccepted
	t.SwapAccept = &Swap{
		Id:        swapAcceptID,
		Message:   swapAcceptMsg,
		Timestamp: time.Now().Unix(),
	}
	t.PsetBase64 = psetBase64

	return true, nil
}

// Complete brings a trade from the Accepted to the Completed status by
// checking that the given PSET completes the one of the SwapAccept message
// and by finalizing it and extracting the raw tx in hex format. Complete must
// be called before the trade expires, otherwise it won't be possible to
// actually complete an accepted trade.
func (t *Trade) Complete(tx string) (bool, error) {
	if t.Status.Code >= TradeStatusCodeCompleted {
		return true, nil
	}

	if !t.IsAccepted() {
		return false, ErrTradeMustBeAccepted
	}

	if t.IsExpired() {
		if !t.Status.Failed {
			t.Status.Failed = true
		}
		return false, ErrTradeExpired
	}

	txDetails, errCode := SwapParserManager.ParseSwapTransaction(tx)
	if errCode >= 0 {
		t.Fail(t.SwapAccept.Id, errCode)
		return false, nil
	}

	swapCompleteId, swapCompleteMsg, errCode := SwapParserManager.SerializeComplete(
		t.SwapAccept.Message,
		tx,
	)
	if errCode >= 0 {
		t.Fail(t.SwapAccept.Id, errCode)
		return false, nil
	}

	t.SwapComplete = &Swap{
		Id:        swapCompleteId,
		Message:   swapCompleteMsg,
		Timestamp: time.Now().Unix(),
	}
	t.Status.Code = TradeStatusCodeCompleted
	t.TxHex = txDetails.TxHex
	t.TxId = txDetails.Txid
	if len(t.PsetBase64) > 0 {
		t.PsetBase64 = txDetails.PsetBase64
	}
	return true, nil
}

// Settle brings the trade from the Completed to the Settled status, unsets the
// expiration time and adds the timestamp of the settlement (it must be a
// blocktime).
func (t *Trade) Settle(settlementTime int64) (bool, error) {
	if t.Status.Code == TradeStatusCodeSettled {
		return true, nil
	}

	if !(t.IsCompleted() || t.IsAccepted()) || t.Status.Failed {
		return false, ErrTradeMustBeCompletedOrAccepted
	}

	t.ExpiryTime = 0
	t.SettlementTime = settlementTime
	t.Status.Code = TradeStatusCodeSettled
	return true, nil
}

// Fail marks the current status of the trade as Failed and adds the SwapFail
// message.
func (t *Trade) Fail(swapID string, errCode int) {
	if t.Status.Failed {
		return
	}

	swapFailID, swapFailMsg := SwapParserManager.SerializeFail(swapID, errCode)
	t.SwapFail = &Swap{
		Id:      swapFailID,
		Message: swapFailMsg,
	}
	t.Status.Failed = true
}

// Expire brings the trade to the Expired status if its expiration date was
// previosly set. This infers that it must be in any of the Accepted, Completed,
// or related failed statuses. This method makes also sure that the expiration
// date has passed before changing the status.
func (t *Trade) Expire() (bool, error) {
	if t.Status.Code == TradeStatusCodeExpired {
		return true, nil
	}

	if t.ExpiryTime <= 0 {
		return false, ErrTradeNullExpiryTime
	}

	if time.Now().Before(time.Unix(t.ExpiryTime, 0)) {
		return false, ErrTradeExpiryTimeNotReached
	}

	t.Status.Code = TradeStatusCodeExpired
	return true, nil
}

// IsEmpty returns whether the Trade is empty.
func (t *Trade) IsEmpty() bool {
	return t.Status.Code == TradeStatusCodeUndefined
}

// IsProposal returns whether the trade is in Proposal status.
func (t *Trade) IsProposal() bool {
	return t.Status.Code == TradeStatusCodeProposal
}

// IsAccepted returns whether the trade is in Accepted status.
func (t *Trade) IsAccepted() bool {
	return t.Status.Code == TradeStatusCodeAccepted
}

// IsCompleted returns whether the trade is in Completed status.
func (t *Trade) IsCompleted() bool {
	return t.Status.Code == TradeStatusCodeCompleted
}

// IsSettled returns whether the trade is in Settled status.
func (t *Trade) IsSettled() bool {
	return t.Status.Code == TradeStatusCodeSettled
}

// IsRejected returns whether the trade has failed.
func (t *Trade) IsRejected() bool {
	return t.Status.Failed
}

// IsExpired returns whether the trade is in Expired status, or if its
// expiration date has passed.
func (t *Trade) IsExpired() bool {
	return t.Status.Code == TradeStatusCodeExpired ||
		(t.ExpiryTime > 0 && time.Now().After(time.Unix(t.ExpiryTime, 0)))
}

// ContainsSwap returns whether a swap identified by its id belongs to the
// current trade.
func (t *Trade) ContainsSwap(swapID string) bool {
	return swapID == t.SwapRequest.Id ||
		swapID == t.SwapAccept.Id ||
		swapID == t.SwapComplete.Id
}

// SwapRequestMessage returns the deserialized swap request message.
func (t *Trade) SwapRequestMessage() *SwapRequest {
	if t.IsEmpty() {
		return nil
	}
	return SwapParserManager.DeserializeRequest(t.SwapRequest.Message)
}

// SwapAcceptMessage returns the deserialized swap accept message, if defined.
func (t *Trade) SwapAcceptMessage() *SwapAccept {
	if t.IsEmpty() || t.IsProposal() {
		return nil
	}

	return SwapParserManager.DeserializeAccept(t.SwapAccept.Message)
}

// SwapCompleteMessage returns the deserialized swap complete message, if defined.
func (t *Trade) SwapCompleteMessage() *SwapComplete {
	if !t.IsCompleted() {
		return nil
	}

	return SwapParserManager.DeserializeComplete(t.SwapComplete.Message)
}

// SwapFailMessage returns the deserialized swap fail message, if defined.
func (t *Trade) SwapFailMessage() *SwapFail {
	if !t.IsRejected() {
		return nil
	}

	return SwapParserManager.DeserializeFail(t.SwapFail.Message)
}

func calculateMarketMarketPrice(
	swapRequest SwapRequest,
	marketQuoteAsset string,
) (price MarketPrice) {
	pricePR := decimal.NewFromInt(int64(swapRequest.GetAmountP())).Div(
		decimal.NewFromInt(int64(swapRequest.GetAmountR())),
	).Truncate(8)
	priceRP := decimal.NewFromInt(int64(swapRequest.GetAmountR())).Div(
		decimal.NewFromInt(int64(swapRequest.GetAmountP())),
	).Truncate(8)

	if swapRequest.GetAssetP() == marketQuoteAsset {
		price = MarketPrice{
			BasePrice:  priceRP.String(),
			QuotePrice: pricePR.String(),
		}
	} else {
		price = MarketPrice{
			BasePrice:  pricePR.String(),
			QuotePrice: priceRP.String(),
		}
	}
	return
}
