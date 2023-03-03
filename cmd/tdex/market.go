package main

import (
	"context"
	"fmt"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/urfave/cli/v2"
)

var (
	marketAccount = cli.Command{
		Name:  "market",
		Usage: "manage a market account of the daemon's wallet",
		Subcommands: []*cli.Command{
			marketNewCmd, marketInfoCmd, marketBalanceCmd, marketListAddressesCmd,
			marketDepositCmd, marketClaimCmd, marketWithdrawCmd,
			marketOpenCmd, marketCloseCmd, marketDropCmd,
			marketUpdateFixedFeeCmd, marketUpdatePercentageFeeCmd, marketReportFeeCmd,
			marketUpdateStrategyCmd, marketUpdatePriceCmd, marketReportCmd,
			marketUpdateAssetsPrecisionCmd,
		},
	}

	marketNewCmd = &cli.Command{
		Name:  "new",
		Usage: "create a new market",
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
			&cli.UintFlag{
				Name:  "base_asset_precision",
				Usage: "the precision for the base asset",
				Value: 0,
			},
			&cli.UintFlag{
				Name:  "quote_asset_precision",
				Usage: "the precision for the quote asset",
				Value: 0,
			},
		},
		Action: newMarketAction,
	}
	marketInfoCmd = &cli.Command{
		Name:   "info",
		Usage:  "get info about the current market",
		Action: marketInfoAction,
	}
	marketBalanceCmd = &cli.Command{
		Name:   "balance",
		Usage:  "DEPRECATED: get the balance of a market",
		Action: marketBalanceAction,
	}
	marketDepositCmd = &cli.Command{
		Name:  "deposit",
		Usage: "generate some address(es) to deposit funds for a market",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "num_of_addresses",
				Usage: "the number of addresses to generate for the market",
			},
		},
		Action: marketDepositAction,
	}
	marketListAddressesCmd = &cli.Command{
		Name:   "addresses",
		Usage:  "list all the derived deposit addresses of a market",
		Action: marketListAddressesAction,
	}
	marketClaimCmd = &cli.Command{
		Name:   "claim",
		Usage:  "DEPRECATED: claim deposits for a market",
		Action: marketClaimAction,
	}
	marketWithdrawCmd = &cli.Command{
		Name:  "withdraw",
		Usage: "withdraw some funds to an address",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "receivers",
				Usage: "list of withdrawal receivers as {asset, amount, address}",
			},
			&cli.Uint64Flag{
				Name:  "millisatsperbyte",
				Usage: "the mSat/byte to pay for the transaction",
				Value: 100,
			},
			&cli.StringFlag{
				Name:     "password",
				Usage:    "the wallet unlocking password as security measure",
				Required: true,
			},
		},
		Action: marketWithdrawAction,
	}
	marketOpenCmd = &cli.Command{
		Name:   "open",
		Usage:  "open a market",
		Action: marketOpenAction,
	}
	marketCloseCmd = &cli.Command{
		Name:   "close",
		Usage:  "close a market",
		Action: marketCloseAction,
	}
	marketDropCmd = &cli.Command{
		Name:   "drop",
		Usage:  "drop a market",
		Action: marketDropAction,
	}
	marketUpdateFixedFeeCmd = &cli.Command{
		Name:  "fixedfee",
		Usage: "updates the current market fixed fee",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:  "base_fee",
				Usage: "set the fixed fee for base asset",
				Value: -1,
			},
			&cli.Int64Flag{
				Name:  "quote_fee",
				Usage: "set the fixed fee for quote asset",
				Value: -1,
			},
		},
		Action: marketUpdateFixedFeeAction,
	}
	marketUpdatePercentageFeeCmd = &cli.Command{
		Name:  "percentagefee",
		Usage: "updates the current market percentage fee",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:  "base_fee",
				Usage: "set the percentage fee for base asset",
				Value: -1,
			},
			&cli.Int64Flag{
				Name:  "quote_fee",
				Usage: "set the percentage fee for quote asset",
				Value: -1,
			},
		},
		Action: marketUpdatePercentageFeeAction,
	}
	marketReportFeeCmd = &cli.Command{
		Name:   "reportfee",
		Usage:  "get a report of the fees collected for the trades of a market.",
		Flags:  []cli.Flag{},
		Action: marketReportFeeAction,
	}
	marketUpdateStrategyCmd = &cli.Command{
		Name:  "strategy",
		Usage: "updates the current market making strategy, either automated or pluggable market making",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "pluggable",
				Usage: "set the strategy as pluggable",
				Value: false,
			},
		},
		Action: marketUpdateStrategyAction,
	}
	marketUpdateAssetsPrecisionCmd = &cli.Command{
		Name:  "precision",
		Usage: "updates the precision of one or both market assets",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "base_asset",
				Usage: "the precision for the base asset",
				Value: -1,
			},
			&cli.IntFlag{
				Name:  "quote_asset",
				Usage: "the precision for the quote asset",
				Value: -1,
			},
		},
		Action: marketUpdateAssetsPrecision,
	}
	marketUpdatePriceCmd = &cli.Command{
		Name:  "price",
		Usage: "updates the price of a market",
		Flags: []cli.Flag{
			&cli.Float64Flag{
				Name:     "base_price",
				Usage:    "the base price, or the amount of quote asset needed to buy 1 BTC of base asset",
				Required: true,
			},
			&cli.Float64Flag{
				Name:     "quote_price",
				Usage:    "the quote price, or the amount of base asset need to buy 1 BTC of quote asset",
				Required: true,
			},
		},
		Action: marketUpdatePriceAction,
	}

	marketReportCmd = &cli.Command{
		Name:   "report",
		Usage:  "get market report about collected fees and trade volume for a specified time range",
		Action: marketReportAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "start",
				Usage: "fetch balances from specific time in the past, please provide end flag also",
			},
			&cli.StringFlag{
				Name:  "end",
				Usage: "fetch balances from specific time in the past til end date, use with start flag",
			},
			&cli.IntFlag{
				Name: "predefined_period",
				Usage: "time predefined periods:\n" +
					"       1 -> last hour\n" +
					"       2 -> last day\n" +
					"       3 -> last month\n" +
					"       4 -> last 3 months\n" +
					"       5 -> year to date\n" +
					"       6 -> all",
				Value: 2,
			},
		},
	}
)

func newMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset := ctx.String("base_asset")
	quoteAsset := ctx.String("quote_asset")
	basePrecision := ctx.Uint("base_asset_precision")
	quotePrecision := ctx.Uint("quote_asset_precision")

	if _, err := client.NewMarket(
		context.Background(), &daemonv2.NewMarketRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			BaseAssetPrecision:  uint32(basePrecision),
			QuoteAssetPrecision: uint32(quotePrecision),
		},
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market created")
	return nil
}

func marketInfoAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	resp, err := client.GetMarketInfo(
		context.Background(), &daemonv2.GetMarketInfoRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

func marketBalanceAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market info")
	return nil
}

func marketDepositAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	numOfAddresses := ctx.Int64("num_of_addresses")
	resp, err := client.DeriveMarketAddresses(
		context.Background(),
		&daemonv2.DeriveMarketAddressesRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
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

func marketListAddressesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	resp, err := client.ListMarketAddresses(
		context.Background(), &daemonv2.ListMarketAddressesRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

func marketClaimAction(ctx *cli.Context) error {
	printDeprecatedWarn("")
	return nil
}

func marketWithdrawAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}
	receivers := ctx.StringSlice("receivers")
	password := ctx.String("password")
	mSatsPerByte := ctx.Uint64("millisatsperbyte")
	outputs, err := parseOutputs(receivers)
	if err != nil {
		return err
	}

	reply, err := client.WithdrawMarket(context.Background(), &daemonv2.WithdrawMarketRequest{
		Market: &tdexv2.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		Outputs:          outputs,
		MillisatsPerByte: mSatsPerByte,
		Password:         password,
	})
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}

func marketOpenAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.OpenMarket(
		context.Background(), &daemonv2.OpenMarketRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market is open")
	return nil
}

func marketCloseAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.CloseMarket(
		context.Background(), &daemonv2.CloseMarketRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market is closed")
	return nil
}

func marketDropAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.DropMarket(
		context.Background(), &daemonv2.DropMarketRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market is dropped")
	return nil
}

func marketUpdateFixedFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	baseFee := ctx.Int64("base_fee")
	quoteFee := ctx.Int64("quote_fee")
	req := &daemonv2.UpdateMarketFixedFeeRequest{
		Market: &tdexv2.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		Fee: &tdexv2.MarketFee{
			BaseAsset:  baseFee,
			QuoteAsset: quoteFee,
		},
	}

	if _, err := client.UpdateMarketFixedFee(
		context.Background(), req,
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market fees have been updated")
	return nil
}

func marketUpdatePercentageFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	baseFee := ctx.Int64("base_fee")
	quoteFee := ctx.Int64("quote_fee")
	req := &daemonv2.UpdateMarketPercentageFeeRequest{
		Market: &tdexv2.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		Fee: &tdexv2.MarketFee{
			BaseAsset:  baseFee,
			QuoteAsset: quoteFee,
		},
	}

	if _, err := client.UpdateMarketPercentageFee(
		context.Background(), req,
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market fees have been updated")
	return nil
}

func marketReportFeeAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market report")
	return nil
}

func marketUpdateStrategyAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	strategy := daemonv2.StrategyType_STRATEGY_TYPE_BALANCED
	if ctx.Bool("pluggable") {
		strategy = daemonv2.StrategyType_STRATEGY_TYPE_PLUGGABLE
	}

	_, err = client.UpdateMarketStrategy(
		context.Background(), &daemonv2.UpdateMarketStrategyRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			StrategyType: strategy,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("strategy has been updated")
	return nil
}

func marketUpdatePriceAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.UpdateMarketPrice(
		context.Background(), &daemonv2.UpdateMarketPriceRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			Price: &tdexv2.Price{
				BasePrice:  ctx.Float64("base_price"),
				QuotePrice: ctx.Float64("quote_price"),
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("price has been updated")
	return nil
}

func marketReportAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	var customPeriod *daemonv2.CustomPeriod
	start := ctx.String("start")
	end := ctx.String("end")
	if start != "" && end != "" {
		customPeriod = &daemonv2.CustomPeriod{
			StartDate: start,
			EndDate:   end,
		}
	}

	var predefinedPeriod daemonv2.PredefinedPeriod
	pp := ctx.Int("predefined_period")
	if pp > 0 {
		predefinedPeriod = daemonv2.PredefinedPeriod(pp)
	}

	reply, err := client.GetMarketReport(
		context.Background(),
		&daemonv2.GetMarketReportRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			TimeRange: &daemonv2.TimeRange{
				PredefinedPeriod: predefinedPeriod,
				CustomPeriod:     customPeriod,
			},
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}

func marketUpdateAssetsPrecision(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	basePrecision := ctx.Int("base_asset")
	quotePrecision := ctx.Int("quote_asset")

	if _, err := client.UpdateMarketAssetsPrecision(
		context.Background(), &daemonv2.UpdateMarketAssetsPrecisionRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			BaseAssetPrecision:  int32(basePrecision),
			QuoteAssetPrecision: int32(quotePrecision),
		},
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("precisions have been updated")
	return nil
}
