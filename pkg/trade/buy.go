package trade

import (
	"encoding/hex"
	"errors"
	"fmt"

	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"

	"github.com/tdex-network/tdex-daemon/pkg/swap"
	"github.com/vulpemventures/go-elements/address"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrNullAddress ...
	ErrNullAddress = errors.New("address must not be null")
	// ErrInvalidAddress ...
	ErrInvalidAddress = errors.New("address is not valid")
	// ErrNullPrivateKey ...
	ErrNullPrivateKey = errors.New("private key must not be null")
	// ErrNullBlindingKey ...
	ErrNullBlindingKey = errors.New("blinding key must not be null")
)

// BuyOrSellOpts is the struct given to Buy/Sell method
type BuyOrSellOpts struct {
	Market      trademarket.Market
	Amount      uint64
	Address     string
	BlindingKey []byte
}

func (o BuyOrSellOpts) validate() error {
	if err := o.Market.Validate(); err != nil {
		return err
	}
	if o.Amount <= 0 {
		return ErrInvalidAmount
	}
	if len(o.Address) <= 0 {
		return ErrNullAddress
	}
	if _, err := address.ToOutputScript(o.Address); err != nil {
		return ErrInvalidAddress
	}
	if len(o.BlindingKey) <= 0 {
		return ErrNullBlindingKey
	}
	return nil
}

// Buy creates a new trade proposal with the given arguments and sends it to
// the server which the inner client is connected to. This method returns the
// SwapAccept serialized message eventually returned by the counter-party.
func (t *Trade) Buy(opts BuyOrSellOpts) ([]byte, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	return t.marketOrderRequest(
		opts.Market,
		tradetype.Buy,
		opts.Amount,
		opts.Address,
		opts.BlindingKey,
	)
}

// BuyOrSellAndCompleteOpts is the struct given to Buy method
type BuyOrSellAndCompleteOpts struct {
	Market      trademarket.Market
	TradeType   int
	Amount      uint64
	PrivateKey  []byte
	BlindingKey []byte
}

func (o BuyOrSellAndCompleteOpts) validate() error {
	if err := o.Market.Validate(); err != nil {
		return err
	}
	if err := tradetype.TradeType(o.TradeType).Validate(); err != nil {
		return err
	}
	if o.Amount <= 0 {
		return ErrInvalidAmount
	}
	if len(o.PrivateKey) <= 0 {
		return ErrNullPrivateKey
	}
	if len(o.BlindingKey) <= 0 {
		return ErrNullBlindingKey
	}
	return nil
}

// BuyAndComplete creates a new trade proposal with the give arguments. The
// transaction of the resulting SwapAccept message is then signed with the
// provided private key, and sent back again to the connected server, which
// will take care of finalizing and broadcasting it.
func (t *Trade) BuyAndComplete(opts BuyOrSellAndCompleteOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}

	w := NewWalletFromKey(opts.PrivateKey, opts.BlindingKey, t.network)
	swapAcceptMsg, err := t.marketOrderRequest(
		opts.Market,
		tradetype.Buy,
		opts.Amount,
		w.Address(),
		opts.BlindingKey,
	)
	if err != nil {
		return "", err
	}

	return t.marketOrderComplete(swapAcceptMsg, w)
}

func (t *Trade) marketOrderRequest(
	market trademarket.Market,
	tradeType tradetype.TradeType,
	amount uint64,
	addr string,
	blindingKey []byte,
) ([]byte, error) {
	unspents, err := t.explorer.GetUnspents(addr, [][]byte{blindingKey})
	if err != nil {
		return nil, err
	}
	if len(unspents) <= 0 {
		return nil, fmt.Errorf("address '%s' is not funded", addr)
	}

	preview, err := t.Preview(PreviewOpts{
		Market:    market,
		TradeType: int(tradeType),
		Amount:    amount,
	})
	if err != nil {
		return nil, err
	}

	outputScript, _ := address.ToOutputScript(addr)
	outputScriptHex := hex.EncodeToString(outputScript)

	psetBase64, err := NewSwapTx(
		unspents,
		blindingKey,
		preview.AssetToSend,
		preview.AmountToSend,
		preview.AssetToReceive,
		preview.AmountToReceive,
		outputScript,
	)
	if err != nil {
		return nil, err
	}

	blindingKeyMap := map[string][]byte{
		outputScriptHex: blindingKey,
	}

	swapRequestMsg, err := swap.Request(swap.RequestOpts{
		AssetToBeSent:      preview.AssetToSend,
		AmountToBeSent:     preview.AmountToSend,
		AssetToReceive:     preview.AssetToReceive,
		AmountToReceive:    preview.AmountToReceive,
		PsetBase64:         psetBase64,
		InputBlindingKeys:  blindingKeyMap,
		OutputBlindingKeys: blindingKeyMap,
	})
	if err != nil {
		return nil, err
	}

	reply, err := t.client.TradePropose(tradeclient.TradeProposeOpts{
		Market:      market,
		SwapRequest: swapRequestMsg,
		TradeType:   tradeType,
	})

	if fail := reply.GetSwapFail(); fail != nil {
		return nil, fmt.Errorf("trade proposal has been rejected for reason: %s", fail.GetFailureMessage())
	}

	return proto.Marshal(reply.GetSwapAccept())
}

func (t *Trade) marketOrderComplete(swapAcceptMsg []byte, w *Wallet) (string, error) {
	swapAccept := &pb.SwapAccept{}
	proto.Unmarshal(swapAcceptMsg, swapAccept)

	psetBase64 := swapAccept.GetTransaction()
	signedPset, err := w.Sign(psetBase64)
	if err != nil {
		return "", err
	}

	_, swapCompleteMsg, err := swap.Complete(swap.CompleteOpts{
		Message:    swapAcceptMsg,
		PsetBase64: signedPset,
	})
	if err != nil {
		return "", err
	}

	reply, err := t.client.TradeComplete(tradeclient.TradeCompleteOpts{
		SwapComplete: swapCompleteMsg,
	})
	if err != nil {
		return "", err
	}
	if swapFail := reply.GetSwapFail(); swapFail != nil {
		return "", errors.New(swapFail.GetFailureMessage())
	}

	return reply.GetTxid(), nil
}
