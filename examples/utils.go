package examples

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/psetv2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var lbtc = network.Regtest.AssetID

func initServices(
	daemonAddr, explorerAddr string,
) (tdexv2.TradeServiceClient, explorer.Service, error) {
	conn, err := grpc.Dial(
		daemonAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to daemon: %s", err)
	}
	client := tdexv2.NewTradeServiceClient(conn)

	explorer, err := esplora.NewService("http://localhost:3001", 15000)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare explorer: %s", err)
	}
	return client, explorer, nil
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

func faucet(
	explorer explorer.Service, w wallet, asset string, amount float64,
) ([]explorer.Utxo, error) {
	_, err := explorer.Faucet(w.address, amount, asset)
	if err != nil {
		return nil, err
	}
	time.Sleep(5 * time.Second)
	return explorer.GetUnspents(w.address, [][]byte{w.blindPrvkey.Serialize()})
}

func fetchMarkets(
	client tdexv2.TradeServiceClient,
) ([]*tdexv2.MarketWithFee, error) {
	res, err := client.ListMarkets(
		context.Background(), &tdexv2.ListMarketsRequest{},
	)
	if err != nil {
		return nil, err
	}
	return res.GetMarkets(), nil
}

func signAndCompleteTrade(
	client tdexv2.TradeServiceClient, w wallet, swap *tdexv2.SwapAccept,
) (string, error) {
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
