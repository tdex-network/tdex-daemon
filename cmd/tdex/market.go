package main

import (
	"context"
	"fmt"
	"strings"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/urfave/cli/v2"
)

var (
	marketAccount = cli.Command{
		Name:  "market",
		Usage: "manage a market account of the daemon's wallet",
		Subcommands: []*cli.Command{
			marketNewCmd, marketInfoCmd, marketListAddressesCmd,
			marketDepositCmd, marketWithdrawCmd,
			marketOpenCmd, marketCloseCmd, marketDropCmd,
			marketUpdateFixedFeeCmd, marketUpdatePercentageFeeCmd,
			marketUpdateStrategyCmd, marketUpdatePriceCmd, marketReportCmd,
			marketUpdateAssetsPrecisionCmd,
		},
	}

	marketNewCmd = &cli.Command{
		Name:  "new",
		Usage: "create a new market",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "base-asset",
				Usage: "the hash of the base asset of the market",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "quote-asset",
				Usage: "the hash of the quote asset of the market",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "optional name for the market",
				Value: "",
			},
			&cli.Uint64Flag{
				Name:  "percentage-base-fee",
				Usage: "the percentage fee on base asset",
				Value: 0,
			},
			&cli.Uint64Flag{
				Name:  "percentage-quote-fee",
				Usage: "the percentage fee on quote asset",
				Value: 0,
			},
			&cli.Uint64Flag{
				Name:  "fixed-base-fee",
				Usage: "the fixed fee on base asset",
				Value: 0,
			},
			&cli.Uint64Flag{
				Name:  "fixed-quote-fee",
				Usage: "the fixed fee on quote asset",
				Value: 0,
			},
			&cli.UintFlag{
				Name:  "base-asset-precision",
				Usage: "the precision for the base asset",
				Value: 0,
			},
			&cli.UintFlag{
				Name:  "quote-asset-precision",
				Usage: "the precision for the quote asset",
				Value: 0,
			},
			&cli.StringFlag{
				Name:  "strategy",
				Usage: "the market strategy to use, either BALANCED or PLUGGABLE",
				Value: "",
			},
		},
		Action: newMarketAction,
	}
	marketInfoCmd = &cli.Command{
		Name:   "info",
		Usage:  "get info about the current market",
		Action: marketInfoAction,
	}
	marketDepositCmd = &cli.Command{
		Name:  "deposit",
		Usage: "generate some address(es) to deposit funds for a market",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "num-of-addresses",
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
	marketWithdrawCmd = &cli.Command{
		Name:  "withdraw",
		Usage: "withdraw some funds to an address",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "receivers",
				Usage: "list of withdrawal receivers as {asset, amount, address}",
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
				Name:  "base-fee",
				Usage: "set the fixed fee for base asset",
				Value: -1,
			},
			&cli.Int64Flag{
				Name:  "quote-fee",
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
				Name:  "base-fee",
				Usage: "set the percentage fee for base asset",
				Value: -1,
			},
			&cli.Int64Flag{
				Name:  "quote-fee",
				Usage: "set the percentage fee for quote asset",
				Value: -1,
			},
		},
		Action: marketUpdatePercentageFeeAction,
	}
	marketUpdateStrategyCmd = &cli.Command{
		Name:  "strategy",
		Usage: "updates the current market making strategy, either automated or pluggable market making",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "balanced",
				Usage: "set the strategy to balanced AMM",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "pluggable",
				Usage: "set the strategy to pluggable",
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
				Name:  "base-asset",
				Usage: "the precision for the base asset",
				Value: -1,
			},
			&cli.IntFlag{
				Name:  "quote-asset",
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
				Name:     "base-price",
				Usage:    "the base price, or the amount of quote asset needed to buy 1 BTC of base asset",
				Required: true,
			},
			&cli.Float64Flag{
				Name:     "quote-price",
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
				Usage: "custom start date expressed in RFC3339 format",
			},
			&cli.StringFlag{
				Name:  "end",
				Usage: "custom end date expressed in RFC3339 format",
			},
			&cli.BoolFlag{
				Name:  "last-hour",
				Usage: "get a market report for the last hour",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "last-day",
				Usage: "get a market report for the last 24 hours",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "last-month",
				Usage: "get a market report for the last month",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "last-three-months",
				Usage: "get a market report for the last 3 months",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "last-year",
				Usage: "get a market report for the last year",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "year-to-date",
				Usage: "get a market report from the beginning of the year until now",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "get a market report including all trades",
				Value: false,
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

	name := ctx.String("name")
	baseAsset := ctx.String("base-asset")
	quoteAsset := ctx.String("quote-asset")
	basePrecision := ctx.Uint("base-asset-precision")
	quotePrecision := ctx.Uint("quote-asset-precision")
	basePercentageFee := ctx.Uint64("percentage-base-fee")
	quotePercentageFee := ctx.Uint64("percentage-quote-fee")
	baseFixedFee := ctx.Uint64("fixed-base-fee")
	quoteFixedFee := ctx.Uint64("fixed-quote-fee")
	strategy := ctx.String("strategy")
	strategyType := daemonv2.StrategyType_STRATEGY_TYPE_UNSPECIFIED
	if len(strategy) > 0 {
		if strings.ToLower(strategy) == "balanced" {
			strategyType = daemonv2.StrategyType_STRATEGY_TYPE_BALANCED
		}
		if strings.ToLower(strategy) == "pluggable" {
			strategyType = daemonv2.StrategyType_STRATEGY_TYPE_PLUGGABLE
		}
	}

	if _, err := client.NewMarket(
		context.Background(), &daemonv2.NewMarketRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			Name:                name,
			BaseAssetPrecision:  uint32(basePrecision),
			QuoteAssetPrecision: uint32(quotePrecision),
			PercentageFee: &tdexv2.MarketFee{
				BaseAsset:  int64(basePercentageFee),
				QuoteAsset: int64(quotePercentageFee),
			},
			FixedFee: &tdexv2.MarketFee{
				BaseAsset:  int64(baseFixedFee),
				QuoteAsset: int64(quoteFixedFee),
			},
			StrategyType: strategyType,
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

	numOfAddresses := ctx.Int64("num-of-addresses")
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
	mSatsPerByte := ctx.Uint64("millisats-per-byte")
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

	baseFee := ctx.Int64("base-fee")
	quoteFee := ctx.Int64("quote-fee")
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

	baseFee := ctx.Int64("base-fee")
	quoteFee := ctx.Int64("quote-fee")
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

	pluggable := ctx.Bool("pluggable")
	balanced := ctx.Bool("balanced")
	if pluggable && balanced {
		return fmt.Errorf("only one strategy type must be specified")
	}
	if !pluggable && !balanced {
		return fmt.Errorf("missing strategy type")
	}

	var strategy daemonv2.StrategyType
	if pluggable {
		strategy = daemonv2.StrategyType_STRATEGY_TYPE_PLUGGABLE
	}
	if balanced {
		strategy = daemonv2.StrategyType_STRATEGY_TYPE_BALANCED
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
				BasePrice:  ctx.Float64("base-price"),
				QuotePrice: ctx.Float64("quote-price"),
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
	var predefinedPeriod daemonv2.PredefinedPeriod
	start := ctx.String("start")
	end := ctx.String("end")
	if ctx.IsSet("start") != ctx.IsSet("end") {
		return fmt.Errorf("both start and end dates must defined for a custom period")
	}
	if start != "" && end != "" {
		customPeriod = &daemonv2.CustomPeriod{
			StartDate: start,
			EndDate:   end,
		}
	} else {
		pp := map[daemonv2.PredefinedPeriod]bool{
			daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_HOUR:         ctx.Bool("last-hour"),
			daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_DAY:          ctx.Bool("last-day"),
			daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_MONTH:        ctx.Bool("last-month"),
			daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_THREE_MONTHS: ctx.Bool("last-three-months"),
			daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_LAST_YEAR:         ctx.Bool("last-year"),
			daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_YEAR_TO_DATE:      ctx.Bool("year-to-date"),
			daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_ALL:               ctx.Bool("all"),
		}
		count := 0
		for period, isSet := range pp {
			if isSet {
				predefinedPeriod = period
				count++
			}
		}
		if count == 0 {
			return fmt.Errorf(
				"missing time range, specifiy either a predefined or a custom one",
			)
		}
		if count > 1 {
			return fmt.Errorf("only one predefined period must be specified")
		}
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

	basePrecision := ctx.Int("base-asset")
	quotePrecision := ctx.Int("quote-asset")

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
