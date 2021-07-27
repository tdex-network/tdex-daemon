package application

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/network"
)

const (
	marketDeposit = iota
	feeDeposit
)

// OperatorService defines the methods of the application layer for the operator service.
type OperatorService interface {
	DepositMarket(
		ctx context.Context,
		baseAsset string,
		quoteAsset string,
		numOfAddresses int,
	) ([]AddressAndBlindingKey, error)
	DepositFeeAccount(
		ctx context.Context,
		numOfAddresses int,
	) ([]AddressAndBlindingKey, error)
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
	UpdateMarketPercentageFee(
		ctx context.Context,
		req MarketWithFee,
	) (*MarketWithFee, error)
	UpdateMarketFixedFee(
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
	ListTrades(
		ctx context.Context,
	) ([]TradeInfo, error)
	ListMarketExternalAddresses(
		ctx context.Context,
		req Market,
	) ([]string, error)
	WithdrawMarketFunds(
		ctx context.Context,
		req WithdrawMarketReq,
	) ([]byte, error)
	FeeAccountBalance(ctx context.Context) (int64, error)
	ClaimMarketDeposit(
		ctx context.Context,
		market Market,
		outpoints []TxOutpoint,
	) error
	ClaimFeeDeposit(
		ctx context.Context,
		outpoints []TxOutpoint,
	) error
	ListMarket(
		ctx context.Context,
	) ([]MarketInfo, error)
	GetCollectedMarketFee(
		ctx context.Context,
		market Market,
	) (*ReportMarketFee, error)
	ListUtxos(ctx context.Context) (map[uint64]UtxoInfoList, error)
	ReloadUtxos(ctx context.Context) error
	DropMarket(ctx context.Context, accountIndex int) error
	AddWebhook(ctx context.Context, hook Webhook) (string, error)
	RemoveWebhook(ctx context.Context, id string) error
	ListWebhooks(ctx context.Context, actionType int) ([]WebhookInfo, error)
}

type operatorService struct {
	repoManager                ports.RepoManager
	explorerSvc                explorer.Service
	blockchainListener         BlockchainListener
	marketBaseAsset            string
	marketFee                  int64
	network                    *network.Network
	feeAccountBalanceThreshold uint64
}

// NewOperatorService is a constructor function for OperatorService.
func NewOperatorService(
	repoManager ports.RepoManager,
	explorerSvc explorer.Service,
	bcListener BlockchainListener,
	marketBaseAsset string,
	marketFee int64,
	net *network.Network,
	feeAccountBalanceThreshold uint64,
) OperatorService {
	return &operatorService{
		repoManager:                repoManager,
		explorerSvc:                explorerSvc,
		blockchainListener:         bcListener,
		marketBaseAsset:            marketBaseAsset,
		marketFee:                  marketFee,
		network:                    net,
		feeAccountBalanceThreshold: feeAccountBalanceThreshold,
	}
}

func (o *operatorService) DepositMarket(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
	numOfAddresses int,
) ([]AddressAndBlindingKey, error) {
	var accountIndex int

	// First case: the assets are given. If are valid and a market exist we need to derive a new address for that account.
	if len(baseAsset) > 0 || len(quoteAsset) > 0 {
		if err := validateAssetString(baseAsset); err != nil {
			return nil, domain.ErrMarketInvalidBaseAsset
		}

		if err := validateAssetString(quoteAsset); err != nil {
			return nil, domain.ErrMarketInvalidQuoteAsset
		}

		if baseAsset != o.marketBaseAsset {
			return nil, domain.ErrMarketInvalidBaseAsset
		}

		_, accountOfExistentMarket, err := o.repoManager.MarketRepository().GetMarketByAsset(
			ctx,
			quoteAsset,
		)
		if err != nil {
			return nil, err
		}
		if accountOfExistentMarket == -1 {
			return nil, ErrMarketNotExist
		}

		accountIndex = accountOfExistentMarket
	} else {
		// Second case: base and quote asset are empty. this means we need to create a new market.
		_, latestAccountIndex, err := o.repoManager.MarketRepository().GetLatestMarket(
			ctx,
		)
		if err != nil {
			return nil, err
		}

		nextAccountIndex := latestAccountIndex + 1
		accountIndex = nextAccountIndex
	}
	if numOfAddresses == 0 {
		numOfAddresses = 1
	}

	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		return nil, err
	}

	list := make([]AddressAndBlindingKey, numOfAddresses, numOfAddresses)
	for i := 0; i < numOfAddresses; i++ {
		info, err := vault.DeriveNextExternalAddressForAccount(accountIndex)
		if err != nil {
			return nil, err
		}
		list[i] = AddressAndBlindingKey{
			Address:     info.Address,
			BlindingKey: hex.EncodeToString(info.BlindingKey),
		}
	}

	go func() {
		if _, err := o.repoManager.RunTransaction(ctx, false, func(ctx context.Context) (interface{}, error) {
			// this makes sure that the market is created if it needs to. Otherwise,
			// this does not commit any change to the marekt repo.
			if _, err := o.repoManager.MarketRepository().GetOrCreateMarket(
				ctx,
				&domain.Market{
					AccountIndex: accountIndex,
					Fee:          o.marketFee,
				},
			); err != nil {
				return nil, err
			}

			if err := o.repoManager.VaultRepository().UpdateVault(
				ctx,
				func(_ *domain.Vault) (*domain.Vault, error) {
					return vault, nil
				}); err != nil {
				return nil, err
			}

			return nil, nil
		}); err != nil {
			log.WithError(err).Warn("unable to persist changes")
		}
	}()

	return list, nil
}

func (o *operatorService) DepositFeeAccount(
	ctx context.Context,
	numOfAddresses int,
) ([]AddressAndBlindingKey, error) {
	if numOfAddresses == 0 {
		numOfAddresses = 1
	}

	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		return nil, err
	}

	list := make([]AddressAndBlindingKey, numOfAddresses, numOfAddresses)
	for i := 0; i < numOfAddresses; i++ {
		info, err := vault.DeriveNextExternalAddressForAccount(domain.FeeAccount)
		if err != nil {
			return nil, err
		}

		list[i] = AddressAndBlindingKey{
			Address:     info.Address,
			BlindingKey: hex.EncodeToString(info.BlindingKey),
		}
	}

	go func() {
		if err := o.repoManager.VaultRepository().UpdateVault(
			ctx,
			func(_ *domain.Vault) (*domain.Vault, error) {
				return vault, nil
			},
		); err != nil {
			log.WithError(err).Warn("unable to update vault")
		}
	}()

	return list, nil
}

func (o *operatorService) OpenMarket(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
) error {
	// check the asset strings
	err := validateAssetString(baseAsset)
	if err != nil {
		return domain.ErrMarketInvalidBaseAsset
	}

	err = validateAssetString(quoteAsset)
	if err != nil {
		return domain.ErrMarketInvalidQuoteAsset
	}

	if baseAsset != o.marketBaseAsset {
		return domain.ErrMarketInvalidBaseAsset
	}

	// check if some addresses of the fee account have been derived already
	if _, err := o.repoManager.VaultRepository().GetAllDerivedExternalAddressesInfoForAccount(
		ctx,
		domain.FeeAccount,
	); err != nil {
		if err == domain.ErrVaultAccountNotFound {
			return ErrFeeAccountNotFunded
		}
		return err
	}

	// check if market exists
	market, _, err := o.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		quoteAsset,
	)
	if err != nil {
		return err
	}

	if market == nil {
		return ErrMarketNotExist
	}

	// open the market
	if err := o.repoManager.MarketRepository().OpenMarket(ctx, quoteAsset); err != nil {
		return err
	}

	return nil
}

func (o *operatorService) CloseMarket(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
) error {
	err := validateAssetString(baseAsset)
	if err != nil {
		return domain.ErrMarketInvalidBaseAsset
	}

	err = validateAssetString(quoteAsset)
	if err != nil {
		return domain.ErrMarketInvalidQuoteAsset
	}

	if baseAsset != o.marketBaseAsset {
		return domain.ErrMarketInvalidBaseAsset
	}

	err = o.repoManager.MarketRepository().CloseMarket(
		ctx,
		quoteAsset,
	)
	if err != nil {
		return err
	}

	return nil
}

// UpdateMarketPercentageFee changes the Liquidity Provider fee for the given market.
// MUST be expressed as basis point.
// Eg. To change the fee on each swap from 0.25% to 1% you need to pass down 100
// The Market MUST be closed before doing this change.
func (o *operatorService) UpdateMarketPercentageFee(
	ctx context.Context,
	req MarketWithFee,
) (*MarketWithFee, error) {
	if err := validateAssetString(req.BaseAsset); err != nil {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	if err := validateAssetString(req.QuoteAsset); err != nil {
		return nil, domain.ErrMarketInvalidQuoteAsset
	}

	if req.BaseAsset != o.marketBaseAsset {
		return nil, ErrMarketNotExist
	}

	mkt, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	if err := mkt.ChangeFeeBasisPoint(req.BasisPoint); err != nil {
		return nil, err
	}

	if err := o.repoManager.MarketRepository().UpdateMarket(
		ctx,
		accountIndex,
		func(_ *domain.Market) (*domain.Market, error) {
			return mkt, nil
		},
	); err != nil {
		return nil, err
	}

	return &MarketWithFee{
		Market: Market{
			BaseAsset:  mkt.BaseAsset,
			QuoteAsset: mkt.QuoteAsset,
		},
		Fee: Fee{
			BasisPoint:    mkt.Fee,
			FixedBaseFee:  mkt.FixedFee.BaseFee,
			FixedQuoteFee: mkt.FixedFee.QuoteFee,
		},
	}, nil
}

// UpdateMarketFixedFee changes the Liquidity Provider fee for the given market.
// Values for both assets MUST be expressed as satoshis.
func (o *operatorService) UpdateMarketFixedFee(
	ctx context.Context,
	req MarketWithFee,
) (*MarketWithFee, error) {
	if err := validateAssetString(req.BaseAsset); err != nil {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	if err := validateAssetString(req.QuoteAsset); err != nil {
		return nil, domain.ErrMarketInvalidQuoteAsset
	}

	if req.BaseAsset != o.marketBaseAsset {
		return nil, ErrMarketNotExist
	}

	mkt, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	if err := mkt.ChangeFixedFee(req.FixedBaseFee, req.FixedQuoteFee); err != nil {
		return nil, err
	}

	if err := o.repoManager.MarketRepository().UpdateMarket(
		ctx,
		accountIndex,
		func(_ *domain.Market) (*domain.Market, error) {
			return mkt, nil
		},
	); err != nil {
		return nil, err
	}

	return &MarketWithFee{
		Market: Market{
			BaseAsset:  mkt.BaseAsset,
			QuoteAsset: mkt.QuoteAsset,
		},
		Fee: Fee{
			BasisPoint:    mkt.Fee,
			FixedBaseFee:  mkt.FixedFee.BaseFee,
			FixedQuoteFee: mkt.FixedFee.QuoteFee,
		},
	}, nil
}

// UpdateMarketPrice rpc updates the price for the given market
func (o *operatorService) UpdateMarketPrice(
	ctx context.Context,
	req MarketWithPrice,
) error {
	if err := validateAssetString(req.BaseAsset); err != nil {
		return domain.ErrMarketInvalidBaseAsset
	}

	if err := validateAssetString(req.QuoteAsset); err != nil {
		return domain.ErrMarketInvalidQuoteAsset
	}

	if req.BaseAsset != o.marketBaseAsset {
		return domain.ErrMarketInvalidBaseAsset
	}

	if err := validateAmount(req.Price.BasePrice); err != nil {
		return domain.ErrMarketInvalidBasePrice
	}

	if err := validateAmount(req.Price.QuotePrice); err != nil {
		return domain.ErrMarketInvalidQuotePrice
	}

	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if accountIndex < 0 {
		return ErrMarketNotExist
	}

	// Updates the base price and the quote price
	return o.repoManager.MarketRepository().UpdatePrices(
		ctx,
		accountIndex,
		domain.Prices{
			BasePrice:  req.Price.BasePrice,
			QuotePrice: req.Price.QuotePrice,
		},
	)
}

// UpdateMarketStrategy changes the current market making strategy,
// either using an automated market making formula or a pluggable price feed
func (o *operatorService) UpdateMarketStrategy(
	ctx context.Context,
	req MarketStrategy,
) error {
	if err := validateAssetString(req.Market.BaseAsset); err != nil {
		return domain.ErrMarketInvalidBaseAsset
	}

	if err := validateAssetString(req.Market.QuoteAsset); err != nil {
		return domain.ErrMarketInvalidQuoteAsset
	}

	if req.BaseAsset != o.marketBaseAsset {
		return ErrMarketNotExist
	}

	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return err
	}

	if accountIndex < 0 {
		return ErrMarketNotExist
	}

	requestStrategy := req.Strategy

	return o.repoManager.MarketRepository().UpdateMarket(
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
				return nil, ErrUnknownStrategy
			}

			return m, nil
		},
	)
}

// ListTrades returns the list of all trads processed by the daemon
func (o *operatorService) ListTrades(
	ctx context.Context,
) ([]TradeInfo, error) {
	trades, err := o.repoManager.TradeRepository().GetAllTrades(ctx)
	if err != nil {
		return nil, err
	}

	return tradesToTradeInfo(trades, o.marketBaseAsset, o.network.Name), nil
}

func (o *operatorService) ListMarketExternalAddresses(
	ctx context.Context,
	req Market,
) ([]string, error) {
	if err := validateAssetString(req.BaseAsset); err != nil {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	if err := validateAssetString(req.QuoteAsset); err != nil {
		return nil, domain.ErrMarketInvalidQuoteAsset
	}

	if req.BaseAsset != o.marketBaseAsset {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAsset(ctx, req.QuoteAsset)
	if err != nil {
		return nil, err
	}

	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	allInfo, err := o.repoManager.VaultRepository().GetAllDerivedExternalAddressesInfoForAccount(
		ctx,
		accountIndex,
	)
	if err != nil {
		return nil, err
	}

	return allInfo.Addresses(), nil
}

// ListMarket a set of informations about all the markets.
func (o *operatorService) ListMarket(
	ctx context.Context,
) ([]MarketInfo, error) {
	markets, err := o.repoManager.MarketRepository().GetAllMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketInfo := make([]MarketInfo, 0, len(markets))
	for _, market := range markets {
		marketInfo = append(marketInfo, MarketInfo{
			AccountIndex: uint64(market.AccountIndex),
			Market: Market{
				BaseAsset:  market.BaseAsset,
				QuoteAsset: market.QuoteAsset,
			},
			Tradable:     market.Tradable,
			StrategyType: market.Strategy.Type,
			Price:        market.Price,
			Fee: Fee{
				BasisPoint:    market.Fee,
				FixedBaseFee:  market.FixedFee.BaseFee,
				FixedQuoteFee: market.FixedFee.QuoteFee,
			},
		})
	}

	return marketInfo, nil
}

func (o *operatorService) GetCollectedMarketFee(
	ctx context.Context,
	market Market,
) (*ReportMarketFee, error) {
	m, _, err := o.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, ErrMarketNotExist
	}

	trades, err := o.repoManager.TradeRepository().GetCompletedTradesByMarket(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	fees := make([]FeeInfo, 0, len(trades))
	total := make(map[string]int64)
	for _, trade := range trades {
		feeBasisPoint := trade.MarketFee
		swapRequest := trade.SwapRequestMessage()
		feeAsset := swapRequest.GetAssetP()
		amountP := swapRequest.GetAmountP()
		_, feeAmount := mathutil.LessFee(amountP, uint64(feeBasisPoint))

		marketPrice := trade.MarketPrice.QuotePrice
		if feeAsset == m.BaseAsset {
			marketPrice = trade.MarketPrice.BasePrice
		}

		fees = append(fees, FeeInfo{
			TradeID:     trade.ID.String(),
			BasisPoint:  feeBasisPoint,
			Asset:       feeAsset,
			Amount:      feeAmount,
			MarketPrice: marketPrice,
		})

		total[feeAsset] += int64(feeAmount)
	}

	return &ReportMarketFee{
		CollectedFees:              fees,
		TotalCollectedFeesPerAsset: total,
	}, nil
}

func (o *operatorService) WithdrawMarketFunds(
	ctx context.Context,
	req WithdrawMarketReq,
) ([]byte, error) {
	if req.BaseAsset != o.marketBaseAsset {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	market, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		return nil, err
	}

	info, err := vault.AllDerivedAddressesInfoForAccount(accountIndex)
	if err != nil {
		return nil, err
	}
	addresses := info.Addresses()

	baseBalance, err := o.repoManager.UnspentRepository().GetUnlockedBalance(ctx, addresses, req.BaseAsset)
	if err != nil {
		return nil, err
	}
	if baseBalance <= 0 {
		return nil, ErrMarketBaseBalanceTooLow
	}
	if req.BalanceToWithdraw.BaseAmount > baseBalance {
		return nil, ErrWithdrawBaseAmountTooBig
	}

	quoteBalance, err := o.repoManager.UnspentRepository().GetUnlockedBalance(ctx, addresses, req.QuoteAsset)
	if err != nil {
		return nil, err
	}
	if quoteBalance <= 0 {
		return nil, ErrMarketQuoteBalanceTooLow
	}
	if req.BalanceToWithdraw.QuoteAmount > quoteBalance {
		return nil, ErrWithdrawQuoteAmountTooBig
	}

	outs := make([]TxOut, 0)
	if req.BalanceToWithdraw.BaseAmount > 0 {
		outs = append(outs, TxOut{
			Asset:   req.BaseAsset,
			Value:   int64(req.BalanceToWithdraw.BaseAmount),
			Address: req.Address,
		})
	}
	if req.BalanceToWithdraw.QuoteAmount > 0 {
		outs = append(outs, TxOut{
			Asset:   req.QuoteAsset,
			Value:   int64(req.BalanceToWithdraw.QuoteAmount),
			Address: req.Address,
		})
	}

	outputs, outputsBlindingKeys, err := parseRequestOutputs(outs)
	if err != nil {
		return nil, err
	}

	marketUnspents, err := o.getAllUnspentsForAccount(ctx, market.AccountIndex)
	if err != nil {
		return nil, err
	}

	feeUnspents, err := o.getAllUnspentsForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return nil, err
	}

	mnemonic, err := vault.GetMnemonicSafe()
	if err != nil {
		return nil, err
	}

	marketAccount, err := vault.AccountByIndex(market.AccountIndex)
	if err != nil {
		return nil, err
	}
	feeAccount, err := vault.AccountByIndex(domain.FeeAccount)
	if err != nil {
		return nil, err
	}

	changePathsByAsset := map[string]string{}
	feeChangePathByAsset := map[string]string{}
	for _, asset := range getAssetsOfOutputs(outputs) {
		info, err := vault.DeriveNextInternalAddressForAccount(accountIndex)
		if err != nil {
			return nil, err
		}

		derivationPath := marketAccount.DerivationPathByScript[info.Script]
		changePathsByAsset[asset] = derivationPath
	}

	feeInfo, err := vault.DeriveNextInternalAddressForAccount(domain.FeeAccount)
	if err != nil {
		return nil, err
	}
	feeChangePathByAsset[o.network.AssetID] =
		feeAccount.DerivationPathByScript[feeInfo.Script]

	txHex, err := sendToMany(sendToManyOpts{
		mnemonic:              mnemonic,
		unspents:              marketUnspents,
		feeUnspents:           feeUnspents,
		outputs:               outputs,
		outputsBlindingKeys:   outputsBlindingKeys,
		changePathsByAsset:    changePathsByAsset,
		feeChangePathByAsset:  feeChangePathByAsset,
		inputPathsByScript:    marketAccount.DerivationPathByScript,
		feeInputPathsByScript: feeAccount.DerivationPathByScript,
		milliSatPerByte:       int(req.MillisatPerByte),
		network:               o.network,
	})
	if err != nil {
		return nil, err
	}

	var txid string
	if req.Push {
		txid, err = o.explorerSvc.BroadcastTransaction(txHex)
		if err != nil {
			return nil, err
		}
	}

	if err := o.repoManager.VaultRepository().UpdateVault(
		ctx,
		func(_ *domain.Vault) (*domain.Vault, error) {
			return vault, nil
		},
	); err != nil {
		return nil, err
	}

	go extractUnspentsFromTxAndUpdateUtxoSet(
		o.repoManager.UnspentRepository(),
		o.repoManager.VaultRepository(),
		o.network,
		txHex,
		market.AccountIndex,
	)

	// Publish message for topic AccountWithdraw to pubsub service.
	if svc := o.blockchainListener.PubSubService(); svc != nil {
		go func() {
			if err != nil {
				log.WithError(err).Warn("an error occured while retrieving quote balance")
				return
			}

			payload := map[string]interface{}{
				"market": map[string]string{
					"base_asset":  req.BaseAsset,
					"quote_asset": req.QuoteAsset,
				},
				"amount_withdraw": map[string]interface{}{
					"base_amount":  req.BalanceToWithdraw.BaseAmount,
					"quote_amount": req.BalanceToWithdraw.QuoteAmount,
				},
				"receiving_address": req.Address,
				"txid":              txid,
				"balance": map[string]uint64{
					"base_balance":  baseBalance - req.BalanceToWithdraw.BaseAmount,
					"quote_balance": quoteBalance - req.BalanceToWithdraw.QuoteAmount,
				},
			}
			message, _ := json.Marshal(payload)
			topics := svc.TopicsByCode()
			topic := topics[AccountWithdraw]
			if err := svc.Publish(topic.Label(), string(message)); err != nil {
				log.WithError(err).Warnf(
					"an error occured while publishing message for topic %s",
					topic.Label(),
				)
			}
		}()
	}

	rawTx, _ := hex.DecodeString(txHex)
	return rawTx, nil
}

func (o *operatorService) FeeAccountBalance(ctx context.Context) (
	int64,
	error,
) {
	info, err := o.repoManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(
		ctx,
		domain.FeeAccount,
	)
	if err != nil {
		return 0, err
	}

	baseAssetAmount, err := o.repoManager.UnspentRepository().GetBalance(
		ctx,
		info.Addresses(),
		o.marketBaseAsset,
	)
	if err != nil {
		return -1, err
	}
	return int64(baseAssetAmount), nil
}

func (o *operatorService) ListUtxos(
	ctx context.Context,
) (map[uint64]UtxoInfoList, error) {
	utxoInfoPerAccount := make(map[uint64]UtxoInfoList)

	unspents := o.repoManager.UnspentRepository().GetAllUnspents(ctx)
	for _, u := range unspents {
		_, accountIndex, err := o.repoManager.VaultRepository().GetAccountByAddress(
			ctx,
			u.Address,
		)
		if err != nil {
			return nil, err
		}

		utxoInfo := utxoInfoPerAccount[uint64(accountIndex)]
		if u.Spent {
			utxoInfo.Spents = appendUtxoInfo(utxoInfo.Spents, u)
		} else if u.Locked {
			utxoInfo.Locks = appendUtxoInfo(utxoInfo.Locks, u)
		} else {
			utxoInfo.Unspents = appendUtxoInfo(utxoInfo.Unspents, u)
		}
		utxoInfoPerAccount[uint64(accountIndex)] = utxoInfo
	}

	return utxoInfoPerAccount, nil
}

// ReloadUtxos triggers reloading of unspents for stored addresses from blockchain
func (o *operatorService) ReloadUtxos(ctx context.Context) error {
	//get all addresses
	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(
		ctx, nil, "", nil,
	)
	if err != nil {
		return err
	}

	addressesInfo := vault.AllDerivedAddressesInfo()
	_, err = fetchAndAddUnspents(
		o.explorerSvc,
		o.repoManager.UnspentRepository(),
		o.blockchainListener,
		addressesInfo,
	)
	return err
}

// ClaimMarketDeposit method add unspents to the market
func (o *operatorService) ClaimMarketDeposit(
	ctx context.Context,
	marketReq Market,
	outpoints []TxOutpoint,
) error {
	if err := validateMarketRequest(marketReq, o.marketBaseAsset); err != nil {
		return err
	}

	market, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAsset(
		ctx, marketReq.QuoteAsset,
	)
	if err != nil {
		return err
	}

	infoPerAccount := make(map[int]domain.AddressesInfo)

	// if market exists, fetch all receiving addresses and relative blinding keys
	// for the known market.
	if market != nil {
		info, err := o.repoManager.VaultRepository().GetAllDerivedExternalAddressesInfoForAccount(
			ctx,
			accountIndex,
		)
		if err != nil {
			return err
		}
		infoPerAccount[accountIndex] = info
	} else {
		// otherwise fetch all non funded market and for each of them fetch all
		// receiving addresses and relative blinding keys
		markets, err := o.getNonFundedMarkets(ctx)
		if err != nil {
			return err
		}
		if len(markets) <= 0 {
			return ErrMissingNonFundedMarkets
		}
		vault, err := o.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
		if err != nil {
			return err
		}

		for _, m := range markets {
			info, err := vault.AllDerivedExternalAddressesInfoForAccount(m.AccountIndex)
			if err != nil {
				return err
			}
			infoPerAccount[m.AccountIndex] = info
		}
	}

	return o.claimDeposit(ctx, infoPerAccount, outpoints, marketDeposit)
}

// ClaimFeeDeposit adds unspents to the Fee Account
func (o *operatorService) ClaimFeeDeposit(
	ctx context.Context,
	outpoints []TxOutpoint,
) error {
	info, err := o.repoManager.VaultRepository().GetAllDerivedExternalAddressesInfoForAccount(
		ctx,
		domain.FeeAccount,
	)
	if err != nil {
		return err
	}

	infoPerAccount := make(map[int]domain.AddressesInfo)
	infoPerAccount[domain.FeeAccount] = info

	return o.claimDeposit(ctx, infoPerAccount, outpoints, feeDeposit)
}

func (o *operatorService) DropMarket(
	ctx context.Context,
	accountIndex int,
) error {
	return o.repoManager.MarketRepository().DeleteMarket(ctx, accountIndex)
}

func (o *operatorService) AddWebhook(_ context.Context, hook Webhook) (string, error) {
	if o.blockchainListener.PubSubService() == nil {
		return "", ErrPubSubServiceNotInitialized
	}

	topics := o.blockchainListener.PubSubService().TopicsByCode()
	topic, ok := topics[hook.ActionType]
	if !ok {
		return "", ErrInvalidActionType
	}

	return o.blockchainListener.PubSubService().Subscribe(
		topic.Label(), hook.Endpoint, hook.Secret,
	)
}

func (o *operatorService) RemoveWebhook(_ context.Context, hookID string) error {
	if o.blockchainListener.PubSubService() == nil {
		return ErrPubSubServiceNotInitialized
	}
	return o.blockchainListener.PubSubService().Unsubscribe("", hookID)
}

func (o *operatorService) ListWebhooks(_ context.Context, actionType int) ([]WebhookInfo, error) {
	pubsubSvc := o.blockchainListener.PubSubService()
	if pubsubSvc == nil {
		return nil, ErrPubSubServiceNotInitialized
	}

	topics := pubsubSvc.TopicsByCode()
	topic, ok := topics[actionType]
	if !ok {
		return nil, ErrInvalidActionType
	}

	subs := pubsubSvc.ListSubscriptionsForTopic(topic.Label())
	hooks := make([]WebhookInfo, 0, len(subs))
	for _, s := range subs {
		hooks = append(hooks, WebhookInfo{
			Id:         s.Id(),
			ActionType: s.Topic().Code(),
			Endpoint:   s.NotifyAt(),
			IsSecured:  s.IsSecured(),
		})
	}
	return hooks, nil
}

func (o *operatorService) getNonFundedMarkets(ctx context.Context) ([]domain.Market, error) {
	markets, err := o.repoManager.MarketRepository().GetAllMarkets(ctx)
	if err != nil {
		return nil, err
	}

	nonFundedMakrets := make([]domain.Market, 0)
	for _, m := range markets {
		if !m.IsFunded() {
			nonFundedMakrets = append(nonFundedMakrets, m)
		}
	}
	return nonFundedMakrets, nil
}

func (o *operatorService) claimDeposit(
	ctx context.Context,
	infoPerAccount map[int]domain.AddressesInfo,
	outpoints []TxOutpoint,
	depositType int,
) error {
	// group all addresses info by script
	infoByScript := make(map[string]domain.AddressInfo)
	for _, info := range infoPerAccount {
		for s, i := range groupAddressesInfoByScript(info) {
			infoByScript[s] = i
		}
	}

	// for each outpoint retrieve the raw tx and output. If the output script
	// exists in infoByScript, increment the counter of the related account and
	// unblind the raw confidential output.
	// Since all outpoints MUST be funds of the same account, at the end of the
	// loop there MUST be only one counter matching the length of the give
	// outpoints.
	counter := make(map[int]int)
	unspents := make([]domain.Unspent, len(outpoints), len(outpoints))
	for i, v := range outpoints {
		confirmed, err := o.explorerSvc.IsTransactionConfirmed(v.Hash)
		if err != nil {
			return err
		}
		if !confirmed {
			return ErrTxNotConfirmed
		}

		tx, err := o.explorerSvc.GetTransaction(v.Hash)
		if err != nil {
			return err
		}

		if len(tx.Outputs()) <= v.Index {
			return ErrInvalidOutpoint
		}

		txOut := tx.Outputs()[v.Index]
		script := hex.EncodeToString(txOut.Script)
		if info, ok := infoByScript[script]; ok {
			counter[info.AccountIndex]++

			unconfidential, ok := BlinderManager.UnblindOutput(
				txOut,
				info.BlindingKey,
			)
			if !ok {
				return errors.New("unable to unblind output")
			}

			unspents[i] = domain.Unspent{
				TxID:            v.Hash,
				VOut:            uint32(v.Index),
				Value:           unconfidential.Value,
				AssetHash:       unconfidential.AssetHash,
				ValueCommitment: bufferutil.CommitmentFromBytes(txOut.Value),
				AssetCommitment: bufferutil.CommitmentFromBytes(txOut.Asset),
				ValueBlinder:    unconfidential.ValueBlinder,
				AssetBlinder:    unconfidential.AssetBlinder,
				ScriptPubKey:    txOut.Script,
				Nonce:           txOut.Nonce,
				RangeProof:      make([]byte, 1),
				SurjectionProof: make([]byte, 1),
				Address:         info.Address,
				Confirmed:       true,
			}
		}
	}

	for accountIndex, count := range counter {
		if count == len(outpoints) {
			if depositType == marketDeposit {
				if err := o.fundMarket(accountIndex, unspents); err != nil {
					return err
				}
				log.Infof("funded market with account %d", accountIndex)
			}

			go func() {
				addUnspentsAsync(o.repoManager.UnspentRepository(), unspents)
				if depositType == feeDeposit {
					if err := o.checkAccountBalance(infoPerAccount[accountIndex]); err != nil {
						log.Warn(err)
						return
					}
					log.Info("fee account funded. Trades can be served")
				}
			}()

			return nil
		}
	}

	return ErrInvalidOutpoints
}

func (o *operatorService) fundMarket(
	accountIndex int,
	unspents []domain.Unspent,
) error {
	outpoints := make([]domain.OutpointWithAsset, 0, len(unspents))
	for _, u := range unspents {
		outpoints = append(outpoints, domain.OutpointWithAsset{
			Txid:  u.TxID,
			Vout:  int(u.VOut),
			Asset: u.AssetHash,
		})
	}

	// Update the market trying to funding attaching the newly found quote asset.
	return o.repoManager.MarketRepository().UpdateMarket(
		context.Background(),
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {
			if err := m.FundMarket(outpoints, o.marketBaseAsset); err != nil {
				return nil, err
			}

			return m, nil
		},
	)
}

func (o *operatorService) checkAccountBalance(accountInfo domain.AddressesInfo) error {
	feeAccountBalance, err := o.repoManager.UnspentRepository().GetBalance(
		context.Background(),
		accountInfo.Addresses(),
		o.marketBaseAsset,
	)
	if err != nil {
		return err
	}

	if feeAccountBalance < o.feeAccountBalanceThreshold {
		return errors.New(
			"fee account balance for account index too low. Trades for markets " +
				"won't be served properly. Fund the fee account as soon as possible",
		)
	}

	return nil
}

func (o *operatorService) getAllUnspentsForAccount(
	ctx context.Context,
	accountIndex int,
) ([]explorer.Utxo, error) {
	info, err := o.repoManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}

	unspents, err := o.repoManager.UnspentRepository().GetAvailableUnspentsForAddresses(
		ctx,
		info.Addresses(),
	)
	if err != nil {
		return nil, err
	}

	utxos := make([]explorer.Utxo, 0, len(unspents))
	for _, u := range unspents {
		utxos = append(utxos, u.ToUtxo())
	}
	return utxos, nil
}

func (o *operatorService) getMarketsForTrades(
	ctx context.Context,
	trades []*domain.Trade,
) (map[string]*domain.Market, error) {
	markets := map[string]*domain.Market{}
	for _, trade := range trades {
		market, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAsset(
			ctx,
			trade.MarketQuoteAsset,
		)
		if err != nil {
			return nil, err
		}
		if accountIndex < 0 {
			return nil, ErrMarketNotExist
		}
		if _, ok := markets[trade.MarketQuoteAsset]; !ok {
			markets[trade.MarketQuoteAsset] = market
		}
	}
	return markets, nil
}

func tradesToTradeInfo(trades []*domain.Trade, marketBaseAsset, network string) []TradeInfo {
	tradeInfo := make([]TradeInfo, 0, len(trades))
	chInfo := make(chan TradeInfo)
	wg := &sync.WaitGroup{}
	wg.Add(len(trades))

	go func() {
		wg.Wait()
		close(chInfo)
	}()

	for _, trade := range trades {
		go tradeToTradeInfo(trade, marketBaseAsset, network, chInfo, wg)
	}

	for info := range chInfo {
		tradeInfo = append(tradeInfo, info)
	}

	// sort by request timestamp
	sort.SliceStable(tradeInfo, func(i, j int) bool {
		return tradeInfo[i].RequestTimeUnix < tradeInfo[j].RequestTimeUnix
	})

	return tradeInfo
}

func tradeToTradeInfo(
	trade *domain.Trade,
	marketBaseAsset, net string,
	chInfo chan TradeInfo,
	wg *sync.WaitGroup,
) {
	if wg != nil {
		defer wg.Done()
	}

	if trade.IsEmpty() {
		return
	}

	info := TradeInfo{
		ID:     trade.ID.String(),
		Status: trade.Status,
		MarketWithFee: MarketWithFee{
			Market{
				BaseAsset:  marketBaseAsset,
				QuoteAsset: trade.MarketQuoteAsset,
			},
			Fee{
				BasisPoint: trade.MarketFee,
			},
		},
		Price:            Price(trade.MarketPrice),
		RequestTimeUnix:  trade.SwapRequest.Timestamp,
		AcceptTimeUnix:   trade.SwapAccept.Timestamp,
		CompleteTimeUnix: trade.SwapComplete.Timestamp,
		SettleTimeUnix:   trade.SettlementTime,
		ExpiryTimeUnix:   trade.ExpiryTime,
	}

	if req := trade.SwapRequestMessage(); req != nil {
		info.SwapInfo = SwapInfo{
			AssetP:  req.GetAssetP(),
			AmountP: req.GetAmountP(),
			AssetR:  req.GetAssetR(),
			AmountR: req.GetAmountR(),
		}
	}

	if fail := trade.SwapFailMessage(); fail != nil {
		info.SwapFailInfo = SwapFailInfo{
			Code:    int(fail.GetFailureCode()),
			Message: fail.GetFailureMessage(),
		}
	}

	if trade.IsSettled() {
		_, outBlindingData, _ := TransactionManager.ExtractBlindingData(
			trade.PsetBase64,
			nil, trade.SwapAcceptMessage().GetOutputBlindingKey(),
		)

		var blinded string
		for _, data := range outBlindingData {
			blinded += fmt.Sprintf(
				"%d,%s,%s,%s,",
				data.Amount, data.Asset,
				hex.EncodeToString(elementsutil.ReverseBytes(data.AmountBlinder)),
				hex.EncodeToString(elementsutil.ReverseBytes(data.AssetBlinder)),
			)
		}
		// remove trailing comma
		blinded = strings.Trim(blinded, ",")

		baseURL := "https://blockstream.info/liquid/tx"
		if net == network.Regtest.Name {
			baseURL = "http://localhost:3001/tx"
		}
		info.TxURL = fmt.Sprintf("%s/%s#blinded=%s", baseURL, trade.TxID, blinded)
	}

	chInfo <- info
	return
}

func validateMarketRequest(marketReq Market, baseAsset string) error {
	if err := validateAssetString(marketReq.BaseAsset); err != nil {
		return err
	}

	if err := validateAssetString(marketReq.QuoteAsset); err != nil {
		return err
	}

	// Checks if base asset is valid
	if marketReq.BaseAsset != baseAsset {
		return domain.ErrMarketInvalidBaseAsset
	}

	return nil
}

func groupAddressesInfoByScript(info domain.AddressesInfo) map[string]domain.AddressInfo {
	group := make(map[string]domain.AddressInfo)
	for _, i := range info {
		group[i.Script] = i
	}
	return group
}

func appendUtxoInfo(list []UtxoInfo, unspent domain.Unspent) []UtxoInfo {
	return append(list, UtxoInfo{
		Outpoint: &TxOutpoint{
			Hash:  unspent.TxID,
			Index: int(unspent.VOut),
		},
		Value: unspent.Value,
		Asset: unspent.AssetHash,
	})
}
