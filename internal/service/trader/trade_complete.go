package tradeservice

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"

	"github.com/tdex-network/tdex-daemon/internal/domain/trade"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
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
		txHex, err := wallet.FinalizeAndExtractTransaction(opts)
		if err != nil {
			return nil, err
		}

		txID, err := explorer.BroadcastTransaction(txHex)
		if err != nil {
			return nil, err
		}

		blocktime := s.getTransactionBlocktime(txID)

		t.Complete(psetBase64, blocktime, txID)

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

// getTransactionBlocktime is an hekper function that attempts to retrieve the
// blocktime of the blokc that includes the given transaction.
// If it fails to retrieve this information
func (s *Service) getTransactionBlocktime(txID string) (blocktime uint64) {
	status, err := s.explorerService.GetTransactionStatus(txID)
	if err != nil {
		time.Sleep(500 * time.Millisecond)
		status, err = s.explorerService.GetTransactionStatus(txID)
		if err != nil {
			now := time.Now()
			log.Warn(fmt.Sprintf(
				"could not retrieve blocktime for tx '%s', using now (%s) instead.",
				txID,
				now,
			))
			return uint64(now.Unix())
		}
	}
	blocktime = uint64(status["block_time"].(int))
	return
}
