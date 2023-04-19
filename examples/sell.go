package examples

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/btcsuite/btcd/txscript"
	"github.com/google/uuid"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/psetv2"
)

func SellExample(daemonAddr, explorerAddr string) error {
	fmt.Printf("---------- SELL LBTC FOR USDt ----------\n\n")

	traderAmount := 0.001

	client, explorer, err := initServices(daemonAddr, explorerAddr)
	if err != nil {
		return err
	}

	traderWallet := newKeyPair()

	fmt.Printf("Trader wallet:\n%s\n\n", traderWallet)

	fmt.Printf("Sending LBTCs to trader...\n\n")
	utxos, err := faucet(explorer, traderWallet, lbtc, traderAmount)
	if err != nil {
		return fmt.Errorf("failed to send LBTC funds to trader: %s", err)
	}

	fmt.Printf("Fetching market from provider...\n\n")
	markets, err := fetchMarkets(client)
	if err != nil {
		return fmt.Errorf("failed to fetch provider markets: %s", err)
	}

	if len(markets) == 0 {
		return fmt.Errorf("provider has no tradable markets")
	}

	targetMarket := markets[0].GetMarket()
	fmt.Printf("Market: %s\n\n", targetMarket)

	usdt := targetMarket.GetQuoteAsset()

	fmt.Printf("Making trade preview...\n\n")
	swapRequest, feeAsset, feeAmount, err := makeTradeSellPreview(
		client, targetMarket, lbtc, usdt, traderAmount, traderWallet, utxos,
	)
	if err != nil {
		return fmt.Errorf("failed to make trade preview: %s", err)
	}

	fmt.Printf("Making trade proposal...\n\n")
	swapAccept, err := makeTradeSellProposal(
		client, targetMarket, swapRequest, feeAsset, feeAmount,
	)
	if err != nil {
		return fmt.Errorf("failed to make trade proposal: %s", err)
	}

	fmt.Printf("Signing and completing trade...\n\n")
	txid, err := signAndCompleteTrade(client, traderWallet, swapAccept)
	if err != nil {
		return fmt.Errorf("failed to complete trade: %s", err)
	}

	fmt.Printf("Completed trade with txid: %s\n\n", txid)
	swapRequest.Transaction = ""
	swapRequest.UnblindedInputs = swapAccept.GetUnblindedInputs()
	fmt.Printf("Swap info: %s\n\n", swapRequest)

	return nil
}

func makeTradeSellPreview(
	client tdexv2.TradeServiceClient,
	market *tdexv2.Market, asset, feeAsset string, amount float64,
	w wallet, utxos []explorer.Utxo,
) (*tdexv2.SwapRequest, string, uint64, error) {
	satsAmount := uint64(amount * math.Pow10(8))
	res, err := client.PreviewTrade(context.Background(), &tdexv2.PreviewTradeRequest{
		Market:   market,
		Type:     tdexv2.TradeType_TRADE_TYPE_SELL,
		Amount:   satsAmount,
		Asset:    asset,
		FeeAsset: feeAsset,
	})
	if err != nil {
		return nil, "", 0, err
	}
	preview := res.GetPreviews()[0]

	feesToAdd := feeAsset == market.GetBaseAsset()
	inAmount := satsAmount
	if feesToAdd {
		inAmount += preview.GetFeeAmount()
	}

	selectedUtxos, change, err := explorer.SelectUnspents(utxos, inAmount, asset)
	if err != nil {
		return nil, "", 0, err
	}

	unblindedIns := make([]*tdexv2.UnblindedInput, 0, len(selectedUtxos))
	ins := make([]psetv2.InputArgs, 0, len(selectedUtxos))
	for i, u := range selectedUtxos {
		ins = append(ins, psetv2.InputArgs{
			Txid:    u.Hash(),
			TxIndex: u.Index(),
		})
		unblindedIns = append(unblindedIns, &tdexv2.UnblindedInput{
			Index:         uint32(i),
			Asset:         u.Asset(),
			Amount:        u.Value(),
			AssetBlinder:  hex.EncodeToString(elementsutil.ReverseBytes(u.AssetBlinder())),
			AmountBlinder: hex.EncodeToString(elementsutil.ReverseBytes(u.ValueBlinder())),
		})
	}

	outAmount := preview.GetAmount()
	if !feesToAdd {
		outAmount -= preview.GetFeeAmount()
	}
	outs := []psetv2.OutputArgs{
		{
			Asset:        preview.GetAsset(),
			Amount:       outAmount,
			Script:       w.outputScript,
			BlindingKey:  w.blindPubkey.SerializeCompressed(),
			BlinderIndex: 0,
		},
	}
	if change > 0 {
		outs = append(outs, psetv2.OutputArgs{
			Asset:        asset,
			Amount:       change,
			Script:       w.outputScript,
			BlindingKey:  w.blindPubkey.SerializeCompressed(),
			BlinderIndex: 0,
		})
	}
	pset, _ := psetv2.New(ins, outs, nil)

	updater, _ := psetv2.NewUpdater(pset)
	for i, u := range selectedUtxos {
		_, prevout, _ := u.Parse()
		updater.AddInWitnessUtxo(i, prevout)
		updater.AddInUtxoRangeProof(i, u.RangeProof())
		updater.AddInSighashType(i, txscript.SigHashAll)
	}

	psetBase64, _ := pset.ToBase64()

	return &tdexv2.SwapRequest{
		Id:              uuid.NewString(),
		AmountP:         satsAmount,
		AssetP:          asset,
		AmountR:         preview.GetAmount(),
		AssetR:          preview.GetAsset(),
		Transaction:     psetBase64,
		UnblindedInputs: unblindedIns,
	}, preview.GetFeeAsset(), preview.GetFeeAmount(), nil
}

func makeTradeSellProposal(
	client tdexv2.TradeServiceClient,
	market *tdexv2.Market, swapRequest *tdexv2.SwapRequest,
	feeAsset string, feeAmount uint64,
) (*tdexv2.SwapAccept, error) {
	res, err := client.ProposeTrade(context.Background(), &tdexv2.ProposeTradeRequest{
		Market:      market,
		Type:        tdexv2.TradeType_TRADE_TYPE_SELL,
		SwapRequest: swapRequest,
		FeeAsset:    feeAsset,
		FeeAmount:   feeAmount,
	})
	if err != nil {
		return nil, err
	}
	if res.GetSwapFail() != nil {
		return nil, fmt.Errorf(res.GetSwapFail().GetFailureMessage())
	}
	return res.GetSwapAccept(), nil
}
