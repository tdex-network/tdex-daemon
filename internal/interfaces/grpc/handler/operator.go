package grpchandler

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
)

type operatorHandler struct {
	operatorSvc application.OperatorService
}

// NewOperatorHandler is a constructor function returning an protobuf OperatorServer.
func NewOperatorHandler(
	operatorSvc application.OperatorService,
) daemonv2.OperatorServiceServer {
	return newOperatorHandler(operatorSvc)
}

func newOperatorHandler(
	operatorSvc application.OperatorService,
) *operatorHandler {
	return &operatorHandler{operatorSvc}
}

func (h *operatorHandler) DeriveFeeAddresses(
	ctx context.Context, req *daemonv2.DeriveFeeAddressesRequest,
) (*daemonv2.DeriveFeeAddressesResponse, error) {
	return h.deriveFeeAddresses(ctx, req)
}

func (h *operatorHandler) ListFeeAddresses(
	ctx context.Context, req *daemonv2.ListFeeAddressesRequest,
) (*daemonv2.ListFeeAddressesResponse, error) {
	return h.listFeeAddresses(ctx, req)
}

func (h *operatorHandler) GetFeeBalance(
	ctx context.Context, req *daemonv2.GetFeeBalanceRequest,
) (*daemonv2.GetFeeBalanceResponse, error) {
	return h.getFeeBalance(ctx, req)
}

func (h *operatorHandler) WithdrawFee(
	ctx context.Context, req *daemonv2.WithdrawFeeRequest,
) (*daemonv2.WithdrawFeeResponse, error) {
	return h.withdrawFee(ctx, req)
}

func (h *operatorHandler) NewMarket(
	ctx context.Context, req *daemonv2.NewMarketRequest,
) (*daemonv2.NewMarketResponse, error) {
	return h.newMarket(ctx, req)
}

func (h *operatorHandler) GetMarketInfo(
	ctx context.Context, req *daemonv2.GetMarketInfoRequest,
) (*daemonv2.GetMarketInfoResponse, error) {
	return h.getMarketInfo(ctx, req)
}

func (h *operatorHandler) DeriveMarketAddresses(
	ctx context.Context, req *daemonv2.DeriveMarketAddressesRequest,
) (*daemonv2.DeriveMarketAddressesResponse, error) {
	return h.deriveMarketAddresses(ctx, req)
}

func (h *operatorHandler) ListMarketAddresses(
	ctx context.Context, req *daemonv2.ListMarketAddressesRequest,
) (*daemonv2.ListMarketAddressesResponse, error) {
	return h.listMarketAddresses(ctx, req)
}

func (h *operatorHandler) OpenMarket(
	ctx context.Context, req *daemonv2.OpenMarketRequest,
) (*daemonv2.OpenMarketResponse, error) {
	return h.openMarket(ctx, req)
}

func (h *operatorHandler) CloseMarket(
	ctx context.Context, req *daemonv2.CloseMarketRequest,
) (*daemonv2.CloseMarketResponse, error) {
	return h.closeMarket(ctx, req)
}

func (h *operatorHandler) DropMarket(
	ctx context.Context, req *daemonv2.DropMarketRequest,
) (*daemonv2.DropMarketResponse, error) {
	return h.dropMarket(ctx, req)
}

func (h *operatorHandler) WithdrawMarket(
	ctx context.Context, req *daemonv2.WithdrawMarketRequest,
) (*daemonv2.WithdrawMarketResponse, error) {
	return h.withdrawMarket(ctx, req)
}

func (h *operatorHandler) GetMarketReport(
	ctx context.Context,
	req *daemonv2.GetMarketReportRequest,
) (*daemonv2.GetMarketReportResponse, error) {
	return h.getMarketReport(ctx, req)
}

func (h *operatorHandler) UpdateMarketPercentageFee(
	ctx context.Context, req *daemonv2.UpdateMarketPercentageFeeRequest,
) (*daemonv2.UpdateMarketPercentageFeeResponse, error) {
	return h.updateMarketPercentageFee(ctx, req)
}

func (h *operatorHandler) UpdateMarketFixedFee(
	ctx context.Context, req *daemonv2.UpdateMarketFixedFeeRequest,
) (*daemonv2.UpdateMarketFixedFeeResponse, error) {
	return h.updateMarketFixedFee(ctx, req)
}

func (o operatorHandler) UpdateMarketAssetsPrecision(
	ctx context.Context, req *daemonv2.UpdateMarketAssetsPrecisionRequest,
) (*daemonv2.UpdateMarketAssetsPrecisionResponse, error) {
	return o.updateMarketAssetsPrecision(ctx, req)
}

func (h *operatorHandler) UpdateMarketPrice(
	ctx context.Context, req *daemonv2.UpdateMarketPriceRequest,
) (*daemonv2.UpdateMarketPriceResponse, error) {
	return h.updateMarketPrice(ctx, req)
}

func (h *operatorHandler) UpdateMarketStrategy(
	ctx context.Context, req *daemonv2.UpdateMarketStrategyRequest,
) (*daemonv2.UpdateMarketStrategyResponse, error) {
	return h.updateMarketStrategy(ctx, req)
}

func (h *operatorHandler) DeriveFeeFragmenterAddresses(
	ctx context.Context, req *daemonv2.DeriveFeeFragmenterAddressesRequest,
) (*daemonv2.DeriveFeeFragmenterAddressesResponse, error) {
	return h.deriveFeeFragmenterAddresses(ctx, req)
}

func (h *operatorHandler) ListFeeFragmenterAddresses(
	ctx context.Context, req *daemonv2.ListFeeFragmenterAddressesRequest,
) (*daemonv2.ListFeeFragmenterAddressesResponse, error) {
	return h.listFeeFragmenterAddresses(ctx, req)
}

func (h *operatorHandler) GetFeeFragmenterBalance(
	ctx context.Context,
	req *daemonv2.GetFeeFragmenterBalanceRequest,
) (*daemonv2.GetFeeFragmenterBalanceResponse, error) {
	return h.getFeeFragmenterBalance(ctx, req)
}

func (h *operatorHandler) FeeFragmenterSplitFunds(
	req *daemonv2.FeeFragmenterSplitFundsRequest,
	stream daemonv2.OperatorService_FeeFragmenterSplitFundsServer,
) error {
	return h.feeFragmenterSplitFunds(req, stream)
}

func (h *operatorHandler) WithdrawFeeFragmenter(
	ctx context.Context, req *daemonv2.WithdrawFeeFragmenterRequest,
) (*daemonv2.WithdrawFeeFragmenterResponse, error) {
	return h.withdrawFeeFragmenter(ctx, req)
}

func (h *operatorHandler) DeriveMarketFragmenterAddresses(
	ctx context.Context, req *daemonv2.DeriveMarketFragmenterAddressesRequest,
) (*daemonv2.DeriveMarketFragmenterAddressesResponse, error) {
	return h.deriveMarketFragmenterAddresses(ctx, req)
}

func (h *operatorHandler) ListMarketFragmenterAddresses(
	ctx context.Context, req *daemonv2.ListMarketFragmenterAddressesRequest,
) (*daemonv2.ListMarketFragmenterAddressesResponse, error) {
	return h.listMarketFragmenterAddresses(ctx, req)
}

func (h *operatorHandler) GetMarketFragmenterBalance(
	ctx context.Context,
	req *daemonv2.GetMarketFragmenterBalanceRequest,
) (*daemonv2.GetMarketFragmenterBalanceResponse, error) {
	return h.getMarketFragmenterBalance(ctx, req)
}

func (h *operatorHandler) MarketFragmenterSplitFunds(
	req *daemonv2.MarketFragmenterSplitFundsRequest,
	stream daemonv2.OperatorService_MarketFragmenterSplitFundsServer,
) error {
	return h.marketFragmenterSplitFunds(req, stream)
}

func (h *operatorHandler) WithdrawMarketFragmenter(
	ctx context.Context, req *daemonv2.WithdrawMarketFragmenterRequest,
) (*daemonv2.WithdrawMarketFragmenterResponse, error) {
	return h.withdrawMarketFragmenter(ctx, req)
}

func (h *operatorHandler) ListMarkets(
	ctx context.Context, req *daemonv2.ListMarketsRequest,
) (*daemonv2.ListMarketsResponse, error) {
	return h.listMarkets(ctx, req)
}

func (h *operatorHandler) ListTrades(
	ctx context.Context, req *daemonv2.ListTradesRequest,
) (*daemonv2.ListTradesResponse, error) {
	return h.listTrades(ctx, req)
}

func (h *operatorHandler) ListUtxos(
	ctx context.Context, req *daemonv2.ListUtxosRequest,
) (*daemonv2.ListUtxosResponse, error) {
	return h.listUtxos(ctx, req)
}

func (h *operatorHandler) AddWebhook(
	ctx context.Context, req *daemonv2.AddWebhookRequest,
) (*daemonv2.AddWebhookResponse, error) {
	return h.addWebhook(ctx, req)
}

func (h *operatorHandler) RemoveWebhook(
	ctx context.Context, req *daemonv2.RemoveWebhookRequest,
) (*daemonv2.RemoveWebhookResponse, error) {
	return h.removeWebhook(ctx, req)
}
func (h *operatorHandler) ListWebhooks(
	ctx context.Context, req *daemonv2.ListWebhooksRequest,
) (*daemonv2.ListWebhooksResponse, error) {
	return h.listWebhooks(ctx, req)
}

func (h *operatorHandler) ListDeposits(
	ctx context.Context, req *daemonv2.ListDepositsRequest,
) (*daemonv2.ListDepositsResponse, error) {
	return h.listDeposits(ctx, req)
}

func (h *operatorHandler) ListWithdrawals(
	ctx context.Context, req *daemonv2.ListWithdrawalsRequest,
) (*daemonv2.ListWithdrawalsResponse, error) {
	return h.listWithdrawals(ctx, req)
}

func (h *operatorHandler) deriveFeeAddresses(
	ctx context.Context, req *daemonv2.DeriveFeeAddressesRequest,
) (*daemonv2.DeriveFeeAddressesResponse, error) {
	numOfAddress, err := parseNumOfAddresses(req.GetNumOfAddresses())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	addresses, err := h.operatorSvc.DeriveFeeAddresses(
		ctx, numOfAddress,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.DeriveFeeAddressesResponse{Addresses: addresses}, nil
}

func (h *operatorHandler) listFeeAddresses(
	ctx context.Context, _ *daemonv2.ListFeeAddressesRequest,
) (*daemonv2.ListFeeAddressesResponse, error) {
	addresses, err := h.operatorSvc.ListFeeExternalAddresses(ctx)
	if err != nil {
		return nil, err
	}

	return &daemonv2.ListFeeAddressesResponse{Addresses: addresses}, nil
}

func (h *operatorHandler) getFeeBalance(
	ctx context.Context, _ *daemonv2.GetFeeBalanceRequest,
) (*daemonv2.GetFeeBalanceResponse, error) {
	feeBalance, err := h.operatorSvc.GetFeeBalance(ctx)
	if err != nil {
		return nil, err
	}

	return &daemonv2.GetFeeBalanceResponse{
		Balance: marketBalanceInfo{feeBalance}.toProto(),
	}, nil
}

func (h *operatorHandler) withdrawFee(
	ctx context.Context, req *daemonv2.WithdrawFeeRequest,
) (*daemonv2.WithdrawFeeResponse, error) {
	outputs, err := parseOutputs(req.GetOutputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	msatsPerByte, err := parseMillisatsPerByte(req.GetMillisatsPerByte())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	password, err := parsePassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txid, err := h.operatorSvc.WithdrawFeeFunds(
		ctx, password, outputs, msatsPerByte,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.WithdrawFeeResponse{Txid: txid}, nil
}

func (h *operatorHandler) newMarket(
	ctx context.Context, req *daemonv2.NewMarketRequest,
) (*daemonv2.NewMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	basePercentageFee, quotePercentageFee, err := parseMarketFee(
		req.GetPercentageFee(),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	basePrecision, err := parsePrecision(req.GetBaseAssetPrecision())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	quotePrecision, err := parsePrecision(req.GetQuoteAssetPrecision())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if _, err := h.operatorSvc.NewMarket(
		ctx, market, req.GetName(),
		uint64(basePercentageFee), uint64(quotePercentageFee),
		basePrecision, quotePrecision,
	); err != nil {
		return nil, err
	}

	return &daemonv2.NewMarketResponse{}, nil
}

func (h *operatorHandler) getMarketInfo(
	ctx context.Context, req *daemonv2.GetMarketInfoRequest,
) (*daemonv2.GetMarketInfoResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := h.operatorSvc.GetMarketInfo(ctx, market)
	if err != nil {
		return nil, err
	}
	return &daemonv2.GetMarketInfoResponse{
		Info: marketInfo{info}.toProto(),
	}, nil
}

func (h *operatorHandler) deriveMarketAddresses(
	ctx context.Context, req *daemonv2.DeriveMarketAddressesRequest,
) (*daemonv2.DeriveMarketAddressesResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	addresses, err := h.operatorSvc.DeriveMarketAddresses(
		ctx, market, int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.DeriveMarketAddressesResponse{
		Addresses: addresses,
	}, nil
}

func (h *operatorHandler) listMarketAddresses(
	ctx context.Context, req *daemonv2.ListMarketAddressesRequest,
) (*daemonv2.ListMarketAddressesResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	addresses, err := h.operatorSvc.ListMarketExternalAddresses(ctx, market)
	if err != nil {
		return nil, err
	}

	return &daemonv2.ListMarketAddressesResponse{
		Addresses: addresses,
	}, nil
}

func (h *operatorHandler) openMarket(
	ctx context.Context, req *daemonv2.OpenMarketRequest,
) (*daemonv2.OpenMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := h.operatorSvc.OpenMarket(ctx, market); err != nil {
		return nil, err
	}

	return &daemonv2.OpenMarketResponse{}, nil
}

func (h *operatorHandler) closeMarket(
	ctx context.Context, req *daemonv2.CloseMarketRequest,
) (*daemonv2.CloseMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := h.operatorSvc.CloseMarket(ctx, market); err != nil {
		return nil, err
	}

	return &daemonv2.CloseMarketResponse{}, nil
}

func (h *operatorHandler) dropMarket(
	ctx context.Context, req *daemonv2.DropMarketRequest,
) (*daemonv2.DropMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := h.operatorSvc.DropMarket(ctx, market); err != nil {
		return nil, err
	}

	return &daemonv2.DropMarketResponse{}, nil
}

func (h *operatorHandler) withdrawMarket(
	ctx context.Context, req *daemonv2.WithdrawMarketRequest,
) (*daemonv2.WithdrawMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	outputs, err := parseOutputs(req.GetOutputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	msatsPerByte, err := parseMillisatsPerByte(req.GetMillisatsPerByte())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	password, err := parsePassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txid, err := h.operatorSvc.WithdrawMarketFunds(
		ctx, password, market, outputs, msatsPerByte,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.WithdrawMarketResponse{Txid: txid}, nil
}

func (h *operatorHandler) getMarketReport(
	ctx context.Context, req *daemonv2.GetMarketReportRequest,
) (*daemonv2.GetMarketReportResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	timeRange, err := parseTimeRange(req.GetTimeRange())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	groupByTimeFrame, err := parseTimeFrame(req.GetTimeFrame())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	report, err := h.operatorSvc.GetMarketReport(
		ctx, market, timeRange, groupByTimeFrame,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.GetMarketReportResponse{
		Report: marketReportInfo{report}.toProto(),
	}, nil
}

func (h *operatorHandler) updateMarketPercentageFee(
	ctx context.Context, req *daemonv2.UpdateMarketPercentageFeeRequest,
) (*daemonv2.UpdateMarketPercentageFeeResponse, error) {
	mkt, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	basePercentageFee, quotePercentageFee, err := parseMarketFee(req.GetFee())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := h.operatorSvc.UpdateMarketPercentageFee(
		ctx, mkt, basePercentageFee, quotePercentageFee,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.UpdateMarketPercentageFeeResponse{
		MarketWithFee: &tdexv2.MarketWithFee{
			Market: market{info.GetMarket()}.toProto(),
			Fee: marketFeeInfo{
				info.GetPercentageFee(), info.GetFixedFee(),
			}.toProto(),
		},
	}, nil
}

func (h *operatorHandler) updateMarketFixedFee(
	ctx context.Context, req *daemonv2.UpdateMarketFixedFeeRequest,
) (*daemonv2.UpdateMarketFixedFeeResponse, error) {
	mkt, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	fixedBaseFee, fixedQuoteFee, err := parseMarketFee(req.GetFee())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := h.operatorSvc.UpdateMarketFixedFee(
		ctx, mkt, fixedBaseFee, fixedQuoteFee,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.UpdateMarketFixedFeeResponse{
		MarketWithFee: &tdexv2.MarketWithFee{
			Market: market{info.GetMarket()}.toProto(),
			Fee: marketFeeInfo{
				info.GetPercentageFee(), info.GetFixedFee(),
			}.toProto(),
		},
	}, nil
}

func (o operatorHandler) updateMarketAssetsPrecision(
	ctx context.Context, req *daemonv2.UpdateMarketAssetsPrecisionRequest,
) (*daemonv2.UpdateMarketAssetsPrecisionResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	baseAssetPrecision := int(req.GetBaseAssetPrecision())
	quoteAssetPrecision := int(req.GetQuoteAssetPrecision())

	if err := o.operatorSvc.UpdateMarketAssetsPrecision(
		ctx, market, baseAssetPrecision, quoteAssetPrecision,
	); err != nil {
		return nil, err
	}

	return &daemonv2.UpdateMarketAssetsPrecisionResponse{}, nil
}

func (h *operatorHandler) updateMarketPrice(
	ctx context.Context, req *daemonv2.UpdateMarketPriceRequest,
) (*daemonv2.UpdateMarketPriceResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	basePrice, quotePrice, err := parsePrice(req.GetPrice())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := h.operatorSvc.UpdateMarketPrice(
		ctx, market, *basePrice, *quotePrice,
	); err != nil {
		return nil, err
	}

	return &daemonv2.UpdateMarketPriceResponse{}, nil
}

func (h *operatorHandler) updateMarketStrategy(
	ctx context.Context, req *daemonv2.UpdateMarketStrategyRequest,
) (*daemonv2.UpdateMarketStrategyResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	strategyType, err := parseStrategy(req.GetStrategyType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := h.operatorSvc.UpdateMarketStrategy(
		ctx, market, strategyType,
	); err != nil {
		return nil, err
	}

	return &daemonv2.UpdateMarketStrategyResponse{}, nil
}

func (h *operatorHandler) deriveFeeFragmenterAddresses(
	ctx context.Context, req *daemonv2.DeriveFeeFragmenterAddressesRequest,
) (*daemonv2.DeriveFeeFragmenterAddressesResponse, error) {
	numOfAddresses, err := parseNumOfAddresses(req.GetNumOfAddresses())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	addresses, err := h.operatorSvc.DeriveFeeFragmenterAddresses(
		ctx, numOfAddresses,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.DeriveFeeFragmenterAddressesResponse{
		Addresses: addresses,
	}, nil
}

func (h *operatorHandler) listFeeFragmenterAddresses(
	ctx context.Context, _ *daemonv2.ListFeeFragmenterAddressesRequest,
) (*daemonv2.ListFeeFragmenterAddressesResponse, error) {
	addresses, err := h.operatorSvc.ListFeeFragmenterExternalAddresses(ctx)
	if err != nil {
		return nil, err
	}

	return &daemonv2.ListFeeFragmenterAddressesResponse{
		Addresses: addresses,
	}, nil
}

func (h *operatorHandler) getFeeFragmenterBalance(
	ctx context.Context, _ *daemonv2.GetFeeFragmenterBalanceRequest,
) (*daemonv2.GetFeeFragmenterBalanceResponse, error) {
	info, err := h.operatorSvc.GetFeeFragmenterBalance(ctx)
	if err != nil {
		return nil, err
	}

	balance := make(map[string]*daemonv2.Balance)
	for asset, bal := range info {
		balance[asset] = marketBalanceInfo{bal}.toProto()
	}

	return &daemonv2.GetFeeFragmenterBalanceResponse{Balance: balance}, nil
}

func (h *operatorHandler) feeFragmenterSplitFunds(
	req *daemonv2.FeeFragmenterSplitFundsRequest,
	stream daemonv2.OperatorService_FeeFragmenterSplitFundsServer,
) error {
	maxFragments, err := parseMaxFragments(req.GetMaxFragments())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	msatsPerByte, err := parseMillisatsPerByte(req.GetMillisatsPerByte())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	chReplies := make(chan ports.FragmenterReply)
	go h.operatorSvc.FeeFragmenterSplitFunds(
		stream.Context(), maxFragments, msatsPerByte, chReplies,
	)

	for reply := range chReplies {
		if reply.GetError() != nil {
			return reply.GetError()
		}

		if err := stream.Send(&daemonv2.FeeFragmenterSplitFundsResponse{
			Message: reply.GetMessage(),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *operatorHandler) withdrawFeeFragmenter(
	ctx context.Context, req *daemonv2.WithdrawFeeFragmenterRequest,
) (*daemonv2.WithdrawFeeFragmenterResponse, error) {
	outputs, err := parseOutputs(req.GetOutputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	msatsPerByte, err := parseMillisatsPerByte(req.GetMillisatsPerByte())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	password, err := parsePassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txid, err := h.operatorSvc.WithdrawFeeFragmenterFunds(
		ctx, password, outputs, msatsPerByte,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.WithdrawFeeFragmenterResponse{Txid: txid}, nil
}

func (h *operatorHandler) deriveMarketFragmenterAddresses(
	ctx context.Context, req *daemonv2.DeriveMarketFragmenterAddressesRequest,
) (*daemonv2.DeriveMarketFragmenterAddressesResponse, error) {
	numOfAddresses, err := parseNumOfAddresses(req.GetNumOfAddresses())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	addresses, err := h.operatorSvc.DeriveMarketFragmenterAddresses(
		ctx, numOfAddresses,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.DeriveMarketFragmenterAddressesResponse{
		Addresses: addresses,
	}, nil
}

func (h *operatorHandler) listMarketFragmenterAddresses(
	ctx context.Context, _ *daemonv2.ListMarketFragmenterAddressesRequest,
) (*daemonv2.ListMarketFragmenterAddressesResponse, error) {
	addresses, err := h.operatorSvc.ListMarketFragmenterExternalAddresses(ctx)
	if err != nil {
		return nil, err
	}

	return &daemonv2.ListMarketFragmenterAddressesResponse{
		Addresses: addresses,
	}, nil
}

func (h *operatorHandler) getMarketFragmenterBalance(
	ctx context.Context, _ *daemonv2.GetMarketFragmenterBalanceRequest,
) (*daemonv2.GetMarketFragmenterBalanceResponse, error) {
	info, err := h.operatorSvc.GetMarketFragmenterBalance(ctx)
	if err != nil {
		return nil, err
	}

	balance := make(map[string]*daemonv2.Balance)
	for asset, bal := range info {
		balance[asset] = marketBalanceInfo{bal}.toProto()
	}

	return &daemonv2.GetMarketFragmenterBalanceResponse{Balance: balance}, nil
}

func (h *operatorHandler) marketFragmenterSplitFunds(
	req *daemonv2.MarketFragmenterSplitFundsRequest,
	stream daemonv2.OperatorService_MarketFragmenterSplitFundsServer,
) error {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	msatsPerByte, err := parseMillisatsPerByte(req.GetMillisatsPerByte())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	chReplies := make(chan ports.FragmenterReply)
	go h.operatorSvc.MarketFragmenterSplitFunds(
		stream.Context(), market, msatsPerByte, chReplies,
	)

	for reply := range chReplies {
		if reply.GetError() != nil {
			return reply.GetError()
		}

		if err := stream.Send(&daemonv2.MarketFragmenterSplitFundsResponse{
			Message: reply.GetMessage(),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *operatorHandler) withdrawMarketFragmenter(
	ctx context.Context, req *daemonv2.WithdrawMarketFragmenterRequest,
) (*daemonv2.WithdrawMarketFragmenterResponse, error) {
	outputs, err := parseOutputs(req.GetOutputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	msatsPerByte, err := parseMillisatsPerByte(req.GetMillisatsPerByte())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	password, err := parsePassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txid, err := h.operatorSvc.WithdrawMarketFragmenterFunds(
		ctx, password, outputs, msatsPerByte,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.WithdrawMarketFragmenterResponse{
		Txid: txid,
	}, nil
}

func (h *operatorHandler) listTrades(
	ctx context.Context, req *daemonv2.ListTradesRequest,
) (*daemonv2.ListTradesResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	page, err := parsePage(req.GetPage())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	trades, err := h.operatorSvc.ListTradesForMarket(ctx, market, page)
	if err != nil {
		return nil, err
	}

	return &daemonv2.ListTradesResponse{Trades: tradesInfo(trades).toProto()}, nil
}

func (h *operatorHandler) listMarkets(
	ctx context.Context, _ *daemonv2.ListMarketsRequest,
) (*daemonv2.ListMarketsResponse, error) {
	infoList, err := h.operatorSvc.ListMarkets(ctx)
	if err != nil {
		return nil, err
	}

	markets := make([]*daemonv2.MarketInfo, 0)
	for _, info := range infoList {
		markets = append(markets, marketInfo{info}.toProto())
	}

	return &daemonv2.ListMarketsResponse{Markets: markets}, nil
}

func (h *operatorHandler) listUtxos(
	ctx context.Context, req *daemonv2.ListUtxosRequest,
) (*daemonv2.ListUtxosResponse, error) {
	accountName, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	page, err := parsePage(req.GetPage())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	spendableUtxosInfo, lockedUtxosInfo, err := h.operatorSvc.ListUtxos(
		ctx, accountName, page,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.ListUtxosResponse{
		SpendableUtxos: utxosInfo(spendableUtxosInfo).toProto(),
		LockedUtxos:    utxosInfo(lockedUtxosInfo).toProto(),
	}, nil
}

func (h *operatorHandler) addWebhook(
	ctx context.Context, req *daemonv2.AddWebhookRequest,
) (*daemonv2.AddWebhookResponse, error) {
	webhook, err := parseWebhook(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	hookID, err := h.operatorSvc.AddWebhook(ctx, webhook)
	if err != nil {
		return nil, err
	}
	return &daemonv2.AddWebhookResponse{Id: hookID}, nil
}

func (h *operatorHandler) removeWebhook(
	ctx context.Context, req *daemonv2.RemoveWebhookRequest,
) (*daemonv2.RemoveWebhookResponse, error) {
	if err := h.operatorSvc.RemoveWebhook(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &daemonv2.RemoveWebhookResponse{}, nil
}

func (h *operatorHandler) listWebhooks(
	ctx context.Context, req *daemonv2.ListWebhooksRequest,
) (*daemonv2.ListWebhooksResponse, error) {
	actionType, err := parseWebhookActionType(req.GetAction())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	hooks, err := h.operatorSvc.ListWebhooks(ctx, actionType)
	if err != nil {
		return nil, err
	}
	return &daemonv2.ListWebhooksResponse{
		WebhookInfo: hooksInfo(hooks).toProto(),
	}, nil
}

func (h *operatorHandler) listDeposits(
	ctx context.Context, req *daemonv2.ListDepositsRequest,
) (*daemonv2.ListDepositsResponse, error) {
	accountName, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	page, err := parsePage(req.GetPage())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	deposits, err := h.operatorSvc.ListDeposits(
		ctx, accountName, page,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.ListDepositsResponse{
		AccountName: accountName,
		Deposits:    depositsInfo(deposits).toProto(),
	}, err
}

func (h *operatorHandler) listWithdrawals(
	ctx context.Context, req *daemonv2.ListWithdrawalsRequest,
) (*daemonv2.ListWithdrawalsResponse, error) {
	accountName, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	page, err := parsePage(req.GetPage())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	withdrawals, err := h.operatorSvc.ListWithdrawals(
		ctx, accountName, page,
	)

	return &daemonv2.ListWithdrawalsResponse{
		AccountName: accountName,
		Withdrawals: withdrawalsInfo(withdrawals).toProto(),
	}, err
}
