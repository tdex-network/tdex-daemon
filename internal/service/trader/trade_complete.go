package tradeservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"

	"github.com/tdex-network/tdex-daemon/internal/domain/trade"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TradeComplete is the domain controller for the TradeComplete RPC
func (s *Service) TradeComplete(req *pb.TradeCompleteRequest, stream pb.Trade_TradeCompleteServer) error {
	ctx := context.Background()
	currentTrade, err := s.tradeRepository.GetTradeBySwapAcceptID(ctx, req.GetSwapComplete().GetAcceptId())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	var reply *pb.TradeCompleteReply
	tradeID := currentTrade.ID()
	if err := s.tradeRepository.UpdateTrade(ctx, &tradeID, func(t *trade.Trade) (*trade.Trade, error) {
		psetBase64 := req.GetSwapComplete().GetTransaction()
		opts := wallet.FinalizeAndExtractTransactionOpts{
			PsetBase64: psetBase64,
		}
		txHex, txID, err := wallet.FinalizeAndExtractTransaction(opts)
		if err != nil {
			return nil, err
		}

		if err := t.Complete(psetBase64, txID); err != nil {
			return nil, err
		}

		if _, err := s.explorerService.BroadcastTransaction(txHex); err != nil {
			return nil, err
		}

		reply = &pb.TradeCompleteReply{
			Txid: txID,
		}
		return t, nil
	}); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err := stream.Send(reply); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}
