package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/txscript"
	"github.com/google/uuid"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/psetv2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	client      tdexv2.TradeServiceClient
	explorerSvc explorer.Service

	lbtc         = network.Regtest.AssetID
	traderAmount = 10.0
)

func initEnv() error {
	conn, err := grpc.Dial("localhost:9945", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %s", err)
	}
	client = tdexv2.NewTradeServiceClient(conn)

	explorerSvc, err = esplora.NewService("http://localhost:3001", 15000)
	if err != nil {
		return fmt.Errorf("failed to prepare explorer: %s", err)
	}
	return nil
}

func main() {
	if err := initEnv(); err != nil {
		log.Fatal(err)
	}

	traderWallet := newKeyPair()
	fmt.Printf("Trader wallet:\n%s\n\n", traderWallet)

	fmt.Printf("Fetching market from provider...\n\n")
	markets, err := fetchMarkets()
	if err != nil {
		log.Fatalf("failed to fetch provider markets: %s", err)
	}

	if len(markets) == 0 {
		log.Fatal("provider has no tradable markets")
	}

	targetMarket := markets[0].GetMarket()
	fmt.Printf("Market: %s\n\n", targetMarket)

	usdt := targetMarket.GetQuoteAsset()

	fmt.Printf("Sending USDt funds to trader...\n\n")
	utxos, err := faucet(traderWallet, usdt, traderAmount)
	if err != nil {
		log.Fatalf("failed to send USDt funds to trader: %s", err)
	}

	fmt.Println(utxos[0].Hash())

	fmt.Printf("Making trade preview...\n\n")
	swapRequest, err := makeTradePreview(
		targetMarket, usdt, lbtc, traderAmount, traderWallet, utxos,
	)
	if err != nil {
		log.Fatalf("failed to make trade preview: %s", err)
	}

	fmt.Printf("Making trade proposal...\n\n")
	swapAccept, err := makeTradeProposal(targetMarket, swapRequest)
	if err != nil {
		log.Fatalf("failed to make trade proposal: %s", err)
	}

	fmt.Printf("Signing and completing trade...\n\n")
	txid, err := signAndCompleteTrade(swapAccept, traderWallet)
	if err != nil {
		log.Fatalf("failed to complete trade: %s", err)
	}

	fmt.Printf("Completed swap: %s\n\n", txid)
	swapRequest.Transaction = swapAccept.GetTransaction()
	swapRequest.UnblindedInputs = swapAccept.GetUnblindedInputs()
	fmt.Printf("Swap info: %s\n", swapRequest)
}

func newKeyPair() wallet {
	w := &wallet{}
	for {
		key, err := btcec.NewPrivateKey()
		if err != nil {
			continue
		}
		w.prvkey = key
		w.pubkey = key.PubKey()
		break
	}
	for {
		key, err := btcec.NewPrivateKey()
		if err != nil {
			continue
		}
		w.blindPrvkey = key
		w.blindPubkey = key.PubKey()
		break
	}
	p2wpkh := payment.FromPublicKey(w.pubkey, &network.Regtest, w.blindPubkey)
	w.address, _ = p2wpkh.ConfidentialWitnessPubKeyHash()
	w.signingScript = p2wpkh.Script
	w.outputScript = p2wpkh.WitnessScript
	return *w
}

func faucet(w wallet, asset string, amount float64) ([]explorer.Utxo, error) {
	_, err := explorerSvc.Faucet(w.address, amount, asset)
	if err != nil {
		return nil, err
	}
	time.Sleep(5 * time.Second)
	return explorerSvc.GetUnspents(w.address, [][]byte{w.blindPrvkey.Serialize()})
}

func fetchMarkets() ([]*tdexv2.MarketWithFee, error) {
	res, err := client.ListMarkets(
		context.Background(), &tdexv2.ListMarketsRequest{},
	)
	if err != nil {
		return nil, err
	}
	return res.GetMarkets(), nil
}

func makeTradePreview(
	market *tdexv2.Market, asset, feeAsset string, amount float64,
	w wallet, utxos []explorer.Utxo,
) (*tdexv2.SwapRequest, error) {
	satsAmount := uint64(amount * math.Pow10(8))
	res, err := client.PreviewTrade(context.Background(), &tdexv2.PreviewTradeRequest{
		Market:   market,
		Type:     tdexv2.TradeType_TRADE_TYPE_BUY,
		Amount:   satsAmount,
		Asset:    asset,
		FeeAsset: feeAsset,
	})
	if err != nil {
		return nil, err
	}
	preview := res.GetPreviews()[0]

	feesToAdd := feeAsset == market.GetQuoteAsset()
	inAmount := satsAmount
	if feesToAdd {
		inAmount += preview.GetFeeAmount()
	}

	selectedUtxos, change, err := explorer.SelectUnspents(utxos, inAmount, asset)
	if err != nil {
		return nil, err
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
		FeeAmount:       preview.GetFeeAmount(),
		FeeAsset:        preview.GetFeeAsset(),
	}, nil
}

func makeTradeProposal(
	market *tdexv2.Market, swapRequest *tdexv2.SwapRequest,
) (*tdexv2.SwapAccept, error) {
	res, err := client.ProposeTrade(context.Background(), &tdexv2.ProposeTradeRequest{
		Market:      market,
		Type:        tdexv2.TradeType_TRADE_TYPE_BUY,
		SwapRequest: swapRequest,
	})
	if err != nil {
		return nil, err
	}
	if res.GetSwapFail() != nil {
		return nil, fmt.Errorf(res.GetSwapFail().GetFailureMessage())
	}
	return res.GetSwapAccept(), nil
}

func signAndCompleteTrade(swap *tdexv2.SwapAccept, w wallet) (string, error) {
	pset, err := psetv2.NewPsetFromBase64(swap.GetTransaction())
	if err != nil {
		return "", err
	}
	signer, err := psetv2.NewSigner(pset)
	if err != nil {
		return "", err
	}
	tx, _ := pset.UnsignedTx()
	sighash := tx.HashForWitnessV0(0, w.signingScript, pset.Inputs[0].GetUtxo().Value, pset.Inputs[0].SigHashType)
	sig := ecdsa.Sign(w.prvkey, sighash[:])
	sigWithHashType := append(sig.Serialize(), byte(pset.Inputs[0].SigHashType))
	if err := signer.SignInput(
		0, sigWithHashType, w.pubkey.SerializeCompressed(), nil, nil,
	); err != nil {
		return "", err
	}

	completedPset, _ := pset.ToBase64()

	res, err := client.CompleteTrade(context.Background(), &tdexv2.CompleteTradeRequest{
		SwapComplete: &tdexv2.SwapComplete{
			AcceptId:    swap.GetId(),
			Transaction: completedPset,
		},
	})
	if err != nil {
		return "", err
	}
	if res.GetSwapFail() != nil {
		return "", fmt.Errorf(res.GetSwapFail().FailureMessage)
	}

	return res.GetTxid(), nil
}

type wallet struct {
	prvkey        *btcec.PrivateKey
	pubkey        *btcec.PublicKey
	blindPrvkey   *btcec.PrivateKey
	blindPubkey   *btcec.PublicKey
	address       string
	outputScript  []byte
	signingScript []byte
}

func (w wallet) String() string {
	return fmt.Sprintf(
		"Keypair: %s %s\nAddress: %s",
		hex.EncodeToString(w.prvkey.Serialize()),
		hex.EncodeToString(w.pubkey.SerializeCompressed()),
		w.address,
	)
}
