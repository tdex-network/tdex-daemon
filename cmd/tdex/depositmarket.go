package main

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
	"sort"
	"time"

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
			Name:  "no-fragment",
			Usage: "disable utxo fragmentation",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "explorer",
			Usage: "explorer endpoint url",
			Value: "http://127.0.0.1:3001",
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

	net, baseAssetKey := getNetworkAndBaseAssetKey(ctx.String("network"))
	explorerUrl := ctx.String("explorer")
	fragmentationDisabled := ctx.Bool("no-fragment")
	quoteAssetOpt := ctx.String("quote_asset")
	baseAssetOpt := ""
	if quoteAssetOpt != "" {
		baseAssetOpt = baseAssetKey
	}

	if fragmentationDisabled {
		resp, err := client.DepositMarket(
			context.Background(), &pboperator.DepositMarketRequest{
				Market: &pbtypes.Market{
					BaseAsset:  baseAssetOpt,
					QuoteAsset: quoteAssetOpt,
				},
				NumOfAddresses: 1,
			},
		)
		if err != nil {
			return err
		}

		printRespJSON(resp)

		return nil
	}

	randomWallet, err := trade.NewRandomWallet(&net)
	if err != nil {
		return err
	}

	log.Warnf("fund address: %v", randomWallet.Address())
	explorerSvc := explorer.NewService(explorerUrl)
	assetValuePair, unspents := findAssetsUnspents(
		randomWallet,
		explorerSvc,
		baseAssetKey,
	)

	log.Info("calculating fragments ...")
	baseFragments, quoteFragments := fragmentUnspents(assetValuePair)
	outsLen := len(baseFragments) + len(quoteFragments)

	log.Infof("base fragments: %v", baseFragments)
	log.Infof("quote fragments: %v", quoteFragments)

	feeAmount := wallet.EstimateTxSize(
		len(unspents),
		outsLen,
		false,
		MinMilliSatPerByte,
	)

	addresses, err := fetchMarketAddresses(
		outsLen,
		client,
		baseAssetOpt,
		quoteAssetOpt,
	)
	if err != nil {
		return err
	}

	outputs := createOutputs(
		baseFragments,
		quoteFragments,
		feeAmount,
		addresses,
		assetValuePair,
		baseAssetKey,
	)

	log.Info("crafting transaction ...")
	txHex, err := craftTransaction(
		randomWallet,
		unspents,
		outputs,
		feeAmount,
		net,
		baseAssetKey,
	)
	if err != nil {
		return err
	}

	log.Info("broadcasting transaction ...")

	for {
		resp, err := explorerSvc.BroadcastTransaction(txHex)
		if err != nil {
			log.Warn(err)
			log.Info("transaction broadcast retry")
			continue
		}
		log.Info(resp)
		break
	}

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

func fragmentUnspents(pair AssetValuePair) ([]uint64, []uint64) {

	baseAssetFragments := make([]uint64, 0)
	quoteAssetFragments := make([]uint64, 0)

	baseSum := uint64(0)
	quoteSum := uint64(0)
	for numOfUtxo, percentage := range fragmentationMap {
		for ; numOfUtxo > 0; numOfUtxo-- {
			baseAssetPart := percent(int(pair.BaseValue), percentage)
			baseSum += uint64(baseAssetPart)
			baseAssetFragments = append(baseAssetFragments, uint64(baseAssetPart))

			quoteAssetPart := percent(int(pair.QuoteValue), percentage)
			quoteSum += uint64(quoteAssetPart)
			quoteAssetFragments = append(quoteAssetFragments, uint64(quoteAssetPart))
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
) (
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
					if k == baseAssetKey {
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
					if k == baseAssetKey {
						if v < MinBaseDeposit {
							log.Warnf(
								"min base deposit is %v please top up",
								MinBaseDeposit,
							)
							continue events
						}
						assetValuePair.BaseAsset = k
						assetValuePair.BaseValue = v
						log.Infof(
							"base asset %v funded with value %v",
							k,
							v,
						)
					} else {
						if v < MinBaseDeposit {
							log.Warnf(
								"min quote deposit is %v, please top up",
								MinQuoteDeposit,
							)
							continue events
						}
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
		}

		time.Sleep(time.Duration(CrawlInterval) * time.Second)
	}

	return assetValuePair, unspents
}

func craftTransaction(
	randomWallet *trade.Wallet,
	unspents []explorer.Utxo,
	outs []TxOut,
	feeAmount uint64,
	network network.Network,
	baseAssetKey string,
) (string, error) {

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

func parseRequestOutputs(reqOutputs []TxOut, network network.Network) (
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
		script, blindingKey, err := parseConfidentialAddress(
			out.Address,
			network,
		)
		if err != nil {
			return nil, nil, err
		}

		output := transaction.NewTxOutput(asset, value, script)
		outputs = append(outputs, output)
		blindingKeys = append(blindingKeys, blindingKey)
	}
	return outputs, blindingKeys, nil
}

func parseConfidentialAddress(
	addr string,
	network network.Network,
) ([]byte, []byte, error) {
	script, err := address.ToOutputScript(addr, network)
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

func getNetworkAndBaseAssetKey(net string) (network.Network, string) {
	if net == network.Regtest.Name {
		return network.Regtest, network.Regtest.AssetID
	}
	return network.Liquid, network.Liquid.AssetID
}
