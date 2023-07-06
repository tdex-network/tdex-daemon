package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/go-elements/address"
)

var (
	marketFragmenterAccount = cli.Command{
		Name:  "marketfragmenter",
		Usage: "manage the market fragmenter account of the daemon's wallet",
		Subcommands: []*cli.Command{
			marketFragmenterBalanceCmd, marketFragmenterDepositCmd,
			marketFragmenterListAddressesCmd, marketFragmenterSplitFundsCmd,
			marketFragmenterWithdrawCmd,
		},
	}

	marketFragmenterBalanceCmd = &cli.Command{
		Name:   "balance",
		Usage:  "get the balance of the market fragmenter account",
		Action: marketFragmenterBalanceAction,
	}
	marketFragmenterDepositCmd = &cli.Command{
		Name:  "deposit",
		Usage: "generate some addresses to receive funds",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "num-of-addresses",
				Usage: "the number of addresses to generate",
			},
		},
		Action: marketFragmenterDepositAction,
	}
	marketFragmenterListAddressesCmd = &cli.Command{
		Name:   "addresses",
		Usage:  "list all the derived deposit addresses of the market fragmenter account",
		Action: marketFragmenterListAddressesAction,
	}
	marketFragmenterSplitFundsCmd = &cli.Command{
		Name:   "split",
		Usage:  "split market fragmenter funds and make them deposits of some market",
		Action: marketFragmenterSplitFundsAction,
	}
	marketFragmenterWithdrawCmd = &cli.Command{
		Name:  "withdraw",
		Usage: "withdraw all the market fragmenter funds",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "receivers",
				Usage:    "list of withdrawal receivers as {asset, amount, address}",
				Required: true,
			},
			&cli.Uint64Flag{
				Name:  "millisats-per-byte",
				Usage: "the mSat/byte to pay for the transaction",
				Value: 100,
			},
			&cli.StringFlag{
				Name:     "password",
				Usage:    "the wallet unlocking password as security measure",
				Required: true,
			},
		},
		Action: marketFragmenterWithdrawAction,
	}
)

func marketFragmenterBalanceAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.GetMarketFragmenterBalance(
		context.Background(), &daemonv2.GetMarketFragmenterBalanceRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}

func marketFragmenterDepositAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	numOfAddresses := ctx.Int64("num-of-addresses")
	resp, err := client.DeriveMarketFragmenterAddresses(
		context.Background(), &daemonv2.DeriveMarketFragmenterAddressesRequest{
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}

func marketFragmenterListAddressesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListMarketFragmenterAddresses(
		context.Background(), &daemonv2.ListMarketFragmenterAddressesRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}

func marketFragmenterSplitFundsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	stream, err := client.MarketFragmenterSplitFunds(
		context.Background(), &daemonv2.MarketFragmenterSplitFundsRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		})
	if err != nil {
		return err
	}

	for {
		fmt.Println()

		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		fmt.Println(reply.GetMessage())
	}

	return nil
}

func marketFragmenterWithdrawAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	receivers := ctx.StringSlice("receivers")
	password := ctx.String("password")
	mSatsPerByte := ctx.Uint64("millisats-per-byte")

	outputs, err := parseOutputs(receivers)
	if err != nil {
		return err
	}

	reply, err := client.WithdrawMarketFragmenter(
		context.Background(), &daemonv2.WithdrawMarketFragmenterRequest{
			Outputs:          outputs,
			MillisatsPerByte: mSatsPerByte,
			Password:         password,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}

func parseOutputs(receivers []string) ([]*daemonv2.TxOutput, error) {
	outputs := make([]*daemonv2.TxOutput, 0, len(receivers))
	for _, r := range receivers {
		m := make([]map[string]interface{}, 0)
		if err := json.Unmarshal([]byte(r), &m); err != nil {
			return nil, fmt.Errorf("failed to parse receivers: %s", err)
		}
		var asset, script, blindKey string
		var amount uint64
		for _, rr := range m {
			if rr["asset"] != nil {
				asset = rr["asset"].(string)
			}
			if rr["address"] != nil {
				addr := rr["address"].(string)
				scriptBytes, err := address.ToOutputScript(addr)
				if err != nil {
					return nil, fmt.Errorf(
						"failed to parse receiver with addr %s: %s", addr, err,
					)
				}
				script = hex.EncodeToString(scriptBytes)
				info, _ := address.FromConfidential(addr)
				if info != nil {
					blindKey = hex.EncodeToString(info.BlindingKey)
				}
			}
			if rr["amount"] != nil {
				amount = uint64(rr["amount"].(float64))
			}
			outputs = append(outputs, &daemonv2.TxOutput{
				Asset:       asset,
				Amount:      amount,
				Script:      script,
				BlindingKey: blindKey,
			})
		}
	}
	return outputs, nil
}
