package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
)

const (
	MinMilliSatPerByte = 150
	CrawlInterval      = 3
	MinBaseDeposit     = 50000
	MinQuoteDeposit    = 50000
)

var fragmentmarket = cli.Command{
	Name: "fragmentmarket",
	Usage: "deposit funds for a market (either existing or to be created) " +
		"into an ephemeral wallet, then split the amount into multiple " +
		"fragments and deposit into the daemon",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "base_asset",
			Usage: "the base asset hash of an existent market",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "quote_asset",
			Usage: "the quote asset hash of an existent market",
			Value: "",
		},
		&cli.StringSliceFlag{
			Name:  "txid",
			Usage: "txid of the funds to resume a fragmentmarket",
		},
	},
	Action: fragmentMarketAction,
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

func fragmentMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	net, err := getNetworkFromState()
	if err != nil {
		return err
	}

	walletType := "market"
	txids := ctx.StringSlice("txid")
	quoteAssetOpt := ctx.String("quote_asset")
	baseAssetOpt := ctx.String("base_asset")
	if baseAssetOpt == "" {
		baseAssetOpt = net.AssetID
	}

	explorerSvc, err := getExplorerFromState()
	if err != nil {
		return fmt.Errorf("error while setting up explorer service: %v", err)
	}

	walletKeys, err := getWalletFromState(walletType)
	if err != nil {
		return err
	}
	if walletKeys == nil && txids != nil {
		log.Info("no ephemeral wallet detected, skipping resume of provided txids")
	}
	if walletKeys != nil && txids == nil {
		return fmt.Errorf("expected to resume previous fragmentation but no txids were provided. Please retry by specifying txids")
	}

	ephWallet, err := getEphemeralWallet(walletType, walletKeys, net)
	if err != nil {
		return err
	}

	if len(txids) > 0 && walletKeys != nil {
		log.Info("resuming previous fragmentation")
		log.Infof("you can optionally send other funds to address: %s", ephWallet.Address())
	}
	log.Info("send funds to address: ", ephWallet.Address())

	var assetValuePair AssetValuePair
	var unspents []explorer.Utxo
	for {
		funds := waitForOperatorFunds()
		funds = append(funds, txids...)

		assetValuePair, unspents, err = findAssetsUnspents(
			ephWallet,
			explorerSvc,
			baseAssetOpt,
			funds,
		)
		if err != nil {
			log.WithError(err).Warn("an unexpected error occured, please retry entering all txids")
			continue
		}
		break
	}

	log.Info("calculating fragments...")
	baseFragments, quoteFragments := fragmentUnspents(
		assetValuePair,
		fragmentationMapConfig,
	)
	feeAmount := estimateFees(len(unspents), len(baseFragments)+len(quoteFragments))
	baseFragments = deductFeeFromFragments(baseFragments, feeAmount)

	numUnspents := len(unspents)
	numFragments := len(baseFragments) + len(quoteFragments)
	log.Infof(
		"detected %d coins that will be split into %d fragments",
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

	if quoteAssetOpt == "" {
		baseAssetOpt = ""
	}

	addresses, err := getMarketDepositAddresses(
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
		ephWallet,
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
			Market: &pbtypes.Market{
				BaseAsset:  assetValuePair.BaseAsset,
				QuoteAsset: assetValuePair.QuoteAsset,
			},
			Outpoints: outpoints,
		},
	); err != nil {
		return err
	}

	flushWallet(walletType)
	log.Info("done")
	return nil
}

func getMarketDepositAddresses(
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
) []TxOut {
	outsLen := len(baseFragments) + len(quoteFragments)
	outputs := make([]TxOut, 0, outsLen)

	index := 0
	for _, v := range baseFragments {
		outputs = append(outputs, TxOut{
			Asset:   assetValuePair.BaseAsset,
			Value:   int64(v),
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
	valuePerAsset := make(map[string]uint64)

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
					log.Warnf("congrats! You just lost %d of asset %s 🎉", v, k)
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
	inBlindData := make([]pset.BlindingDataLike, 0, dataLen)
	for _, u := range unspents {
		asset, _ := hex.DecodeString(u.Asset())
		inBlindData = append(inBlindData, pset.BlindingData{
			Value:               u.Value(),
			Asset:               elementsutil.ReverseBytes(asset),
			ValueBlindingFactor: u.ValueBlinder(),
			AssetBlindingFactor: u.AssetBlinder(),
		})
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
	outputs := make([]*transaction.TxOutput, 0, outLen)
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
		outputs = append(outputs, output)
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

func getEphemeralWallet(
	walletType string,
	walletKeys map[string]string,
	net *network.Network,
) (*trade.Wallet, error) {
	if walletKeys != nil {
		privKey, _ := hex.DecodeString(walletKeys["privateKey"])
		blindKey, _ := hex.DecodeString(walletKeys["blindingKey"])
		return trade.NewWalletFromKey(privKey, blindKey, net), nil
	}

	w, err := trade.NewRandomWallet(net)
	if err != nil {
		return nil, err
	}
	wallet := fmt.Sprintf(
		`{"privateKey": "%s", "blindingKey": "%s"}`,
		hex.EncodeToString(w.PrivateKey()),
		hex.EncodeToString(w.BlindingKey()),
	)
	walletKey := fmt.Sprintf("%s_wallet", walletType)
	if err := setState(map[string]string{walletKey: wallet}); err != nil {
		return nil, err
	}
	return w, nil
}

func deductFeeFromFragments(fragments []uint64, feeAmount uint64) []uint64 {
	f := make([]uint64, len(fragments))
	copy(f, fragments)

	amountToPay := int64(feeAmount)
	for amountToPay > 0 {
		fLen := len(f) - 1
		lastFragment := int64(f[fLen])
		if amountToPay >= lastFragment {
			f = f[:fLen]
		} else {
			f[fLen] -= uint64(amountToPay)
		}
		amountToPay -= lastFragment
	}
	return f
}
