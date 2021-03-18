package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Propose brings an Empty trade to the Propose status by first validating the
// provided arguments.
func (t *Trade) Propose(
	swapRequest SwapRequest,
	marketQuoteAsset string,
	marketFeeBasisPoint int64,
	traderPubkey []byte,
) (bool, error) {
	if t.Status.Code >= Proposal {
		return true, nil
	}

	now := uint64(time.Now().Unix())
	t.TraderPubkey = traderPubkey
	t.MarketQuoteAsset = marketQuoteAsset
	t.SwapRequest.ID = swapRequest.GetId()
	t.SwapRequest.Timestamp = now
	t.PsetBase64 = swapRequest.GetTransaction()
	t.Status = ProposalStatus

	msg, err := SwapParserManager.SerializeRequest(swapRequest)
	if err != nil {
		t.Fail(
			swapRequest.GetId(),
			err.Code,
			err.Error(),
		)
		return false, nil
	}

	price := calculateMarketPrices(swapRequest, marketQuoteAsset)

	t.SwapRequest.Message = msg
	t.MarketPrice = price
	t.MarketFee = marketFeeBasisPoint
	return true, nil
}

// Accept brings a trade from the Proposal to the Accepted status by validating
// the provided arguemtn against the the SwapRequest message and sets its
// expiration time.
func (t *Trade) Accept(
	psetBase64 string,
	inputBlindingKeys,
	outputBlindingKeys map[string][]byte,
	expiryDuration uint64,
) (bool, error) {
	if t.Status.Code >= Accepted {
		return true, nil
	}

	if t.Status != ProposalStatus {
		return false, ErrTradeMustBeProposal
	}

	swapAcceptID, swapAcceptMsg, err := SwapParserManager.SerializeAccept(AcceptArgs{
		RequestMessage:     t.SwapRequest.Message,
		Transaction:        psetBase64,
		InputBlindingKeys:  inputBlindingKeys,
		OutputBlindingKeys: outputBlindingKeys,
	})
	if err != nil {
		t.Fail(
			t.SwapRequest.ID,
			err.Code,
			err.Error(),
		)
		return false, nil
	}

	now := uint64(time.Now().Unix())
	t.ExpiryTime = now + expiryDuration
	t.Status = AcceptedStatus
	t.SwapAccept.ID = swapAcceptID
	t.SwapAccept.Message = swapAcceptMsg
	t.SwapAccept.Timestamp = now
	t.PsetBase64 = psetBase64
	t.TxID, _ = PsetParserManager.GetTxID(psetBase64)

	return true, nil
}

// CompleteResult is return type of Complete method.
type CompleteResult struct {
	OK    bool
	TxHex string
	TxID  string
}

// Complete brings a trade from the Accepted to the Completed status by
// checiking that the given PSET completes the one of the SwapAccept message
// and by finalizing it and extracting the raw tx in hex format. Complete must
// be called before the trade expires, otherwise it won't be possible to
// actually complete an accepted trade.
func (t *Trade) Complete(psetBase64 string) (*CompleteResult, error) {
	if t.Status.Code >= Completed {
		return &CompleteResult{OK: true, TxHex: t.TxHex, TxID: t.TxID}, nil
	}

	if !t.IsAccepted() {
		return nil, ErrTradeMustBeAccepted
	}

	if t.IsExpired() {
		if !t.Status.Failed {
			t.Status.Failed = true
		}
		return nil, ErrTradeExpired
	}

	swapCompleteID, swapCompleteMsg, err := SwapParserManager.SerializeComplete(
		t.SwapRequest.Message,
		t.SwapAccept.Message,
		psetBase64,
	)
	if err != nil {
		t.Fail(
			t.SwapAccept.ID,
			err.Code,
			err.Error(),
		)
		return &CompleteResult{OK: false}, nil
	}

	txHex, _ := PsetParserManager.GetTxHex(psetBase64)
	t.SwapComplete.ID = swapCompleteID
	t.SwapComplete.Message = swapCompleteMsg
	t.Status = CompletedStatus
	t.PsetBase64 = psetBase64
	t.TxHex = txHex
	return &CompleteResult{OK: true, TxHex: txHex, TxID: t.TxID}, nil
}

// Settle brings the trade from the Completed to the Settled status, unsets the
// expiration time and adds the timestamp of the settlement (it must be a
// blocktime).
func (t *Trade) Settle(settlementTime uint64) (bool, error) {
	if t.Status.Code == Settled {
		return true, nil
	}

	if !t.IsCompleted() && !t.Status.Failed {
		return false, ErrTradeMustBeCompleted
	}

	t.ExpiryTime = 0
	t.SettlementTime = settlementTime
	t.Status = SettledStatus
	return true, nil
}

// Fail marks the current status of the trade as Failed and adds the SwapFail
// message.
func (t *Trade) Fail(swapID string, errCode int, errMsg string) {
	if t.Status.Failed {
		return
	}

	swapFailID, swapFailMsg := SwapParserManager.SerializeFail(swapID, errCode, errMsg)
	t.SwapFail.ID = swapFailID
	t.SwapFail.Message = swapFailMsg
	t.Status.Failed = true
}

// IsEmpty returns whether the Trade is empty.
func (t *Trade) IsEmpty() bool {
	return t.Status == EmptyStatus
}

// IsProposal returns whether the trade is in Proposal status.
func (t *Trade) IsProposal() bool {
	return t.Status.Code == Proposal
}

// IsAccepted returns whether the trade is in Accepted status.
func (t *Trade) IsAccepted() bool {
	return t.Status.Code == Accepted
}

// IsCompleted returns whether the trade is in Completed status.
func (t *Trade) IsCompleted() bool {
	return t.Status.Code == Completed
}

// IsSettled returns whether the trade is in Settled status.
func (t *Trade) IsSettled() bool {
	return t.Status.Code == Settled
}

// IsRejected returns whether the trade has failed.
func (t *Trade) IsRejected() bool {
	return t.Status.Failed
}

// IsExpired returns whether the trade has reached the expiration date, ie if
// the trade is not in Settle status and now is after the expriation date
func (t *Trade) IsExpired() bool {
	now := uint64(time.Now().Unix())
	return !t.IsSettled() && !t.IsEmpty() && now >= t.ExpiryTime
}

// ContainsSwap returns whether a swap identified by its id belongs to the
// current trade.
func (t *Trade) ContainsSwap(swapID string) bool {
	return swapID == t.SwapRequest.ID ||
		swapID == t.SwapAccept.ID ||
		swapID == t.SwapComplete.ID
}

// SwapRequestMessage returns the deserialized swap request message.
func (t *Trade) SwapRequestMessage() SwapRequest {
	if t.IsEmpty() {
		return nil
	}
	s, _ := SwapParserManager.DeserializeRequest(t.SwapRequest.Message)
	return s
}

// SwapAcceptMessage returns the deserialized swap accept message, if defined.
func (t *Trade) SwapAcceptMessage() SwapAccept {
	if t.IsEmpty() || t.IsProposal() {
		return nil
	}

	s, _ := SwapParserManager.DeserializeAccept(t.SwapAccept.Message)
	return s
}

// SwapCompleteMessage returns the deserialized swap complete message, if defined.
func (t *Trade) SwapCompleteMessage() SwapComplete {
	if !t.IsCompleted() {
		return nil
	}

	s, _ := SwapParserManager.DeserializeComplete(t.SwapComplete.Message)
	return s
}

// SwapFailMessage returns the deserialized swap fail message, if defined.
func (t *Trade) SwapFailMessage() SwapFail {
	if !t.IsRejected() {
		return nil
	}

	s, _ := SwapParserManager.DeserializeFail(t.SwapFail.Message)
	return s
}

func calculateMarketPrices(
	swapRequest SwapRequest,
	marketQuoteAsset string,
) (price Prices) {
	pricePR := decimal.NewFromInt(int64(swapRequest.GetAmountP())).Div(
		decimal.NewFromInt(int64(swapRequest.GetAmountR())),
	).Truncate(8)
	priceRP := decimal.NewFromInt(int64(swapRequest.GetAmountR())).Div(
		decimal.NewFromInt(int64(swapRequest.GetAmountP())),
	).Truncate(8)

	if swapRequest.GetAssetP() == marketQuoteAsset {
		price = Prices{
			BasePrice:  priceRP,
			QuotePrice: pricePR,
		}
	} else {
		price = Prices{
			BasePrice:  pricePR,
			QuotePrice: priceRP,
		}
	}
	return
}
