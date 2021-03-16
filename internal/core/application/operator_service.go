package application

import (
	"context"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/transaction"
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
			return nil, domain.ErrInvalidBaseAsset
		}

		if err := validateAssetString(quoteAsset); err != nil {
			return nil, domain.ErrInvalidQuoteAsset
		}

		// Checks if base asset is valid
		if baseAsset != config.GetString(config.BaseAssetKey) {
			return nil, domain.ErrInvalidBaseAsset
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
			return nil, domain.ErrMarketNotExist
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
		if _, err := o.marketRepository.GetOrCreateMarket(ctx, nextAccountIndex); err != nil {
			return nil, err
		}

		accountIndex = nextAccountIndex
	} else if baseAsset != config.GetString(config.BaseAssetKey) {
		return nil, domain.ErrInvalidBaseAsset
	} else {
		return nil, domain.ErrInvalidQuoteAsset
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
		return domain.ErrInvalidBaseAsset
	}

	err = validateAssetString(quoteAsset)
	if err != nil {
		return domain.ErrInvalidQuoteAsset
	}

	if baseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrInvalidBaseAsset
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
		return domain.ErrMarketNotExist
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
		return domain.ErrInvalidBaseAsset
	}

	err = validateAssetString(quoteAsset)
	if err != nil {
		return domain.ErrInvalidQuoteAsset
	}

	if baseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrInvalidBaseAsset
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
	err := validateAssetString(req.BaseAsset)
	if err != nil {
		return nil, domain.ErrInvalidBaseAsset
	}

	err = validateAssetString(req.QuoteAsset)
	if err != nil {
		return nil, domain.ErrInvalidQuoteAsset
	}

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
	if accountIndex < 0 {
		return nil, domain.ErrMarketNotExist
	}

	//Updates the fee and the fee asset
	err = o.marketRepository.UpdateMarket(
		ctx,
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {
			if err := m.ChangeFee(req.BasisPoint); err != nil {
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
		return domain.ErrInvalidBaseAsset
	}

	err = validateAssetString(req.QuoteAsset)
	if err != nil {
		return domain.ErrInvalidQuoteAsset
	}

	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrInvalidBaseAsset
	}

	// validate the new prices amount
	err = validateAmount(req.Price.BasePrice)
	if err != nil {
		return domain.ErrInvalidBasePrice
	}

	// validate the new prices amount
	err = validateAmount(req.Price.QuotePrice)
	if err != nil {
		return domain.ErrInvalidQuotePrice
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
		return domain.ErrMarketNotExist
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
		return domain.ErrInvalidBaseAsset
	}

	err = validateAssetString(req.Market.QuoteAsset)
	if err != nil {
		return domain.ErrInvalidQuoteAsset
	}

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

	if accountIndex < 0 {
		return domain.ErrMarketNotExist
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
		return nil, domain.ErrInvalidBaseAsset
	}

	err = validateAssetString(req.QuoteAsset)
	if err != nil {
		return nil, domain.ErrInvalidQuoteAsset
	}

	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return nil, domain.ErrInvalidBaseAsset
	}

	market, _, err := o.marketRepository.GetMarketByAsset(ctx, req.QuoteAsset)
	if err != nil {
		return nil, err
	}

	if market == nil {
		return nil, domain.ErrMarketNotExist
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
		return nil, domain.ErrMarketNotExist
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
		return nil, domain.ErrInvalidBaseAsset
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
		return nil, domain.ErrMarketNotExist
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

type depositType string

var (
	marketDeposit depositType = "MarketDeposit"
	feeDeposit    depositType = "FeeDeposit"
)

// ClaimMarketDeposit method add unspents to the market
func (o *operatorService) ClaimMarketDeposit(
	ctx context.Context,
	marketReq Market,
	outpoints []TxOutpoint,
) error {
	addressesPerAccount := make(map[int][]string)
	bkPairs := make([][]byte, 0)

	err := validateMarketRequest(marketReq)
	if err != nil {
		return err
	}

	market, accountIndex, err := o.marketRepository.GetMarketByAsset(ctx,
		marketReq.QuoteAsset)
	if err != nil {
		return err
	}

	// if market exist fetch addresses and blinding keys based on accountIndex
	if market != nil {
		addresses, bk, err := o.vaultRepository.
			GetAllDerivedAddressesAndBlindingKeysForAccount(
				ctx,
				accountIndex,
			)
		if err != nil {
			return err
		}
		addressesPerAccount[accountIndex] = addresses
		bkPairs = bk
	} else {
		// if market doesnt exist fetch addresses and blinding keys of
		// all accounts
		vault, err := o.vaultRepository.GetOrCreateVault(
			ctx,
			nil,
			"",
		)
		if err != nil {
			return err
		}

		markets, err := o.marketRepository.GetNonTradableMarkets(ctx)
		if err != nil {
			return err
		}
		accountIds := make([]int, 0)
		for _, v := range markets {
			accountIds = append(accountIds, v.AccountIndex)
		}

		addressesBlindingKeys := vault.AddressesBlindingKeysGroupByAccount(accountIds)
		for k, v := range addressesBlindingKeys {
			addressesPerAccount[k] = v.Addresses
			bkPairs = v.BlindingKeys

		}
	}

	return o.claimDeposit(
		ctx,
		addressesPerAccount,
		bkPairs,
		outpoints,
		marketDeposit,
	)
}

func validateMarketRequest(marketReq Market) error {
	err := validateAssetString(marketReq.BaseAsset)
	if err != nil {
		return domain.ErrInvalidBaseAsset
	}

	err = validateAssetString(marketReq.QuoteAsset)
	if err != nil {
		return domain.ErrInvalidQuoteAsset
	}

	if marketReq.BaseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrInvalidBaseAsset
	}

	return nil
}

// ClaimFeeDeposit adds unspents to the Fee Account
func (o *operatorService) ClaimFeeDeposit(
	ctx context.Context,
	outpoints []TxOutpoint,
) error {
	addressesPerAccount := make(map[int][]string)

	addresses, bkPairs, err := o.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return err
	}

	addressesPerAccount[domain.FeeAccount] = addresses

	return o.claimDeposit(
		ctx,
		addressesPerAccount,
		bkPairs,
		outpoints,
		feeDeposit,
	)
}

// claimDeposit loops through outpoints and tries to find out to which account
//unspents belongs to, if depositType=="MarketDeposit" it will FundMarket
func (o *operatorService) claimDeposit(
	ctx context.Context,
	addressesPerAccount map[int][]string,
	bkPairs [][]byte,
	outpoints []TxOutpoint,
	depositType depositType,
) error {
	accountIndex := 0

	scriptBlindingKeyPairs := make(map[string]struct {
		accountIndex int
		address      string
		blindingKey  [][]byte
	})

	for accountIndex, addresses := range addressesPerAccount {
		for _, addr := range addresses {
			script, err := address.ToOutputScript(addr)
			if err != nil {
				return err
			}
			scriptBlindingKeyPairs[hex.EncodeToString(script)] = struct {
				accountIndex int
				address      string
				blindingKey  [][]byte
			}{accountIndex: accountIndex, address: addr, blindingKey: bkPairs}
		}
	}

	unspents := make([]domain.Unspent, 0)
	for _, v := range outpoints {
		confirmed, err := o.explorerSvc.IsTransactionConfirmed(v.Hash)
		if err != nil {
			return err
		}
		if !confirmed {
			return ErrTxNotConfirmed
		}

		txHex, err := o.explorerSvc.GetTransactionHex(v.Hash)
		if err != nil {
			return err
		}
		tx, err := transaction.NewTxFromHex(txHex)
		if err != nil {
			return err
		}

		if len(tx.Outputs) <= v.Index {
			return fmt.Errorf(
				"tx: %v, doesnt have outpoint at index: %v",
				v.Hash,
				v.Index,
			)
		}
		script := hex.EncodeToString(tx.Outputs[v.Index].Script)

		//here we match outpoint script with relevant address and
		//blinding key, stored in a daemon,
		//so we can unblind unspents and eventually store them
		if val, ok := scriptBlindingKeyPairs[script]; ok {
			//in case of marketDeposit we need to update base and quote assets
			accountIndex = val.accountIndex
			utxo, err := o.explorerSvc.GetUnspents(
				val.address,
				val.blindingKey,
			)
			if err != nil {
				return err
			}

			for _, v := range utxo {
				u := domain.Unspent{
					TxID:            v.Hash(),
					VOut:            v.Index(),
					Value:           v.Value(),
					AssetHash:       v.Asset(),
					ValueCommitment: v.ValueCommitment(),
					AssetCommitment: v.AssetCommitment(),
					ValueBlinder:    v.ValueBlinder(),
					AssetBlinder:    v.AssetBlinder(),
					ScriptPubKey:    v.Script(),
					Nonce:           v.Nonce(),
					RangeProof:      v.RangeProof(),
					SurjectionProof: v.SurjectionProof(),
					Confirmed:       v.IsConfirmed(),
					Address:         val.address,
				}
				unspents = append(unspents, u)
			}

		} else {
			return fmt.Errorf(
				"outpoint: %v at index: %v not owned by market",
				v.Hash,
				v.Index,
			)
		}
	}

	if err := o.unspentRepository.AddUnspents(ctx, unspents); err != nil {
		return err
	}

	if depositType == marketDeposit {
		if err := o.fundMarket(
			ctx,
			accountIndex,
			unspents,
		); err != nil {
			return err
		}
	} else {
		if err := o.checkFeeAccountBalance(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (o *operatorService) fundMarket(
	ctx context.Context,
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
	return o.marketRepository.UpdateMarket(
		ctx,
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {
			if err := m.FundMarket(outpoints); err != nil {
				return nil, err
			}

			return m, nil
		},
	)
}

func (o *operatorService) checkFeeAccountBalance(ctx context.Context) error {
	addresses, _, err := o.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return err
	}

	feeAccountBalance, err := o.unspentRepository.GetBalance(
		ctx,
		addresses,
		config.GetString(config.BaseAssetKey),
	)
	if err != nil {
		return err
	}

	if feeAccountBalance < uint64(config.GetInt(config.FeeAccountBalanceThresholdKey)) {
		log.Warn(
			"fee account balance for account index too low. Trades for markets won't be " +
				"served properly. Fund the fee account as soon as possible",
		)
	} else {
		log.Info("fee account funded. Trades can be served")
	}

	return nil
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
			return nil, domain.ErrMarketNotExist
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
