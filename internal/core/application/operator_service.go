package application

import (
	"context"
	"encoding/hex"
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/transaction"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
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
	marketRepository           domain.MarketRepository
	vaultRepository            domain.VaultRepository
	tradeRepository            domain.TradeRepository
	unspentRepository          domain.UnspentRepository
	explorerSvc                explorer.Service
	blockchainListener         BlockchainListener
	marketBaseAsset            string
	marketFee                  int64
	network                    *network.Network
	feeAccountBalanceThreshold uint64
}

// NewOperatorService is a constructor function for OperatorService.
func NewOperatorService(
	marketRepository domain.MarketRepository,
	vaultRepository domain.VaultRepository,
	tradeRepository domain.TradeRepository,
	unspentRepository domain.UnspentRepository,
	explorerSvc explorer.Service,
	bcListener BlockchainListener,
	marketBaseAsset string,
	marketFee int64,
	net *network.Network,
	feeAccountBalanceThreshold uint64,
) OperatorService {
	return &operatorService{
		marketRepository:           marketRepository,
		vaultRepository:            vaultRepository,
		tradeRepository:            tradeRepository,
		unspentRepository:          unspentRepository,
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
		if baseAsset != o.marketBaseAsset {
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
		fee := o.marketFee
		if _, err := o.marketRepository.GetOrCreateMarket(
			ctx,
			&domain.Market{AccountIndex: nextAccountIndex, Fee: fee},
		); err != nil {
			return nil, err
		}

		accountIndex = nextAccountIndex
	} else if baseAsset != o.marketBaseAsset {
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
		func(v *domain.Vault) (*domain.Vault, error) {
			for i := 0; i < numOfAddresses; i++ {
				info, err := v.DeriveNextExternalAddressForAccount(accountIndex)
				if err != nil {
					return nil, err
				}
				list[i] = AddressAndBlindingKey{
					Address:     info.Address,
					BlindingKey: hex.EncodeToString(info.BlindingKey),
				}
				addresses[i] = info.Address
			}

			return v, nil
		}); err != nil {
		return nil, err
	}

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
		func(v *domain.Vault) (*domain.Vault, error) {
			for i := 0; i < numOfAddresses; i++ {
				info, err := v.DeriveNextExternalAddressForAccount(domain.FeeAccount)
				if err != nil {
					return nil, err
				}

				list = append(list, AddressAndBlindingKey{
					Address:     info.Address,
					BlindingKey: hex.EncodeToString(info.BlindingKey),
				})
			}

			return v, nil
		},
	); err != nil {
		return nil, err
	}

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
	if _, err := o.vaultRepository.GetAllDerivedExternalAddressesInfoForAccount(
		ctx,
		domain.FeeAccount,
	); err != nil {
		if err == domain.ErrVaultAccountNotFound {
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

	if baseAsset != o.marketBaseAsset {
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
	if req.BaseAsset != o.marketBaseAsset {
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
	if req.BaseAsset != o.marketBaseAsset {
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
	if req.BaseAsset != o.marketBaseAsset {
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

	if req.BaseAsset != o.marketBaseAsset {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	market, _, err := o.marketRepository.GetMarketByAsset(ctx, req.QuoteAsset)
	if err != nil {
		return nil, err
	}

	if market == nil {
		return nil, ErrMarketNotExist
	}

	allInfo, err := o.vaultRepository.GetAllDerivedExternalAddressesInfoForAccount(
		ctx,
		market.AccountIndex,
	)
	if err != nil {
		return nil, err
	}

	return allInfo.Addresses(), nil
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
) ([]byte, error) {
	if req.BaseAsset != o.marketBaseAsset {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	market, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	if accountIndex < 0 {
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

	var txHex string

	if err := o.vaultRepository.UpdateVault(
		ctx,
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
				info, err := v.DeriveNextInternalAddressForAccount(market.AccountIndex)
				if err != nil {
					return nil, err
				}

				derivationPath := marketAccount.DerivationPathByScript[info.Script]
				changePathsByAsset[asset] = derivationPath
			}

			feeInfo, err := v.DeriveNextInternalAddressForAccount(domain.FeeAccount)
			if err != nil {
				return nil, err
			}
			feeChangePathByAsset[o.network.AssetID] =
				feeAccount.DerivationPathByScript[feeInfo.Script]

			_txHex, err := sendToMany(sendToManyOpts{
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

			if req.Push {
				if _, err := o.explorerSvc.BroadcastTransaction(_txHex); err != nil {
					return nil, err
				}
			}

			txHex = _txHex

			return v, nil
		},
	); err != nil {
		return nil, err
	}

	go extractUnspentsFromTxAndUpdateUtxoSet(
		o.unspentRepository,
		o.vaultRepository,
		o.network,
		txHex,
		market.AccountIndex,
	)

	rawTx, _ := hex.DecodeString(txHex)
	return rawTx, nil
}

func (o *operatorService) FeeAccountBalance(ctx context.Context) (
	int64,
	error,
) {
	info, err := o.vaultRepository.GetAllDerivedAddressesInfoForAccount(
		ctx,
		domain.FeeAccount,
	)
	if err != nil {
		return 0, err
	}

	baseAssetAmount, err := o.unspentRepository.GetBalance(
		ctx,
		info.Addresses(),
		o.marketBaseAsset,
	)
	if err != nil {
		return -1, err
	}
	return int64(baseAssetAmount), nil
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

	market, accountIndex, err := o.marketRepository.GetMarketByAsset(ctx,
		marketReq.QuoteAsset)
	if err != nil {
		return err
	}

	infoPerAccount := make(map[int]domain.AddressesInfo)

	// if market exist,s fetch all receiving addresses and relative blinding keys
	// for the known market.
	if market != nil {
		info, err := o.vaultRepository.GetAllDerivedExternalAddressesInfoForAccount(
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
		vault, err := o.vaultRepository.GetOrCreateVault(ctx, nil, "", nil)
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
	info, err := o.vaultRepository.GetAllDerivedExternalAddressesInfoForAccount(
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

func (o *operatorService) getNonFundedMarkets(ctx context.Context) ([]domain.Market, error) {
	markets, err := o.marketRepository.GetAllMarkets(ctx)
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

		// TODO: Add a GetTransaction method to explorer interface to prevent
		// direct usage of go-elements here.
		txHex, err := o.explorerSvc.GetTransactionHex(v.Hash)
		if err != nil {
			return err
		}
		tx, _ := transaction.NewTxFromHex(txHex)

		if len(tx.Outputs) <= v.Index {
			return ErrInvalidOutpoint
		}

		txOut := tx.Outputs[v.Index]
		script := hex.EncodeToString(txOut.Script)
		if info, ok := infoByScript[script]; ok {
			counter[info.AccountIndex]++

			unconfidential, ok := transactionutil.UnblindOutput(
				txOut,
				info.BlindingKey,
			)
			if !ok {
				return errors.New("unable to unblind output")
			}

			unspents[i] = domain.Unspent{
				TxID:            tx.TxHash().String(),
				VOut:            uint32(v.Index),
				Value:           unconfidential.Value,
				AssetHash:       unconfidential.AssetHash,
				ValueCommitment: bufferutil.CommitmentFromBytes(txOut.Value),
				AssetCommitment: bufferutil.CommitmentFromBytes(txOut.Value),
				ValueBlinder:    unconfidential.ValueBlinder,
				AssetBlinder:    unconfidential.AssetBlinder,
				ScriptPubKey:    txOut.Script,
				Nonce:           txOut.Nonce,
				RangeProof:      txOut.RangeProof,
				SurjectionProof: txOut.SurjectionProof,
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
				addUnspents(o.unspentRepository, unspents)
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
	return o.marketRepository.UpdateMarket(
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
	feeAccountBalance, err := o.unspentRepository.GetBalance(
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
	info, err := o.vaultRepository.GetAllDerivedAddressesInfoForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}

	unspents, err := o.unspentRepository.GetAvailableUnspentsForAddresses(
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
			RequestTimeUnix:  trade.SwapRequest.Timestamp,
			AcceptTimeUnix:   trade.SwapAccept.Timestamp,
			CompleteTimeUnix: trade.SwapComplete.Timestamp,
			ExpiryTimeUnix:   trade.ExpiryTime,
		}

		swapInfos = append(swapInfos, newSwapInfo)
	}

	return swapInfos
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
