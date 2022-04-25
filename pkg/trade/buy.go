package trade

import (
	"encoding/hex"
	"errors"
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"

	"github.com/tdex-network/tdex-daemon/pkg/swap"
	"github.com/vulpemventures/go-elements/address"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrNullAddress ...
	ErrNullAddress = errors.New("address must not be null")
	// ErrInvalidAsset ...
	ErrInvalidAsset = errors.New("asset must be a 32-byte array in hex format")
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
	Asset       string
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
	if buf, err := hex.DecodeString(o.Asset); err != nil || len(buf) != 32 {
		return ErrInvalidAsset
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
		opts.Asset,
		opts.Address,
		opts.BlindingKey,
	)
}

// BuyOrSellAndCompleteOpts is the struct given to Buy method
type BuyOrSellAndCompleteOpts struct {
	Market      trademarket.Market
	Amount      uint64
	Asset       string
	PrivateKey  []byte
	BlindingKey []byte
}

func (o BuyOrSellAndCompleteOpts) validate() error {
	if err := o.Market.Validate(); err != nil {
		return err
	}
	if o.Amount <= 0 {
		return ErrInvalidAmount
	}
	if buf, err := hex.DecodeString(o.Asset); err != nil || len(buf) != 32 {
		return ErrInvalidAsset
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
		opts.Asset,
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
	asset string,
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
		Asset:     asset,
	})
	if err != nil {
		return nil, err
	}

	outputScript, _ := address.ToOutputScript(addr)
	outputScriptHex := hex.EncodeToString(outputScript)

	psetBase64, err := NewSwapTx(
		unspents,
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
		AssetToSend:        preview.AssetToSend,
		AmountToSend:       preview.AmountToSend,
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
	swapAccept := &tdexv1.SwapAccept{}
	proto.Unmarshal(swapAcceptMsg, swapAccept)

	psetBase64 := swapAccept.GetTransaction()
	signedPset, err := w.Sign(psetBase64)
	if err != nil {
		return "", err
	}

	_, swapCompleteMsg, err := swap.Complete(swap.CompleteOpts{
		Message:     swapAcceptMsg,
		Transaction: signedPset,
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
