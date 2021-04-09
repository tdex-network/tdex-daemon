package main

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"

	"github.com/urfave/cli/v2"
)

const (
	MinFee          = 5000
	MaxNumOfOutputs = 150
)

var depositfee = cli.Command{
	Name:   "depositfee",
	Usage:  "get a deposit address for the fee account used to subsidize liquid network fees",
	Action: depositFeeAction,
	Flags: []cli.Flag{
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
		&cli.IntFlag{
			Name:  "num_of_addresses",
			Usage: "the number of addresses to retrieve",
		},
	},
}

func depositFeeAction(ctx *cli.Context) error {
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

	if fragmentationDisabled {
		numOfAddresses := ctx.Int64("num_of_addresses")
		resp, err := client.DepositFeeAccount(
			context.Background(), &pboperator.DepositFeeAccountRequest{
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
		log.WithError(err).Panic("error while setting up explorer service")
	}

	randomWallet, err := trade.NewRandomWallet(net)
	if err != nil {
		return err
	}

	log.Info("fund address: ", randomWallet.Address())

	baseAssetValue, unspents := findBaseAssetsUnspents(
		randomWallet,
		explorerSvc,
		baseAssetKey,
	)

	log.Info("calculating fragments ...")
	baseFragments := fragmentFeeUnspents(
		baseAssetValue,
		MinFee,
		MaxNumOfOutputs,
	)
	log.Infof(
		"fetched %d funds that will be split into %d fragments",
		len(unspents),
		len(baseFragments),
	)

	addresses, err := fetchFeeAccountAddresses(
		len(baseFragments),
		client,
	)
	if err != nil {
		log.Error(err)
	}

	inLen := len(unspents)
	ins := make([]int, inLen, inLen)
	for i := range unspents {
		ins[i] = wallet.P2WPKH
	}
	outLen := len(baseFragments)
	outs := make([]int, outLen, outLen)
	for i := range baseFragments {
		outs[i] = wallet.P2WPKH
	}
	estimatedSize := wallet.EstimateTxSize(ins, nil, nil, outs, nil)
	feeAmount := uint64(float64(estimatedSize) * 0.1)

	outputs := createOutputsForDepositFeeTransaction(
		baseFragments,
		feeAmount,
		addresses,
		baseAssetKey,
	)

	txHex, err := craftTransaction(
		randomWallet,
		unspents,
		outputs,
		feeAmount,
		net,
		baseAssetKey,
		[][]byte{randomWallet.BlindingKey()},
	)
	if err != nil {
		log.Error(err)
	}

	var txID string
retry:
	for {
		resp, err := explorerSvc.BroadcastTransaction(txHex)
		if err != nil {
			log.Warn(err)
			log.Info("transaction broadcast retry")
			continue retry
		}
		log.Info(resp)
		txID = resp
		break retry
	}

	outpoints := make([]*pboperator.TxOutpoint, 0, len(outputs))
	for i := 0; i < len(outputs); i++ {
		outpoints = append(outpoints, &pboperator.TxOutpoint{
			Hash:  txID,
			Index: int32(i),
		})
	}

	//wait so that tx get confirmed
	time.Sleep(65 * time.Second)

	_, err = client.ClaimFeeDeposit(
		context.Background(), &pboperator.ClaimFeeDepositRequest{
			Outpoints: outpoints,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// fragmentFeeUnspents returns slice of minFragmentValue's up to maxNumOfFragments
// maxNumOfFragments'th element of slice could be greater than minFragmentValue
// method is used to fragment fee account unspent to more fragments that are
// going to be used for paying transaction fee
func fragmentFeeUnspents(
	valueToBeFragmented uint64,
	minFragmentValue uint64,
	maxNumOfFragments int,
) []uint64 {
	res := make([]uint64, 0)
	for i := 0; valueToBeFragmented >= minFragmentValue && i < maxNumOfFragments; i++ {
		res = append(res, minFragmentValue)
		valueToBeFragmented -= minFragmentValue
	}
	if valueToBeFragmented > 0 {
		res[len(res)-1] += valueToBeFragmented
	}

	return res
}

func createOutputsForDepositFeeTransaction(
	baseFragments []uint64,
	feeAmount uint64,
	addresses []string,
	baseAssetKey string,
) []TxOut {
	outsLen := len(baseFragments)
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

	return outputs
}

// findBaseAssetsUnspents polls blockchain until base asset unspent is noticed
func findBaseAssetsUnspents(
	randomWallet *trade.Wallet,
	explorerSvc explorer.Service,
	baseAssetKey string,
) (
	uint64,
	[]explorer.Utxo,
) {

	var unspents []explorer.Utxo
	var err error
	valuePerAsset := make(map[string]uint64, 0)

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

			for _, v := range unspents {
				valuePerAsset[v.Asset()] += v.Value()
			}

			if baseTotalAmount, ok := valuePerAsset[baseAssetKey]; ok {
				if baseTotalAmount < MinFee {
					log.Warnf(
						"min base deposit is %v please top up",
						MinFee,
					)
					continue events
				}
				log.Infof(
					"base asset %v funded with value %v",
					baseAssetKey,
					baseTotalAmount,
				)
				break events
			}

		}

		time.Sleep(time.Duration(CrawlInterval) * time.Second)
	}

	return valuePerAsset[baseAssetKey], unspents
}

func fetchFeeAccountAddresses(
	numOfAddresses int,
	client pboperator.OperatorClient,
) ([]string, error) {
	resp, err := client.DepositFeeAccount(
		context.Background(), &pboperator.DepositFeeAccountRequest{
			NumOfAddresses: int64(numOfAddresses),
		},
	)
	if err != nil {
		return nil, err
	}

	addresses := make([]string, 0, len(resp.GetAddressWithBlindingKey()))
	for _, v := range resp.GetAddressWithBlindingKey() {
		addresses = append(addresses, v.Address)
	}

	return addresses, nil
}
