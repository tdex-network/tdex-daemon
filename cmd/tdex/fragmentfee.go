package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"

	"github.com/urfave/cli/v2"
)

const (
	MinFee          = 5000
	MaxNumOfOutputs = 150
)

var fragmentfee = cli.Command{
	Name: "fragmentfee",
	Usage: "deposit funds for fee account into an ephemeral wallet, then " +
		"split the amount into multiple fragments and deposit into the daemon",
	Action: fragmentFeeAction,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "txid",
			Usage: "txid of the funds to resume a previous fragmentfee",
		},
	},
}

func fragmentFeeAction(ctx *cli.Context) error {
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

	walletType := "fee"
	txids := ctx.StringSlice("txid")

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
		return fmt.Errorf(
			"expected to resume previous fragmentation but no txids were provided." +
				" Please retry by specifying txids with --txids",
		)
	}

	ephWallet, err := getEphemeralWallet(walletType, walletKeys, net)
	if err != nil {
		return fmt.Errorf("unable to restore ephemeral wallet: %v", err)
	}

	if len(txids) > 0 && walletKeys != nil {
		log.Info("resuming previous fragmentation")
		log.Infof("you can optionally send other funds to address: %s", ephWallet.Address())
	} else {
		log.Infof("send funds to address: %s", ephWallet.Address())
	}

	var baseAssetValue uint64
	var unspents []explorer.Utxo
	for {
		funds := waitForOperatorFunds()
		funds = append(funds, txids...)

		baseAssetValue, unspents, err = findBaseAssetsUnspents(
			ephWallet,
			explorerSvc,
			baseAssetKey,
			funds,
		)
		if err != nil {
			log.WithError(err).Warn("an unexpected error occured, please retry entering all txids")
			continue
		}
		break
	}

	log.Info("calculating fragments...")
	baseFragments := fragmentFeeUnspents(baseAssetValue, MinFee, MaxNumOfOutputs)

	numUnspents := len(unspents)
	numFragments := len(baseFragments)
	log.Infof(
		"detected %d fund(s) of total amount %d that will be split into %d fragments",
		numUnspents,
		baseAssetValue,
		numFragments,
	)

	addresses, err := getFeeDepositAddresses(numFragments, client)
	if err != nil {
		return err
	}

	feeAmount := estimateFees(numUnspents, numFragments)

	log.Info("crafting transaction...")
	txHex, err := craftTransaction(
		ephWallet,
		unspents,
		baseFragments, nil,
		addresses,
		feeAmount,
		net,
		AssetValuePair{BaseAsset: baseAssetKey},
	)
	if err != nil {
		return err
	}

	log.Info("sending transactions...")
	txID, err := explorerSvc.BroadcastTransaction(txHex)
	if err != nil {
		return err
	}
	log.Infof("fee account funding txid: %s", txID)

	log.Info("waiting for tx to get confirmed...")
	if err := waitUntilTxConfirmed(explorerSvc, txID); err != nil {
		return err
	}

	log.Info("claiming fee deposits...")
	outpoints := createOutpoints(txID, numFragments)
	if _, err := client.ClaimFeeDeposit(
		context.Background(), &pboperator.ClaimFeeDepositRequest{
			Outpoints: outpoints,
		},
	); err != nil {
		return err
	}

	flushWallet(walletType)
	log.Info("done")
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
		if len(res) > 0 {
			res[len(res)-1] += valueToBeFragmented
		} else {
			res = append(res, valueToBeFragmented)
		}
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
		// deduct fee from last(largest) fragment
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
	txids []string,
) (uint64, []explorer.Utxo, error) {
	unspents := make([]explorer.Utxo, 0)
	valuePerAsset := make(map[string]uint64, 0)

	for _, txid := range txids {
		u, err := getUnspents(explorerSvc, randomWallet, txid)
		if err != nil {
			return 0, nil, err
		}
		if len(u) > 0 {
			for _, v := range u {
				valuePerAsset[v.Asset()] += v.Value()
				if v.Asset() == baseAssetKey {
					unspents = append(unspents, v)
				}
			}
		}
	}

	if baseTotalAmount, ok := valuePerAsset[baseAssetKey]; ok {
		if baseTotalAmount < MinFee {
			log.Warnf(
				"min base deposit is %v please top up with another depositfee operation",
				MinFee,
			)
		} else {
			log.Infof(
				"base asset %v funded with value %v",
				baseAssetKey,
				baseTotalAmount,
			)
		}
	}

	return valuePerAsset[baseAssetKey], unspents, nil
}

func getFeeDepositAddresses(
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

func estimateFees(numIns, numOuts int) uint64 {
	ins := make([]int, numIns, numIns)
	for i := 0; i < numIns; i++ {
		ins[i] = wallet.P2WPKH
	}

	outs := make([]int, numOuts, numOuts)
	for i := 0; i < numOuts; i++ {
		outs[i] = wallet.P2WPKH
	}

	size := wallet.EstimateTxSize(ins, nil, nil, outs, nil)
	return uint64(float64(size) * 0.1)
}

func waitUntilTxConfirmed(explorerSvc explorer.Service, txid string) error {
	for {
		isConfirmed, err := explorerSvc.IsTransactionConfirmed(txid)
		if err != nil {
			return err
		}
		if isConfirmed {
			return nil
		}
		sleepTime := 20 * time.Second
		time.Sleep(sleepTime)
	}
}

func createOutpoints(txid string, numOuts int) []*pboperator.TxOutpoint {
	outpoints := make([]*pboperator.TxOutpoint, 0, numOuts)
	for i := 0; i < numOuts; i++ {
		outpoints = append(outpoints, &pboperator.TxOutpoint{
			Hash:  txid,
			Index: int32(i),
		})
	}
	return outpoints
}

func getUnspents(explorerSvc explorer.Service, w *trade.Wallet, txid string) ([]explorer.Utxo, error) {
	tx, err := explorerSvc.GetTransaction(txid)
	if err != nil {
		return nil, err
	}

	utxos := make([]explorer.Utxo, 0)
	_, script := w.Script()
	for i, out := range tx.Outputs() {
		if bytes.Equal(out.Script, script) {
			revealed, _ := transactionutil.UnblindOutput(out, w.BlindingKey())
			var valueCommitment, assetCommitment string
			if out.IsConfidential() {
				valueCommitment = hex.EncodeToString(out.Value)
				assetCommitment = hex.EncodeToString(out.Asset)
			}

			utxos = append(utxos, esplora.NewWitnessUtxo(
				txid,
				uint32(i),
				revealed.Value,
				revealed.AssetHash,
				valueCommitment,
				assetCommitment,
				revealed.ValueBlinder,
				revealed.AssetBlinder,
				script,
				out.Nonce,
				out.RangeProof,
				out.SurjectionProof,
				tx.Confirmed(),
			))
		}
	}

	return utxos, nil
}

func waitForOperatorFunds() []string {
	reader := bufio.NewReader(os.Stdin)
	log.Info("Enter txid of fund(s) separated by a white space: ")
	in, _ := reader.ReadString('\n')
	trimmedIn := strings.Trim(in, "\n")
	if trimmedIn == "" {
		return nil
	}
	return strings.Split(trimmedIn, " ")
}
