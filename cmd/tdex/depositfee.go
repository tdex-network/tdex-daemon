package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"time"

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
	},
}

func depositFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	net, baseAssetKey := getNetworkAndBaseAssetKey(ctx.String("network"))
	explorerUrl := ctx.String("explorer")
	fragmentationDisabled := ctx.Bool("no-fragment")

	if fragmentationDisabled {
		resp, err := client.DepositFeeAccount(
			context.Background(), &pboperator.DepositFeeAccountRequest{
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
	baseAssetPair, unspents := findBaseAssetsUnspents(
		randomWallet,
		explorerSvc,
		baseAssetKey,
	)

	log.Info("calculating fragments ...")
	baseFragments := fragmentFeeUnspents(
		baseAssetPair.BaseValue,
		MinFee,
		MaxNumOfOutputs,
	)
	log.Infof("base fragments: %v", baseFragments)

	addresses, err := fetchFeeAccountAddresses(
		len(baseFragments),
		client,
	)
	if err != nil {
		log.Error(err)
	}

	feeAmount := wallet.EstimateTxSize(
		len(unspents),
		len(baseFragments),
		false,
		MinMilliSatPerByte,
	)

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

retry:
	for {
		resp, err := explorerSvc.BroadcastTransaction(txHex)
		if err != nil {
			log.Warn(err)
			log.Info("transaction broadcast retry")
			continue retry
		}
		log.Info(resp)
		break retry
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

type BaseAssetPair struct {
	BaseValue uint64
	BaseAsset string
}

// findBaseAssetsUnspents polls blockchain until base asset unspent is noticed
func findBaseAssetsUnspents(
	randomWallet *trade.Wallet,
	explorerSvc explorer.Service,
	baseAssetKey string,
) (
	BaseAssetPair,
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

			switch len(valuePerAsset) {
			case 1:
				for k, v := range valuePerAsset {
					if k == baseAssetKey {
						if v < MinFee {
							log.Warnf(
								"min base deposit is %v please top up",
								MinFee,
							)
							continue events
						}
						log.Infof(
							"base asset %v funded with value %v",
							k,
							v,
						)
						break events
					}

				}
			}
		}

		time.Sleep(time.Duration(CrawlInterval) * time.Second)
	}

	return BaseAssetPair{
		BaseValue: valuePerAsset[baseAssetKey],
		BaseAsset: baseAssetKey,
	}, unspents
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
