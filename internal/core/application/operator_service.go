package application

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
)

type OperatorService interface {
	DepositMarket(
		ctx context.Context,
		quoteAsset string,
	) (string, error)
	DepositFeeAccount(
		ctx context.Context,
	) (address string, blindingKey string, err error)
	OpenMarket(
		ctx context.Context,
		baseAsset string,
		quoteAsset string,
	) error
	CloseMarket(
		ctx context.Context,
		baseAsset string,
		quoteAsset string,
	) error
	UpdateMarketFee(
		ctx context.Context,
		req MarketWithFee,
	) (*MarketWithFee, error)
	UpdateMarketPrice(
		ctx context.Context,
		req MarketWithPrice,
	) error
	UpdateMarketStrategy(
		ctx context.Context,
		req MarketStrategy,
	) error
	ListSwaps(
		ctx context.Context,
	) (*pb.ListSwapsReply, error)
}

type operatorService struct {
	marketRepository  domain.MarketRepository
	vaultRepository   domain.VaultRepository
	tradeRepository   domain.TradeRepository
	unspentRepository domain.UnspentRepository
	explorerSvc       explorer.Service
	crawlerSvc        crawler.Service
}

func NewOperatorService(
	marketRepository domain.MarketRepository,
	vaultRepository domain.VaultRepository,
	tradeRepository domain.TradeRepository,
	unspentRepository domain.UnspentRepository,
	explorerSvc explorer.Service,
	crawlerSvc crawler.Service,
) OperatorService {
	return &operatorService{
		marketRepository:  marketRepository,
		vaultRepository:   vaultRepository,
		tradeRepository:   tradeRepository,
		unspentRepository: unspentRepository,
		explorerSvc:       explorerSvc,
		crawlerSvc:        crawlerSvc,
	}
}

func (o *operatorService) DepositMarket(
	ctx context.Context,
	quoteAsset string,
) (string, error) {

	var address string

	var accountIndex int
	_, a, err := o.marketRepository.GetMarketByAsset(
		ctx,
		quoteAsset,
	)
	if err != nil {
		return "", err
	}
	accountIndex = a

	if accountIndex == 0 {
		_, latestAccountIndex, err := o.marketRepository.GetLatestMarket(
			ctx,
		)
		if err != nil {
			return "", err
		}

		accountIndex = latestAccountIndex + 1
		_, err = o.marketRepository.GetOrCreateMarket(ctx, accountIndex)
		if err != nil {
			return "", err
		}
	}

	err = o.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			addr, _, blindingKey, err := v.DeriveNextExternalAddressForAccount(accountIndex)
			if err != nil {
				return nil, err
			}

			address = addr

			o.crawlerSvc.AddObservable(&crawler.AddressObservable{
				AccountIndex: accountIndex,
				Address:      addr,
				BlindingKey:  blindingKey,
			})

			return v, nil
		})
	if err != nil {
		return "", err
	}

	return address, nil
}

func (o *operatorService) DepositFeeAccount(
	ctx context.Context,
) (address string, blindingKey string, err error) {
	err = o.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			addr, _, blindKey, err := v.DeriveNextExternalAddressForAccount(
				domain.FeeAccount,
			)
			if err != nil {
				return nil, err
			}

			address = addr
			blindingKey = hex.EncodeToString(blindKey)

			o.crawlerSvc.AddObservable(&crawler.AddressObservable{
				AccountIndex: domain.FeeAccount,
				Address:      addr,
				BlindingKey:  blindKey,
			})

			return v, nil
		})
	return
}

func (o *operatorService) OpenMarket(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
) error {
	if baseAsset != config.GetString(config.BaseAssetKey) {
		return errors.New("invalid base asset")
	}

	market, marketAccountIndex, err := o.marketRepository.GetMarketByAsset(ctx, quoteAsset)
	if err != nil {
		return err
	}

	var outpoints []domain.OutpointWithAsset
	if market == nil {
		_, marketAccountIndex, err = o.marketRepository.GetLatestMarket(ctx)
		if err != nil {
			return err
		}

		addresses, _, err :=
			o.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, marketAccountIndex)
		if err != nil {
			return err
		}
		unspents, err := o.unspentRepository.GetUnspentsForAddresses(ctx, addresses)
		if err != nil {
			return err
		}

		outpoints = make([]domain.OutpointWithAsset, 0, len(unspents))
		for _, u := range unspents {
			outpoints = append(outpoints, domain.OutpointWithAsset{
				Txid:  u.TxID,
				Vout:  int(u.VOut),
				Asset: u.AssetHash,
			})
		}
	}

	if err := o.marketRepository.UpdateMarket(ctx, marketAccountIndex, func(m *domain.Market) (*domain.Market, error) {
		if m.IsTradable() {
			return m, nil
		}

		if len(outpoints) > 0 {
			if err := m.FundMarket(outpoints); err != nil {
				return nil, err
			}
		}

		if err := m.MakeTradable(); err != nil {
			return nil, err
		}
		return m, nil
	}); err != nil {
		return err
	}

	return nil
}

func (o *operatorService) CloseMarket(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
) error {
	if baseAsset != config.GetString(config.BaseAssetKey) {
		return errors.New("invalid base asset")
	}

	err := o.marketRepository.CloseMarket(
		ctx,
		quoteAsset,
	)
	if err != nil {
		return err
	}

	return nil
}

// UpdateMarketFee changes the Liquidity Provider fee for the given market.
// MUST be expressed as basis point.
// Eg. To change the fee on each swap from 0.25% to 1% you need to pass down 100
// The Market MUST be closed before doing this change.
func (o *operatorService) UpdateMarketFee(
	ctx context.Context,
	req MarketWithFee,
) (*MarketWithFee, error) {

	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return nil, domain.ErrMarketNotExist
	}
	//Checks if market exist
	_, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	//Updates the fee and the fee asset
	err = o.marketRepository.UpdateMarket(
		ctx,
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {
			if err := m.ChangeFee(req.BasisPoint); err != nil {
				return nil, err
			}

			if err := m.ChangeFeeAsset(req.FeeAsset); err != nil {
				return nil, err
			}

			return m, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Ignore errors. If we reached this point it must exists.
	mkt, _ := o.marketRepository.GetOrCreateMarket(
		ctx,
		accountIndex,
	)

	return &MarketWithFee{
		Market: Market{
			BaseAsset:  mkt.BaseAsset,
			QuoteAsset: mkt.QuoteAsset,
		},
		Fee: Fee{
			FeeAsset:   mkt.FeeAsset,
			BasisPoint: mkt.Fee,
		},
	}, nil
}

// UpdateMarketPrice rpc updates the price for the given market
func (o *operatorService) UpdateMarketPrice(
	ctx context.Context,
	req MarketWithPrice,
) error {
	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrMarketNotExist
	}
	//Checks if market exist
	_, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return err
	}

	//Updates the base price and the quote price
	return o.marketRepository.UpdateMarket(
		ctx,
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {

			if err := m.ChangeBasePrice(req.BasePrice); err != nil {
				return nil, err
			}

			if err := m.ChangeQuotePrice(req.QuotePrice); err != nil {
				return nil, err
			}

			return m, nil
		},
	)
}

// UpdateMarketStrategy changes the current market making strategy, either using an automated
// market making formula or a pluggable price feed
func (o *operatorService) UpdateMarketStrategy(
	ctx context.Context,
	req MarketStrategy,
) error {
	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrMarketNotExist
	}
	//Checks if market exist
	_, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return err
	}

	//For now we support only BALANCED or PLUGGABLE (ie. price feed)
	requestStrategy := req.Strategy
	//Updates the strategy
	return o.marketRepository.UpdateMarket(
		ctx,
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {

			switch requestStrategy {

			case domain.StrategyTypePluggable:
				if err := m.MakeStrategyPluggable(); err != nil {
					return nil, err
				}

			case domain.StrategyTypeBalanced:
				if err := m.MakeStrategyBalanced(); err != nil {
					return nil, err
				}

			default:
				return nil, errors.New("strategy not supported")
			}

			return m, nil
		},
	)
}

// ListSwaps returns the list of all swaps processed by the daemon
func (o *operatorService) ListSwaps(
	ctx context.Context,
) (*pb.ListSwapsReply, error) {
	trades, err := o.tradeRepository.GetAllTrades(ctx)
	if err != nil {
		return nil, err
	}

	markets, err := o.getMarketsForTrades(ctx, trades)
	if err != nil {
		return nil, err
	}

	swaps := tradesToSwapInfo(markets, trades)
	return &pb.ListSwapsReply{
		Swaps: swaps,
	}, nil
}

func (o *operatorService) getMarketsForTrades(
	ctx context.Context,
	trades []*domain.Trade,
) (map[string]*domain.Market, error) {
	markets := map[string]*domain.Market{}
	for _, trade := range trades {
		market, _, err := o.marketRepository.GetMarketByAsset(
			ctx,
			trade.MarketQuoteAsset,
		)
		if err != nil {
			return nil, err
		}
		if _, ok := markets[trade.MarketQuoteAsset]; !ok {
			markets[trade.MarketQuoteAsset] = market
		}
	}
	return markets, nil
}

func tradesToSwapInfo(
	markets map[string]*domain.Market,
	trades []*domain.Trade,
) []*pb.SwapInfo {
	info := make([]*pb.SwapInfo, 0, len(trades))
	for _, trade := range trades {
		requestMsg := trade.SwapRequestMessage()
		fee := &pbtypes.Fee{
			Asset:      markets[trade.MarketQuoteAsset].FeeAsset,
			BasisPoint: markets[trade.MarketQuoteAsset].Fee,
		}
		i := &pb.SwapInfo{
			Status:           trade.Status.Code,
			AmountP:          requestMsg.GetAmountP(),
			AssetP:           requestMsg.GetAssetP(),
			AmountR:          requestMsg.GetAmountR(),
			AssetR:           requestMsg.GetAssetR(),
			MarketFee:        fee,
			RequestTimeUnix:  trade.SwapRequestTime(),
			AcceptTimeUnix:   trade.SwapAcceptTime(),
			CompleteTimeUnix: trade.SwapCompleteTime(),
			ExpiryTimeUnix:   trade.SwapExpiryTime(),
		}
		info = append(info, i)
	}
	return info
}
