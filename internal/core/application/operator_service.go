package application

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
)

// OperatorService defines the methods of the application layer for the operator service.
type OperatorService interface {
	DepositMarket(
		ctx context.Context,
		baseAsset string,
		quoteAsset string,
		numOfAddresses int,
	) ([]string, error)
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
	) ([]SwapInfo, error)
	ListMarketExternalAddresses(
		ctx context.Context,
		req Market,
	) ([]string, error)
	WithdrawMarketFunds(
		ctx context.Context,
		req WithdrawMarketReq,
	) (
		[]byte,
		error,
	)
	FeeAccountBalance(ctx context.Context) (
		int64,
		error,
	)
	ListMarket(
		ctx context.Context,
	) ([]MarketInfo, error)
	GetCollectedMarketFee(
		ctx context.Context,
		market Market,
	) (*ReportMarketFee, error)
}

type operatorService struct {
	marketRepository   domain.MarketRepository
	vaultRepository    domain.VaultRepository
	tradeRepository    domain.TradeRepository
	unspentRepository  domain.UnspentRepository
	explorerSvc        explorer.Service
	blockchainListener BlockchainListener
}

// NewOperatorService is a constructor function for OperatorService.
func NewOperatorService(
	marketRepository domain.MarketRepository,
	vaultRepository domain.VaultRepository,
	tradeRepository domain.TradeRepository,
	unspentRepository domain.UnspentRepository,
	explorerSvc explorer.Service,
	bcListener BlockchainListener,
) OperatorService {
	return &operatorService{
		marketRepository:   marketRepository,
		vaultRepository:    vaultRepository,
		tradeRepository:    tradeRepository,
		unspentRepository:  unspentRepository,
		explorerSvc:        explorerSvc,
		blockchainListener: bcListener,
	}
}

func (o *operatorService) DepositMarket(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
	numOfAddresses int,
) ([]string, error) {
	var accountIndex int

	// First case: the assets are given. If are valid and a market exist we need to derive a new address for that account.
	if len(baseAsset) > 0 && len(quoteAsset) > 0 {
		// check the asset strings
		if err := validateAssetString(baseAsset); err != nil {
			return nil, domain.ErrMarketInvalidBaseAsset
		}

		if err := validateAssetString(quoteAsset); err != nil {
			return nil, domain.ErrMarketInvalidQuoteAsset
		}

		// Checks if base asset is valid
		if baseAsset != config.GetString(config.BaseAssetKey) {
			return nil, domain.ErrMarketInvalidBaseAsset
		}

		//Checks if quote asset exists
		_, accountOfExistentMarket, err := o.marketRepository.GetMarketByAsset(
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
	} else if len(baseAsset) == 0 && len(quoteAsset) == 0 {
		// Second case: base and quote asset are empty. this means we need to create a new market.
		_, latestAccountIndex, err := o.marketRepository.GetLatestMarket(
			ctx,
		)
		if err != nil {
			return nil, err
		}

		nextAccountIndex := latestAccountIndex + 1
		fee := int64(config.GetFloat(config.DefaultFeeKey) * 100)
		if _, err := o.marketRepository.GetOrCreateMarket(
			ctx,
			&domain.Market{AccountIndex: nextAccountIndex, Fee: fee},
		); err != nil {
			return nil, err
		}

		accountIndex = nextAccountIndex
	} else if baseAsset != config.GetString(config.BaseAssetKey) {
		return nil, domain.ErrMarketInvalidBaseAsset
	} else {
		return nil, domain.ErrMarketInvalidQuoteAsset
	}
	if numOfAddresses == 0 {
		numOfAddresses = 1
	}

	list := make([]AddressAndBlindingKey, numOfAddresses, numOfAddresses)
	addresses := make([]string, numOfAddresses, numOfAddresses)
	//Derive an address for that specific market
	if err := o.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			for i := 0; i < numOfAddresses; i++ {
				addr, _, blindingKey, err := v.DeriveNextExternalAddressForAccount(
					accountIndex,
				)
				if err != nil {
					return nil, err
				}

				list[i] = AddressAndBlindingKey{
					Address:     addr,
					BlindingKey: hex.EncodeToString(blindingKey),
				}
				addresses[i] = addr
			}

			return v, nil
		}); err != nil {
		return nil, err
	}

	go o.observeAddressesForAccount(accountIndex, list)

	return addresses, nil
}

func (o *operatorService) DepositFeeAccount(
	ctx context.Context,
	numOfAddresses int,
) ([]AddressAndBlindingKey, error) {
	if numOfAddresses == 0 {
		numOfAddresses = 1
	}

	list := make([]AddressAndBlindingKey, 0, numOfAddresses)
	if err := o.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			for i := 0; i < numOfAddresses; i++ {
				addr, _, blindKey, err := v.DeriveNextExternalAddressForAccount(
					domain.FeeAccount,
				)
				if err != nil {
					return nil, err
				}

				list = append(list, AddressAndBlindingKey{
					Address:     addr,
					BlindingKey: hex.EncodeToString(blindKey),
				})
			}

			return v, nil
		},
	); err != nil {
		return nil, err
	}

	go o.observeAddressesForAccount(domain.FeeAccount, list)

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

	if baseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrMarketInvalidBaseAsset
	}

	// check if the crawler is observing at least one addresse
	if _, err := o.vaultRepository.GetAllDerivedExternalAddressesForAccount(
		ctx,
		domain.FeeAccount,
	); err != nil {
		// TODO: replace this with a variable that must be created at domain level
		if strings.Contains(err.Error(), "account not found") {
			return ErrFeeAccountNotFunded
		}
		return err
	}

	// check if market exists
	market, _, err := o.marketRepository.GetMarketByAsset(
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
	if err := o.marketRepository.OpenMarket(ctx, quoteAsset); err != nil {
		return err
	}

	return nil
}

func (o *operatorService) CloseMarket(
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

	if baseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrMarketInvalidBaseAsset
	}

	err = o.marketRepository.CloseMarket(
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
	// check the asset strings
	if err := validateAssetString(req.BaseAsset); err != nil {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	if err := validateAssetString(req.QuoteAsset); err != nil {
		return nil, domain.ErrMarketInvalidQuoteAsset
	}

	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return nil, ErrMarketNotExist
	}
	//Checks if market exist
	_, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	//Updates the fee and the fee asset
	if err := o.marketRepository.UpdateMarket(
		ctx,
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {
			if err := m.ChangeFee(req.BasisPoint); err != nil {
				return nil, err
			}
			return m, nil
		},
	); err != nil {
		return nil, err
	}

	// Ignore errors. If we reached this point it must exists.
	mkt, _ := o.marketRepository.GetOrCreateMarket(
		ctx,
		&domain.Market{AccountIndex: accountIndex},
	)

	return &MarketWithFee{
		Market: Market{
			BaseAsset:  mkt.BaseAsset,
			QuoteAsset: mkt.QuoteAsset,
		},
		Fee: Fee{
			BasisPoint: mkt.Fee,
		},
	}, nil
}

// UpdateMarketPrice rpc updates the price for the given market
func (o *operatorService) UpdateMarketPrice(
	ctx context.Context,
	req MarketWithPrice,
) error {
	// check the asset strings
	err := validateAssetString(req.BaseAsset)
	if err != nil {
		return domain.ErrMarketInvalidBaseAsset
	}

	err = validateAssetString(req.QuoteAsset)
	if err != nil {
		return domain.ErrMarketInvalidQuoteAsset
	}

	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrMarketInvalidBaseAsset
	}

	// validate the new prices amount
	err = validateAmount(req.Price.BasePrice)
	if err != nil {
		return domain.ErrMarketInvalidBasePrice
	}

	// validate the new prices amount
	err = validateAmount(req.Price.QuotePrice)
	if err != nil {
		return domain.ErrMarketInvalidQuotePrice
	}

	//Checks if market exist
	_, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if accountIndex < 0 {
		return ErrMarketNotExist
	}

	//Updates the base price and the quote price
	return o.marketRepository.UpdatePrices(
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
	// check the asset strings
	err := validateAssetString(req.Market.BaseAsset)
	if err != nil {
		return domain.ErrMarketInvalidBaseAsset
	}

	err = validateAssetString(req.Market.QuoteAsset)
	if err != nil {
		return domain.ErrMarketInvalidQuoteAsset
	}

	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return ErrMarketNotExist
	}
	//Checks if market exist
	_, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return err
	}

	if accountIndex < 0 {
		return ErrMarketNotExist
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
				return nil, ErrUnknownStrategy
			}

			return m, nil
		},
	)
}

// ListSwaps returns the list of all swaps processed by the daemon
func (o *operatorService) ListSwaps(
	ctx context.Context,
) ([]SwapInfo, error) {
	trades, err := o.tradeRepository.GetAllTrades(ctx)
	if err != nil {
		return nil, err
	}

	markets, err := o.getMarketsForTrades(ctx, trades)
	if err != nil {
		return nil, err
	}

	swaps := tradesToSwapInfo(markets, trades)
	return swaps, nil
}

func (o *operatorService) ListMarketExternalAddresses(
	ctx context.Context,
	req Market,
) ([]string, error) {
	// check the asset strings
	err := validateAssetString(req.BaseAsset)
	if err != nil {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	err = validateAssetString(req.QuoteAsset)
	if err != nil {
		return nil, domain.ErrMarketInvalidQuoteAsset
	}

	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	market, _, err := o.marketRepository.GetMarketByAsset(ctx, req.QuoteAsset)
	if err != nil {
		return nil, err
	}

	if market == nil {
		return nil, ErrMarketNotExist
	}

	return o.vaultRepository.GetAllDerivedExternalAddressesForAccount(
		ctx,
		market.AccountIndex,
	)
}

//ListMarket a set of informations about all the markets.
func (o *operatorService) ListMarket(
	ctx context.Context,
) ([]MarketInfo, error) {
	markets, err := o.marketRepository.GetAllMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketInfos := make([]MarketInfo, len(markets))

	for index, market := range markets {
		marketInfos[index] = MarketInfo{
			Market: Market{
				BaseAsset:  market.BaseAsset,
				QuoteAsset: market.QuoteAsset,
			},
			Fee: Fee{
				BasisPoint: market.Fee,
			},
			Tradable:     market.Tradable,
			StrategyType: market.Strategy.Type,
		}
	}

	return marketInfos, nil
}

func (o *operatorService) GetCollectedMarketFee(
	ctx context.Context,
	market Market,
) (*ReportMarketFee, error) {
	m, _, err := o.marketRepository.GetMarketByAsset(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, ErrMarketNotExist
	}

	trades, err := o.tradeRepository.GetCompletedTradesByMarket(
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
) (
	[]byte,
	error,
) {
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	var rawTx []byte

	market, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	if accountIndex == -1 {
		return nil, ErrMarketNotExist
	}

	outs := make([]TxOut, 0)
	if req.BalanceToWithdraw.BaseAmount > 0 {
		outs = append(outs, TxOut{
			Asset:   req.BaseAsset,
			Value:   req.BalanceToWithdraw.BaseAmount,
			Address: req.Address,
		})
	}
	if req.BalanceToWithdraw.QuoteAmount > 0 {
		outs = append(outs, TxOut{
			Asset:   req.QuoteAsset,
			Value:   req.BalanceToWithdraw.QuoteAmount,
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
	if len(marketUnspents) <= 0 {
		return nil, ErrWalletNotFunded
	}

	feeUnspents, err := o.getAllUnspentsForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return nil, err
	}
	if len(feeUnspents) <= 0 {
		return nil, ErrWalletNotFunded
	}

	addressesToObserve := make([]AddressAndBlindingKey, 0)
	err = o.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			mnemonic, err := v.GetMnemonicSafe()
			if err != nil {
				return nil, err
			}
			marketAccount, err := v.AccountByIndex(market.AccountIndex)
			if err != nil {
				return nil, err
			}
			feeAccount, err := v.AccountByIndex(domain.FeeAccount)
			if err != nil {
				return nil, err
			}

			changePathsByAsset := map[string]string{}
			feeChangePathByAsset := map[string]string{}
			for _, asset := range getAssetsOfOutputs(outputs) {
				addr, script, blindkey, err := v.DeriveNextInternalAddressForAccount(
					market.AccountIndex,
				)
				if err != nil {
					return nil, err
				}

				derivationPath := marketAccount.DerivationPathByScript[script]
				changePathsByAsset[asset] = derivationPath
				addressesToObserve = append(
					addressesToObserve,
					AddressAndBlindingKey{
						Address:     addr,
						BlindingKey: hex.EncodeToString(blindkey),
					},
				)
			}

			feeAddress, script, feeBlindkey, err :=
				v.DeriveNextInternalAddressForAccount(domain.FeeAccount)
			if err != nil {
				return nil, err
			}
			feeChangePathByAsset[config.GetNetwork().AssetID] =
				feeAccount.DerivationPathByScript[script]

			addressesToObserve = append(
				addressesToObserve,
				AddressAndBlindingKey{
					Address:     feeAddress,
					BlindingKey: hex.EncodeToString(feeBlindkey),
				},
			)

			txHex, _, err := sendToMany(sendToManyOpts{
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
			})
			if err != nil {
				return nil, err
			}

			if req.Push {
				if _, err := o.explorerSvc.BroadcastTransaction(txHex); err != nil {
					return nil, err
				}
			}

			rawTx, _ = hex.DecodeString(txHex)

			return v, nil
		},
	)
	if err != nil {
		return nil, err
	}

	go o.observeAddressesForAccount(market.AccountIndex, addressesToObserve)

	return rawTx, nil
}

func (o *operatorService) FeeAccountBalance(ctx context.Context) (
	int64,
	error,
) {
	addresses, _, err := o.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return 0, err
	}

	baseAssetAmount, err := o.unspentRepository.GetBalance(
		ctx,
		addresses,
		config.GetString(config.BaseAssetKey),
	)
	if err != nil {
		return -1, err
	}
	return int64(baseAssetAmount), nil
}

func (o *operatorService) getAllUnspentsForAccount(
	ctx context.Context,
	accountIndex int,
) ([]explorer.Utxo, error) {
	addresses, _, err := o.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}

	unspents, err := o.unspentRepository.GetAvailableUnspentsForAddresses(
		ctx,
		addresses,
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

func (o *operatorService) observeAddressesForAccount(accountIndex int, list []AddressAndBlindingKey) {
	for _, l := range list {
		blindkey, _ := hex.DecodeString(l.BlindingKey)
		o.blockchainListener.StartObserveAddress(accountIndex, l.Address, blindkey)
		time.Sleep(200 * time.Millisecond)
	}
}

func (o *operatorService) getMarketsForTrades(
	ctx context.Context,
	trades []*domain.Trade,
) (map[string]*domain.Market, error) {
	markets := map[string]*domain.Market{}
	for _, trade := range trades {
		market, accountIndex, err := o.marketRepository.GetMarketByAsset(
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

func tradesToSwapInfo(
	markets map[string]*domain.Market,
	trades []*domain.Trade,
) []SwapInfo {
	swapInfos := make([]SwapInfo, 0, len(trades))
	for _, trade := range trades {
		requestMsg := trade.SwapRequestMessage()

		fee := Fee{
			BasisPoint: markets[trade.MarketQuoteAsset].Fee,
		}

		newSwapInfo := SwapInfo{
			Status:           int32(trade.Status.Code),
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

		swapInfos = append(swapInfos, newSwapInfo)
	}

	return swapInfos
}
