package main

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
	"sort"
	"time"

	"github.com/vulpemventures/go-elements/network"

	"github.com/urfave/cli/v2"
)

const (
	MinMilliSatPerByte = 100
	CrawlInterval      = 3
)

var depositmarket = cli.Command{
	Name:  "depositmarket",
	Usage: "get a deposit address for a given market or create a new one",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "base_asset",
			Usage: "the base asset hash of an existent market",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "quote_asset",
			Usage: "the base asset hash of an existent market",
			Value: "",
		},
	},
	Action: depositMarketAction,
}

type AssetValuePair struct {
	BaseValue  uint64
	BaseAsset  string
	QuoteValue uint64
	QuoteAsset string
}

var fragmentationMap = map[int]int{
	1: 30,
	2: 15,
	3: 10,
	5: 2,
}

func depositMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	//TODO change network
	randomWallet, err := trade.NewRandomWallet(&network.Regtest)
	if err != nil {
		return err
	}

	log.Warnf("fund address: %v", randomWallet.Address())

	//********
	//TODO update url using config
	explorerSvc := explorer.NewService("http://127.0.0.1:3001")
	log.Info("start crafting transaction")
	assetValuePair, unspents := findAssetsUnspents(randomWallet, explorerSvc)
	//********

	depositMarketResp, err := client.DepositMarket(
		context.Background(), &pboperator.DepositMarketRequest{
			Market: &pbtypes.Market{
				BaseAsset:  ctx.String("base_asset"),
				QuoteAsset: ctx.String("quote_asset"),
			},
		},
	)
	if err != nil {
		return err
	}
	log.Info(depositMarketResp)

	//********
	txHex, err := craftTransaction(
		randomWallet,
		assetValuePair,
		unspents,
		depositMarketResp.GetAddress(),
	)
	if err != nil {
		return err
	}
	log.Info(txHex)
	//********

	resp, err := explorerSvc.BroadcastTransaction(txHex)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}

//TODO test, rethink
func fragmentUnspents(pair AssetValuePair) ([]uint64, []uint64) {

	baseAssetFragments := make([]uint64, 0)
	quoteAssetFragments := make([]uint64, 0)

	for numOfUtxo, percentage := range fragmentationMap {
		for ; numOfUtxo > 0; numOfUtxo-- {
			baseAssetPart := percent(int(pair.BaseValue), percentage)
			baseAssetFragments = append(baseAssetFragments, uint64(baseAssetPart))

			quoteAssetPart := percent(int(pair.QuoteValue), percentage)
			quoteAssetFragments = append(quoteAssetFragments, uint64(quoteAssetPart))
		}
	}

	sort.Slice(baseAssetFragments, func(i, j int) bool {
		return baseAssetFragments[i] < baseAssetFragments[j]
	})

	sort.Slice(quoteAssetFragments, func(i, j int) bool {
		return quoteAssetFragments[i] < quoteAssetFragments[j]
	})

	return baseAssetFragments, quoteAssetFragments
}

func percent(num int, percent int) float64 {
	return (float64(num) * float64(percent)) / float64(100)
}

func findAssetsUnspents(randomWallet *trade.Wallet, explorerSvc explorer.Service) (
	AssetValuePair,
	[]explorer.Utxo,
) {

	var assetValuePair AssetValuePair
	var unspents []explorer.Utxo
	var err error

events:
	for {
		unspents, err = explorerSvc.GetUnspents(
			randomWallet.Address(),
			[][]byte{randomWallet.BlindingKey()},
		)
		if err != nil {
			log.Warn(err)
		}

		if len(unspents) > 0 {

			valuePerAsset := make(map[string]uint64, 0)
			for _, v := range unspents {
				valuePerAsset[v.Asset()] += v.Value()
			}

			switch len(valuePerAsset) {
			case 1:
				for k, v := range valuePerAsset {
					if k == config.GetString(config.BaseAssetKey) {
						log.Warnf(
							"only base asset %v funded with value %v",
							k,
							v,
						)
					} else {
						log.Warnf(
							"only quote asset %v funded with value %v",
							k,
							v,
						)
					}
				}
			case 2:
				for k, v := range valuePerAsset {
					if k == config.GetString(config.BaseAssetKey) {
						assetValuePair.BaseAsset = k
						assetValuePair.BaseValue = v
						log.Infof(
							"base asset %v funded with value %v",
							k,
							v,
						)
					} else {
						assetValuePair.QuoteAsset = k
						assetValuePair.QuoteValue = v
						log.Infof(
							"quote asset %v funded with value %v",
							k,
							v,
						)
					}
				}
				break events
			}
		} else {
			log.Warnf("no funds detected for address %v", randomWallet.Address())
		}

		time.Sleep(time.Duration(CrawlInterval) * time.Second)
	}

	return assetValuePair, unspents
}

func craftTransaction(
	randomWallet *trade.Wallet,
	assetValuePair AssetValuePair,
	unspents []explorer.Utxo,
	outAddress string,
) (string, error) {
	baseFragments, quoteFragments := fragmentUnspents(assetValuePair)
	outsLen := len(baseFragments) + len(quoteFragments)

	feeAmount := wallet.EstimateTxSize(
		len(unspents),
		outsLen,
		false,
		MinMilliSatPerByte,
	)

	outs := make([]TxOut, 0, outsLen)
	for i, v := range baseFragments {
		value := int64(v)
		if i == len(baseFragments)-1 {
			value = int64(v) - int64(feeAmount)
		}
		outs = append(outs, TxOut{
			Asset:   config.GetString(config.BaseAssetKey),
			Value:   value,
			Address: outAddress,
		})
	}
	for _, v := range quoteFragments {
		outs = append(outs, TxOut{
			Asset:   assetValuePair.QuoteAsset,
			Value:   int64(v),
			Address: outAddress,
		})
	}

	outputs, outputsBlindingKeys, err := parseRequestOutputs(outs)
	if err != nil {
		return "", err
	}

	ptx, err := pset.New(
		make([]*transaction.TxInput, 0, len(unspents)),
		make([]*transaction.TxOutput, 0, len(outputs)),
		2,
		0,
	)
	if err != nil {
		return "", err
	}

	ptx, err = addInsAndOutsToPset(ptx, unspents, outputs)
	if err != nil {
		return "", err
	}

	inputBlindingKeys := [][]byte{
		randomWallet.BlindingKey(),
		randomWallet.BlindingKey(),
	}
	blinder, err := pset.NewBlinder(
		ptx,
		inputBlindingKeys,
		outputsBlindingKeys,
		nil,
		nil,
	)
	if err != nil {
		return "", err
	}

	err = blinder.Blind()
	if err != nil {
		return "", err
	}

	updater, err := pset.NewUpdater(ptx)
	if err != nil {
		return "", err
	}

	feeValue, _ := confidential.SatoshiToElementsValue(feeAmount)
	lbtc, err := bufferutil.AssetHashToBytes(
		config.GetString(config.BaseAssetKey),
	)
	if err != nil {
		return "", err
	}
	feeOutput := transaction.NewTxOutput(lbtc, feeValue[:], []byte{})
	updater.AddOutput(feeOutput)

	ptxBase64, err := ptx.ToBase64()
	if err != nil {
		return "", err
	}

	signedPtxBase64, err := randomWallet.Sign(ptxBase64)
	if err != nil {
		return "", err
	}

	signedPtx, err := pset.NewPsetFromBase64(signedPtxBase64)
	if err != nil {
		return "", err
	}

	valid, err := signedPtx.ValidateAllSignatures()
	if err != nil {
		return "", err
	}

	if !valid {
		return "", errors.New("invalid signatures")
	}

	err = pset.FinalizeAll(signedPtx)
	if err != nil {
		return "", err
	}

	finalTx, err := pset.Extract(signedPtx)
	if err != nil {
		return "", err
	}

	txHex, err := finalTx.ToHex()
	if err != nil {
		return "", err
	}

	return txHex, nil
}

type TxOut struct {
	Asset   string
	Value   int64
	Address string
}

func parseRequestOutputs(reqOutputs []TxOut) (
	[]*transaction.TxOutput,
	[][]byte,
	error,
) {
	outputs := make([]*transaction.TxOutput, 0, len(reqOutputs))
	blindingKeys := make([][]byte, 0, len(reqOutputs))

	for _, out := range reqOutputs {
		asset, err := bufferutil.AssetHashToBytes(out.Asset)
		if err != nil {
			return nil, nil, err
		}
		value, err := bufferutil.ValueToBytes(uint64(out.Value))
		if err != nil {
			return nil, nil, err
		}
		script, blindingKey, err := parseConfidentialAddress(out.Address)
		if err != nil {
			return nil, nil, err
		}

		output := transaction.NewTxOutput(asset, value, script)
		outputs = append(outputs, output)
		blindingKeys = append(blindingKeys, blindingKey)
	}
	return outputs, blindingKeys, nil
}

func parseConfidentialAddress(addr string) ([]byte, []byte, error) {
	script, err := address.ToOutputScript(addr, *config.GetNetwork())
	if err != nil {
		return nil, nil, err
	}
	ctAddr, err := address.FromConfidential(addr)
	if err != nil {
		return nil, nil, err
	}
	return script, ctAddr.BlindingKey, nil
}

func addInsAndOutsToPset(
	ptx *pset.Pset,
	inputsToAdd []explorer.Utxo,
	outputsToAdd []*transaction.TxOutput,
) (*pset.Pset, error) {
	updater, err := pset.NewUpdater(ptx)
	if err != nil {
		return nil, err
	}

	for _, in := range inputsToAdd {
		input, witnessUtxo, _ := in.Parse()
		updater.AddInput(input)
		err := updater.AddInWitnessUtxo(witnessUtxo, len(ptx.Inputs)-1)
		if err != nil {
			return nil, err
		}
	}

	for _, out := range outputsToAdd {
		updater.AddOutput(out)
	}

	return ptx, nil
}
