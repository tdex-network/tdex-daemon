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
	"github.com/vulpemventures/go-elements/network"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"

	"github.com/urfave/cli/v2"
)

const (
	MinFee          = 5000
	MaxNumOfOutputs = 50
)

var fragmentfee = cli.Command{
	Name: "fragmentfee",
	Usage: "deposit funds for fee account into an ephemeral wallet, then " +
		"split the amount into multiple fragments and deposit into the daemon",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "txid",
			Usage: "txid of the funds to resume a previous fragmentfee",
		},
		&cli.Uint64Flag{
			Name: "max_fragments",
			Usage: fmt.Sprintf(
				"specify the max number of fragments created. "+
					"Values over %d will be overridden to %d",
				MaxNumOfOutputs, MaxNumOfOutputs,
			),
			Value: MaxNumOfOutputs,
		},
		&cli.StringFlag{
			Name:  "recover_funds_to_address",
			Usage: "specify an address where to send funds stuck into the fragmenter to",
		},
	},
	Action: fragmentFeeAction,
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
	recoverAddress := ctx.String("recover_funds_to_address")
	maxNumOfFragments := ctx.Int("max_fragments")
	if maxNumOfFragments > MaxNumOfOutputs {
		maxNumOfFragments = MaxNumOfOutputs
	}

	if recoverAddress != "" {
		return recoverFundsToAddress(net, walletType, recoverAddress)
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
		return fmt.Errorf(
			"expected to resume previous fragmentation but no txids were provided." +
				" Please retry by specifying txids with --txid",
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
	baseFragments := fragmentFeeUnspents(baseAssetValue, MinFee, maxNumOfFragments)
	feeAmount := estimateFees(len(unspents), len(baseFragments))
	baseFragments = deductFeeFromFragments(baseFragments, feeAmount)

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

// findBaseAssetsUnspents polls blockchain until base asset unspent is noticed
func findBaseAssetsUnspents(
	randomWallet *trade.Wallet,
	explorerSvc explorer.Service,
	baseAssetKey string,
	txids []string,
) (uint64, []explorer.Utxo, error) {
	unspents := make([]explorer.Utxo, 0)
	valuePerAsset := make(map[string]uint64)

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
	ins := make([]int, 0, numIns)
	for i := 0; i < numIns; i++ {
		ins = append(ins, wallet.P2WPKH)
	}

	outs := make([]int, 0, numOuts)
	for i := 0; i < numOuts; i++ {
		outs = append(outs, wallet.P2WPKH)
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
	log.Info("Enter txid of fund(s) separated by a white space [press ENTER to skip or confirm]: ")
	in, _ := reader.ReadString('\n')
	trimmedIn := strings.Trim(in, "\n")
	if trimmedIn == "" {
		return nil
	}
	return strings.Split(trimmedIn, " ")
}

func recoverFundsToAddress(net *network.Network, walletType, addr string) error {
	explorerSvc, err := getExplorerFromState()
	if err != nil {
		return fmt.Errorf("error while setting up explorer service: %v", err)
	}

	walletKeys, err := getWalletFromState(walletType)
	if err != nil {
		return err
	}
	if walletKeys == nil {
		return fmt.Errorf("no ephemeral wallet detected, aborting")
	}

	ephWallet, err := getEphemeralWallet(walletType, walletKeys, net)
	if err != nil {
		return fmt.Errorf("unable to restore ephemeral wallet: %v", err)
	}

	log.Info("recovering all unspents owned by fragmenter...")
	unspents, err := explorerSvc.GetUnspents(ephWallet.Address(), [][]byte{ephWallet.BlindingKey()})
	if err != nil {
		return err
	}

	totalAmountPerAsset := make(map[string]uint64)
	for _, u := range unspents {
		totalAmountPerAsset[u.Asset()] += u.Value()
	}

	numIns := len(unspents)
	numOuts := len(totalAmountPerAsset)
	feeAmount := estimateFees(numIns, numOuts)

	if totalLbtcAmount := totalAmountPerAsset[net.AssetID]; totalLbtcAmount < feeAmount {
		return fmt.Errorf("fragment does not hold enough funds to pay for network fees")
	}
	totalAmountPerAsset[net.AssetID] -= feeAmount

	log.Infof("found %d unspents with total amount per asset:", len(unspents))

	outs := make([]TxOut, 0, len(totalAmountPerAsset))
	for asset, amount := range totalAmountPerAsset {
		msg := fmt.Sprintf("%s: %d", asset, amount)
		if asset == net.AssetID {
			msg += (" (network fees deducted)")
		}
		log.Info(msg)
		outs = append(outs, TxOut{
			Asset:   asset,
			Address: addr,
			Value:   int64(amount),
		})
	}

	confirmed := waitForOperatorConfirmation()
	if !confirmed {
		log.Info("aborting")
		return nil
	}

	txHex, err := buildFinalizedTx(net, ephWallet, unspents, outs, feeAmount)
	if err != nil {
		return err
	}

	txid, err := explorerSvc.BroadcastTransaction(txHex)
	if err != nil {
		return err
	}

	log.Infof("funds sent to address %s in tx with id: %s", addr, txid)

	flushWallet(walletType)
	log.Info("done")
	return nil
}

func waitForOperatorConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	log.Info("do you want to proceed [y/n]?")
	in, _ := reader.ReadString('\n')
	trimmedIn := strings.Trim(in, "\n")
	if trimmedIn == "" {
		return true
	}
	return strings.ToLower(trimmedIn) == "y"
}
