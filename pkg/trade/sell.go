package trade

import (
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"
)

// Sell creates a new trade proposal with the given arguments for selling a
// certain amount of base asset of the provided market pair. This proposal is
// then sent to the server which the inner client is connected to and,
// eventually, the resulting SwapAccept serialized message returned by the
// counter-party is returned to the caller.
func (t *Trade) Sell(opts BuyOrSellOpts) ([]byte, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	return t.marketOrderRequest(
		opts.Market,
		tradetype.Sell,
		opts.Amount,
		opts.Address,
		opts.BlindingKey,
	)
}

// SellAndComplete creates a new trade proposal with the given arguments for
// selling a certain amount of base asset of the provided market pair.
// If the proposal is accepted, it's then counter-signed with the provided
// private key and sent back again to the server that will take care of
// finalizing and broadcasting the transaction
func (t *Trade) SellAndComplete(opts BuyOrSellAndCompleteOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}

	w := newWalletFromKey(opts.PrivateKey, opts.BlindingKey, t.network)
	swapAcceptMsg, err := t.marketOrderRequest(
		opts.Market,
		tradetype.Sell,
		opts.Amount,
		w.address(),
		opts.BlindingKey,
	)
	if err != nil {
		return "", err
	}

	return t.marketOrderComplete(swapAcceptMsg, w)
}
