package trade

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	pkgswap "github.com/tdex-network/tdex-daemon/pkg/swap"
)

func (s *Service) TradePreview(
	ctx context.Context, market ports.Market,
	tradeType ports.TradeType, amount uint64, asset string,
) (ports.TradePreview, error) {
	if asset != market.GetBaseAsset() && asset != market.GetQuoteAsset() {
		return nil, fmt.Errorf("asset must match one of those of the market")
	}

	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		log.WithError(err).Debug("failed to fetch market")
		return nil, ErrServiceUnavailable
	}

	if !mkt.IsTradable() {
		return nil, ErrMarketUnavailable
	}

	balance, err := s.wallet.Account().GetBalance(ctx, mkt.Name)
	if err != nil {
		log.WithError(err).Warn("failed to fetch market balance")
		return nil, ErrServiceUnavailable
	}

	return tradePreview(*mkt, balance, tradeType, asset, amount)
}

func (s *Service) TradePropose(
	ctx context.Context, market ports.Market,
	tradeType ports.TradeType, swapRequest ports.SwapRequest,
) (ports.SwapAccept, ports.SwapFail, int64, error) {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		log.WithError(err).Debug("failed to fetch market")
		return nil, nil, -1, ErrServiceUnavailable
	}

	if !mkt.IsTradable() {
		return nil, nil, -1, ErrMarketUnavailable
	}

	balance, err := s.wallet.Account().GetBalance(ctx, mkt.Name)
	if err != nil {
		log.WithError(err).Warn("failed to fetch market balance")
		return nil, nil, -1, ErrServiceUnavailable
	}

	trade := domain.NewTrade()

	defer func() {
		ctx := context.Background()
		if err := s.repoManager.TradeRepository().AddTrade(ctx, trade); err != nil {
			log.WithError(err).Warnf("failed to add trade with id %s", trade.Id)
			return
		}
		log.Debugf("added new trade with id %s", trade.Id)
	}()

	if !isValidTradePrice(*mkt, balance, tradeType, swapRequest, s.priceSlippage) {
		trade.Fail(
			swapRequest.GetId(), pkgswap.ErrCodeBadPricingSwapRequest,
		)
		return nil, trade.SwapFailMessage(), -1, nil
	}

	if ok, _ := trade.Propose(
		swapRequestInfo{swapRequest}.toDomain(),
		mkt.Name, mkt.BaseAsset, mkt.QuoteAsset,
		mkt.PercentageFee, mkt.FixedFee.BaseFee, mkt.FixedFee.QuoteFee, nil,
	); !ok {
		return nil, trade.SwapFailMessage(), -1, nil
	}

	signedPset, selectedUtxos, tradeExpiryTime, err := s.wallet.CompleteSwap(
		mkt.Name, swapRequest,
	)
	if err != nil {
		log.WithError(err).Warn("failed to complete swap request")
		return nil, nil, 0, ErrServiceUnavailable
	}

	unblindedIns := make([]domain.UnblindedInput, 0, len(selectedUtxos))
	for i, u := range selectedUtxos {
		unblindedIns = append(unblindedIns, domain.UnblindedInput{
			Index:         uint32(i),
			Asset:         u.GetAsset(),
			Amount:        u.GetValue(),
			AssetBlinder:  u.GetAssetBlinder(),
			AmountBlinder: u.GetValueBlinder(),
		})
	}
	if ok, _ := trade.Accept(signedPset, unblindedIns, tradeExpiryTime); !ok {
		return nil, trade.SwapFailMessage(), -1, nil
	}

	s.wallet.RegisterHandlerForUtxoEvent(
		s.makeTradeSettledOrExpired(trade.Id, selectedUtxos),
	)

	return swapAcceptInfo{trade.SwapAcceptMessage()}, nil,
		trade.ExpiryTime, nil
}

func (s *Service) TradeComplete(
	ctx context.Context,
	swapComplete ports.SwapComplete, swapFail ports.SwapFail,
) (txidRes string, swapFailRes ports.SwapFail, err error) {
	if swapFail != nil {
		return s.tradeFail(ctx, swapFail)
	}

	trade, err := s.repoManager.TradeRepository().GetTradeBySwapAcceptId(
		ctx, swapComplete.GetAcceptId(),
	)
	if err != nil {
		return "", nil, err
	}
	if trade.IsExpired() {
		return "", nil, fmt.Errorf("trade is expired")
	}

	defer func() {
		if _err := s.repoManager.TradeRepository().UpdateTrade(
			ctx, trade.Id, func(_ *domain.Trade) (*domain.Trade, error) {
				return trade, nil
			},
		); _err != nil {
			err = _err
			txidRes = ""
			swapFailRes = nil
		}
	}()

	ok, _ := trade.Complete(swapComplete.GetTransaction())
	if !ok {
		return "", trade.SwapFailMessage(), nil
	}

	log.Debugf("trade with id %s completed", trade.Id)

	txid, err := s.wallet.Transaction().BroadcastTransaction(ctx, trade.TxHex)
	if err != nil {
		log.WithError(err).Debugf("failed to broadcast trade with id %s", trade.Id)
		trade.Fail(
			trade.SwapAccept.Id, pkgswap.ErrCodeFailedToBroadcast,
		)
		return "", trade.SwapFailMessage(), nil
	}

	log.Debugf("trade with id %s broadcasted: %s", trade.Id, txid)

	return txid, nil, nil
}

func (s *Service) tradeFail(
	ctx context.Context, swapFail ports.SwapFail,
) (string, ports.SwapFail, error) {
	swapAcceptId := swapFail.GetMessageId()
	trade, err := s.repoManager.TradeRepository().GetTradeBySwapAcceptId(
		ctx, swapAcceptId,
	)
	if err != nil {
		return "", nil, err
	}

	trade.Fail(
		swapAcceptId, pkgswap.ErrCodeAborted,
	)

	if err := s.repoManager.TradeRepository().UpdateTrade(
		ctx, trade.Id, func(_ *domain.Trade) (*domain.Trade, error) {
			return trade, nil
		},
	); err != nil {
		log.WithError(err).Warnf("failed to abort trade with id %s", trade.Id)
		return "", nil, ErrServiceUnavailable
	}

	return "", trade.SwapFailMessage(), nil
}

func (s *Service) makeTradeSettledOrExpired(
	tradeId string, tradeUtxos []ports.Utxo,
) func(ports.WalletUtxoNotification) bool {
	contains := func(utxos []ports.Utxo, utxo ports.Utxo) bool {
		for _, u := range utxos {
			if u.GetTxid() == utxo.GetTxid() && u.GetIndex() == utxo.GetIndex() {
				return true
			}
		}
		return false
	}

	return func(notification ports.WalletUtxoNotification) bool {
		eventType := notification.GetEventType()
		utxos := notification.GetUtxos()
		found := false
		for _, u := range tradeUtxos {
			if contains(utxos, u) {
				found = true
				break
			}
		}
		if !found {
			return false
		}

		ctx := context.Background()
		trade, _ := s.repoManager.TradeRepository().GetTradeById(ctx, tradeId)
		var tradeStatus string

		if eventType.IsSpent() {
			settlementTimestamp := time.Now().Unix()
			if status := utxos[0].GetSpentStatus(); status != nil &&
				status.GetBlockInfo() != nil &&
				status.GetBlockInfo().GetTimestamp() > 0 {
				settlementTimestamp = status.GetBlockInfo().GetTimestamp()
			}
			//nolint
			trade.Settle(settlementTimestamp)
			tradeStatus = "settled"
		} else if eventType.IsUnlocked() {
			//nolint
			trade.Expire()
			tradeStatus = "expired"
		}

		//nolint
		s.repoManager.TradeRepository().UpdateTrade(
			ctx, trade.Id, func(_ *domain.Trade) (*domain.Trade, error) {
				return trade, nil
			},
		)
		log.Debugf("trade with id %s %s", trade.Id, tradeStatus)

		if trade.IsSettled() {
			go func() {
				balance, _ := s.wallet.Account().GetBalance(
					context.Background(), trade.MarketName,
				)
				if err := s.pubsub.PublishTradeSettledTopic(
					trade.MarketName, balance, *trade,
				); err != nil {
					log.WithError(err).Warnf(
						"pubsub: failed to publish topic for settled trade with id %s",
						trade.Id,
					)
				} else {
					log.Debugf(
						"pubsub: published topic for settled trade with id %s", trade.Id,
					)
				}
			}()
		}

		return true
	}
}
