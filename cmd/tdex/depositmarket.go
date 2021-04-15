package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/tdex-network/tdex-protobuf/generated/go/types"
	"github.com/vulpemventures/go-elements/elementsutil"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"

	"github.com/urfave/cli/v2"
)

const (
	MinMilliSatPerByte = 150
	CrawlInterval      = 3
	MinBaseDeposit     = 50000
	MinQuoteDeposit    = 50000
)

var depositmarket = cli.Command{
	Name:  "depositmarket",
	Usage: "get a deposit address for a given market or create a new one",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "quote_asset",
			Usage: "the quote asset hash of an existent market",
			Value: "",
		},
		&cli.BoolFlag{
			Name:  "no_fragments",
			Usage: "disable utxo fragmentation",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "explorer",
			Usage: "explorer endpoint url",
			Value: "http://127.0.0.1:3001",
		},
		&cli.IntFlag{
			Name:  "num_of_addresses",
			Usage: "the number of addresses to generate for the market",
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

var fragmentationMapConfig = map[int]int{
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

	net, err := getNetworkFromState()
	if err != nil {
		return err
	}
	baseAssetKey := net.AssetID
	fragmentationDisabled := ctx.Bool("no-fragment")
	quoteAssetOpt := ctx.String("quote_asset")
	baseAssetOpt := ""
	if quoteAssetOpt != "" {
		baseAssetOpt = baseAssetKey
	}

	if fragmentationDisabled {
		numOfAddresses := ctx.Int64("num_of_addresses")
		resp, err := client.DepositMarket(
			context.Background(), &pboperator.DepositMarketRequest{
				Market: &pbtypes.Market{
					BaseAsset:  baseAssetOpt,
					QuoteAsset: quoteAssetOpt,
				},
				NumOfAddresses: numOfAddresses,
			},
		)
		if err != nil {
			return err
		}

		printRespJSON(resp)
		return nil
	}

	explorerSvc, err := getExplorerFromState()
	if err != nil {
		return fmt.Errorf("error while setting up explorer service: %v", err)
	}

	randomWallet, err := trade.NewRandomWallet(net)
	if err != nil {
		return err
	}
	log.Info("send funds to address: ", randomWallet.Address())

	funds := waitForOperatorFunds()

	assetValuePair, unspents, err := findAssetsUnspents(
		randomWallet,
		explorerSvc,
		baseAssetKey,
		funds,
	)
	if err != nil {
		return err
	}

	log.Info("calculating fragments...")
	baseFragments, quoteFragments := fragmentUnspents(
		assetValuePair,
		fragmentationMapConfig,
	)
	numUnspents := len(unspents)
	numFragments := len(baseFragments) + len(quoteFragments)
	log.Infof(
		"fetched %d funds that will be split into %d fragments",
		numUnspents,
		numFragments,
	)
	log.Infof(
		"base asset amount %d will be split into %d fragments",
		assetValuePair.BaseValue,
		len(baseFragments),
	)
	log.Infof(
		"quote asset amount %d will be split into %d fragments",
		assetValuePair.QuoteValue,
		len(quoteFragments),
	)

	feeAmount := estimateFees(numUnspents, numFragments)

	addresses, err := fetchMarketAddresses(
		numFragments,
		client,
		baseAssetOpt,
		quoteAssetOpt,
	)
	if err != nil {
		return err
	}

	log.Info("crafting transaction...")
	txHex, err := craftTransaction(
		randomWallet,
		unspents,
		baseFragments, quoteFragments,
		addresses,
		feeAmount,
		net,
		assetValuePair,
	)
	if err != nil {
		return err
	}

	log.Info("sending transaction...")
	txID, err := explorerSvc.BroadcastTransaction(txHex)
	if err != nil {
		return fmt.Errorf("failed to braodcast tx: %v", err)
	}

	log.Infof("market account funding txid: %s", txID)

	log.Info("waiting for tx to get confirmed...")
	if err := waitUntilTxConfirmed(explorerSvc, txID); err != nil {
		return err
	}

	log.Info("claiming market deposits...")
	outpoints := createOutpoints(txID, numFragments)
	if _, err := client.ClaimMarketDeposit(
		context.Background(), &pboperator.ClaimMarketDepositRequest{
			Market: &types.Market{
				BaseAsset:  assetValuePair.BaseAsset,
				QuoteAsset: assetValuePair.QuoteAsset,
			},
			Outpoints: outpoints,
		},
	); err != nil {
		return err
	}

	log.Info("done")
	return nil
}

func fetchMarketAddresses(
	outsLen int,
	client pboperator.OperatorClient,
	baseAssetOpt string,
	quoteAssetOpt string,
) ([]string, error) {
	depositMarket, err := client.DepositMarket(
		context.Background(), &pboperator.DepositMarketRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAssetOpt,
				QuoteAsset: quoteAssetOpt,
			},
			NumOfAddresses: int64(outsLen),
		},
	)
	if err != nil {
		return nil, err
	}

	return depositMarket.GetAddresses(), nil
}

func createOutputs(
	baseFragments, quoteFragments []uint64,
	feeAmount uint64,
	addresses []string,
	assetValuePair AssetValuePair,
	baseAssetKey string,
) []TxOut {
	outsLen := len(baseFragments) + len(quoteFragments)
	outputs := make([]TxOut, 0, outsLen)

	index := 0
	for i, v := range baseFragments {
		value := int64(v)
		//deduct fee from last(largest) fragment
		if i == len(baseFragments)-1 {
			value = int64(v) - int64(feeAmount)
		}
		outputs = append(outputs, TxOut{
			Asset:   baseAssetKey,
			Value:   value,
			Address: addresses[index],
		})
		index++
	}
	for _, v := range quoteFragments {
		outputs = append(outputs, TxOut{
			Asset:   assetValuePair.QuoteAsset,
			Value:   int64(v),
			Address: addresses[index],
		})
		index++
	}

	return outputs
}

func fragmentUnspents(pair AssetValuePair, fragmentationMap map[int]int) ([]uint64, []uint64) {
	baseAssetFragments := make([]uint64, 0)
	quoteAssetFragments := make([]uint64, 0)

	baseSum := uint64(0)
	quoteSum := uint64(0)
	for numOfUtxo, percentage := range fragmentationMap {
		for ; numOfUtxo > 0; numOfUtxo-- {
			if pair.BaseValue > 0 {
				baseAssetPart := percent(int(pair.BaseValue), percentage)
				baseSum += uint64(baseAssetPart)
				baseAssetFragments = append(baseAssetFragments, uint64(baseAssetPart))
			}

			if pair.QuoteValue > 0 {
				quoteAssetPart := percent(int(pair.QuoteValue), percentage)
				quoteSum += uint64(quoteAssetPart)
				quoteAssetFragments = append(quoteAssetFragments, uint64(quoteAssetPart))
			}
		}
	}

	sort.Slice(baseAssetFragments, func(i, j int) bool {
		return baseAssetFragments[i] < baseAssetFragments[j]
	})

	sort.Slice(quoteAssetFragments, func(i, j int) bool {
		return quoteAssetFragments[i] < quoteAssetFragments[j]
	})

	//if there is rest, created when calculating percentage,
	//add it to last fragment
	if baseSum != pair.BaseValue {
		baseRest := pair.BaseValue - baseSum
		if baseRest > 0 {
			baseAssetFragments[len(baseAssetFragments)-1] =
				baseAssetFragments[len(baseAssetFragments)-1] + baseRest
		}
	}

	//if there is rest, created when calculating percentage,
	//add it to last fragment
	if quoteSum != pair.QuoteValue {
		quoteRest := pair.QuoteValue - quoteSum
		if quoteRest > 0 {
			quoteAssetFragments[len(quoteAssetFragments)-1] =
				quoteAssetFragments[len(quoteAssetFragments)-1] + quoteRest
		}
	}

	return baseAssetFragments, quoteAssetFragments
}

func percent(num int, percent int) float64 {
	return (float64(num) * float64(percent)) / float64(100)
}

func findAssetsUnspents(
	randomWallet *trade.Wallet,
	explorerSvc explorer.Service,
	baseAssetKey string,
	txids []string,
) (AssetValuePair, []explorer.Utxo, error) {
	var assetValuePair AssetValuePair
	var unspents []explorer.Utxo
	valuePerAsset := make(map[string]uint64, 0)

	for _, txid := range txids {
		u, err := getUnspents(explorerSvc, randomWallet, txid)
		if err != nil {
			return assetValuePair, nil, err
		}

		if len(u) > 0 {
			for _, v := range u {
				valuePerAsset[v.Asset()] += v.Value()
			}
			unspents = append(unspents, u...)
		}
	}

	for k, v := range valuePerAsset {
		if k == baseAssetKey {
			assetValuePair.BaseAsset = k
			assetValuePair.BaseValue = v
		} else {
			if assetValuePair.QuoteAsset == "" {
				assetValuePair.QuoteAsset = k
				assetValuePair.QuoteValue = v
			} else {
				if k != assetValuePair.QuoteAsset {
					log.Warnf("congrats! You just lost %d of asset %s 🎉", k, v)
				}
			}
		}
	}

	if assetValuePair.BaseValue == 0 {
		log.Warnf("base asset not funded, you'll need to make another depositmarket operation")
	} else if assetValuePair.BaseValue < MinBaseDeposit {
		log.Warnf(
			"min base deposit is %v, you'll need to topup another depositmarket operation",
			MinBaseDeposit,
		)
	}
	if assetValuePair.QuoteValue == 0 {
		log.Warn("quote asset not funded, you'll need to make another depositmarket operation")
	} else if assetValuePair.QuoteValue < MinQuoteDeposit {
		log.Warnf(
			"min quote deposit is %v, you'll need to topup another depositmarket operation",
			MinQuoteDeposit,
		)
	}

	return assetValuePair, unspents, nil
}

func craftTransaction(
	randomWallet *trade.Wallet,
	unspents []explorer.Utxo,
	baseFragments, quoteFragments []uint64,
	addresses []string,
	feeAmount uint64,
	network *network.Network,
	assetValuePair AssetValuePair,
) (string, error) {
	baseAssetKey := assetValuePair.BaseAsset
	outs := createOutputs(
		baseFragments,
		quoteFragments,
		feeAmount,
		addresses,
		assetValuePair,
		baseAssetKey,
	)

	outputs, outputsBlindingKeys, err := parseRequestOutputs(outs, network)
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

	dataLen := len(unspents)
	inBlindData := make([]pset.BlindingDataLike, dataLen, dataLen)
	for i, u := range unspents {
		asset, _ := hex.DecodeString(u.Asset())
		inBlindData[i] = pset.BlindingData{
			Value:               u.Value(),
			Asset:               elementsutil.ReverseBytes(asset),
			ValueBlindingFactor: u.ValueBlinder(),
			AssetBlindingFactor: u.AssetBlinder(),
		}
	}

	blinder, err := pset.NewBlinder(
		ptx,
		inBlindData,
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

	feeValue, _ := elementsutil.SatoshiToElementsValue(feeAmount)
	lbtc, err := bufferutil.AssetHashToBytes(baseAssetKey)
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

func parseRequestOutputs(reqOutputs []TxOut, network *network.Network) (
	[]*transaction.TxOutput,
	map[int][]byte,
	error,
) {
	outLen := len(reqOutputs)
	outputs := make([]*transaction.TxOutput, outLen, outLen)
	blindingKeys := make(map[int][]byte)

	for i, out := range reqOutputs {
		asset, err := bufferutil.AssetHashToBytes(out.Asset)
		if err != nil {
			return nil, nil, err
		}
		value, err := bufferutil.ValueToBytes(uint64(out.Value))
		if err != nil {
			return nil, nil, err
		}
		script, blindingKey, err := parseConfidentialAddress(
			out.Address,
			network,
		)
		if err != nil {
			return nil, nil, err
		}

		output := transaction.NewTxOutput(asset, value, script)
		outputs[i] = output
		blindingKeys[i] = blindingKey
	}
	return outputs, blindingKeys, nil
}

func parseConfidentialAddress(
	addr string,
	network *network.Network,
) ([]byte, []byte, error) {
	script, err := address.ToOutputScript(addr)
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
