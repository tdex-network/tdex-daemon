package application

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/circuitbreaker"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pkgswap "github.com/tdex-network/tdex-daemon/pkg/swap"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/pset"
)

type TradeService interface {
	GetTradableMarkets(ctx context.Context) ([]MarketWithFee, error)
	GetMarketPrice(
		ctx context.Context,
		market Market,
		tradeType int,
		amount uint64,
		asset string,
	) (*PriceWithFee, error)
	TradePropose(
		ctx context.Context,
		market Market,
		tradeType int,
		swapRequest domain.SwapRequest,
	) (domain.SwapAccept, domain.SwapFail, uint64, error)
	TradeComplete(
		ctx context.Context,
		swapComplete domain.SwapComplete,
		swapFail domain.SwapFail,
	) (string, domain.SwapFail, error)
	GetMarketBalance(
		ctx context.Context,
		market Market,
	) (*BalanceWithFee, error)
}

type tradeService struct {
	repoManager                ports.RepoManager
	explorerSvc                explorer.Service
	blockchainListener         BlockchainListener
	marketBaseAsset            string
	expiryDuration             time.Duration
	priceSlippage              decimal.Decimal
	network                    *network.Network
	feeAccountBalanceThreshold uint64

	lock *sync.Mutex
}

func NewTradeService(
	repoManager ports.RepoManager,
	explorerSvc explorer.Service,
	bcListener BlockchainListener,
	marketBaseAsset string,
	expiryDuration time.Duration,
	priceSlippage decimal.Decimal,
	net *network.Network,
	feeAccountBalanceThreshold uint64,
) TradeService {
	return newTradeService(
		repoManager,
		explorerSvc,
		bcListener,
		marketBaseAsset,
		expiryDuration,
		priceSlippage,
		net,
		feeAccountBalanceThreshold,
	)
}

func newTradeService(
	repoManager ports.RepoManager,
	explorerSvc explorer.Service,
	bcListener BlockchainListener,
	marketBaseAsset string,
	expiryDuration time.Duration,
	priceSlippage decimal.Decimal,
	net *network.Network,
	feeAccountBalanceThreshold uint64,
) *tradeService {
	return &tradeService{
		repoManager:                repoManager,
		explorerSvc:                explorerSvc,
		blockchainListener:         bcListener,
		marketBaseAsset:            marketBaseAsset,
		expiryDuration:             expiryDuration,
		priceSlippage:              priceSlippage,
		network:                    net,
		feeAccountBalanceThreshold: feeAccountBalanceThreshold,
		lock:                       &sync.Mutex{},
	}
}

func (t *tradeService) GetTradableMarkets(ctx context.Context) (
	[]MarketWithFee,
	error,
) {
	vault, err := t.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		log.Debugf("error while retrieving vault: %s", err)
		return nil, ErrServiceUnavailable
	}
	if vault.IsLocked() {
		log.Debug("vault is locked")
		return nil, ErrServiceUnavailable
	}

	tradableMarkets, err := t.repoManager.MarketRepository().GetTradableMarkets(ctx)
	if err != nil {
		log.Debugf("error while retrieving markets: %s", err)
		return nil, ErrServiceUnavailable
	}

	marketsWithFee := make([]MarketWithFee, 0, len(tradableMarkets))
	for _, mkt := range tradableMarkets {
		marketsWithFee = append(marketsWithFee, MarketWithFee{
			Market: Market{
				BaseAsset:  mkt.BaseAsset,
				QuoteAsset: mkt.QuoteAsset,
			},
			Fee: Fee{
				BasisPoint:    mkt.Fee,
				FixedBaseFee:  mkt.FixedFee.BaseFee,
				FixedQuoteFee: mkt.FixedFee.QuoteFee,
			},
		})
	}

	return marketsWithFee, nil
}

func (t *tradeService) GetMarketPrice(
	ctx context.Context,
	market Market,
	tradeType int,
	amount uint64,
	asset string,
) (*PriceWithFee, error) {
	if err := market.Validate(); err != nil {
		return nil, err
	}

	if err := validateAssetString(asset); err != nil {
		return nil, errors.New("invalid asset")
	}
	if asset != market.BaseAsset && asset != market.QuoteAsset {
		return nil, errors.New("asset must match one of those of the market")
	}

	mkt, mktAccountIndex, err := t.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		log.Debugf("error while retrieving market: %s", err)
		return nil, ErrServiceUnavailable
	}
	if mktAccountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	if !mkt.IsTradable() {
		return nil, domain.ErrMarketIsClosed
	}

	_, unspents, err := t.getInfoAndUnspentsForAccount(ctx, mktAccountIndex)
	if err != nil {
		log.Debugf("error while retrieving unspents: %s", err)
		return nil, ErrServiceUnavailable
	}

	return previewForMarket(unspents, mkt, tradeType, amount, asset)
}

func (t *tradeService) GetMarketBalance(
	ctx context.Context,
	market Market,
) (*BalanceWithFee, error) {
	if err := market.Validate(); err != nil {
		return nil, err
	}

	m, accountIndex, err := t.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		log.WithError(err).Debug("error while retrieving market")
		return nil, ErrServiceUnavailable
	}
	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	balance, err := getUnlockedBalanceForMarket(t.repoManager, ctx, m)
	if err != nil {
		log.WithError(err).Debug("error while retrieving balance")
		return nil, ErrServiceUnavailable
	}

	return &BalanceWithFee{
		Balance: *balance,
		Fee: Fee{
			BasisPoint:    m.Fee,
			FixedBaseFee:  m.FixedFee.BaseFee,
			FixedQuoteFee: m.FixedFee.QuoteFee,
		},
	}, nil
}

func (t *tradeService) TradePropose(
	ctx context.Context,
	market Market,
	tradeType int,
	swapRequest domain.SwapRequest,
) (domain.SwapAccept, domain.SwapFail, uint64, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if err := market.Validate(); err != nil {
		return nil, nil, 0, err
	}

	vault, err := t.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		log.Debugf("error while retrieving vault: %s", err)
		return nil, nil, 0, ErrServiceUnavailable
	}
	if vault.IsLocked() {
		log.Debug("vault is locked")
		return nil, nil, 0, ErrServiceUnavailable
	}

	mkt, marketAccountIndex, err := t.repoManager.MarketRepository().GetMarketByAsset(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		log.Debugf("error while retrieving market: %s", err)
		return nil, nil, 0, ErrServiceUnavailable
	}
	if marketAccountIndex < 0 {
		return nil, nil, 0, ErrMarketNotExist
	}

	// Eventually, check fee and market accounts to notify for low balances.
	defer func() {
		go checkFeeAndMarketBalances(
			t.repoManager, t.blockchainListener.PubSubService(),
			ctx, mkt, t.network.AssetID, t.feeAccountBalanceThreshold,
		)
	}()

	// get all unspents for market account (both as []domain.Unspents and as
	// []explorer.Utxo)along with private blinding keys and signing derivation
	// paths for respectively unblinding and signing them later
	marketInfo, marketUnspents, err :=
		t.getInfoAndUnspentsForAccount(ctx, marketAccountIndex)
	if err != nil {
		log.Debugf("error while retrieving market account addresses and unspents: %s", err)
		return nil, nil, 0, ErrServiceUnavailable
	}

	// Check we got at least one
	if len(marketUnspents) <= 0 {
		return nil, nil, 0, ErrMarketNotFunded
	}

	// ... and the same for fee account (we'll need to top-up fees)
	feeInfo, feeUnspents, err :=
		t.getInfoAndUnspentsForAccount(ctx, domain.FeeAccount)
	if err != nil {
		log.Debugf("error while retrieving fee account addresses and unspents: %s", err)
		return nil, nil, 0, ErrServiceUnavailable
	}
	// Check we got at least one
	if len(feeUnspents) <= 0 {
		return nil, nil, 0, ErrFeeAccountNotFunded
	}

	cb := circuitbreaker.NewCircuitBreaker()
	iBlocktime, err := cb.Execute(func() (interface{}, error) {
		return t.explorerSvc.GetBlockHeight()
	})
	if err != nil {
		log.Debugf("error while retrieving block height: %s", err)
		return nil, nil, 0, ErrServiceUnavailable
	}
	blocktime := iBlocktime.(int)

	// parse swap proposal and possibly accept
	var swapAccept domain.SwapAccept
	var swapFail domain.SwapFail
	var swapExpiryTime uint64

	var fillProposalResult *FillProposalResult
	var outInfo *domain.AddressInfo
	var changeInfo *domain.AddressInfo
	var feeChangeInfo *domain.AddressInfo
	var mnemonic []string

	trade := domain.NewTrade()
	if ok, _ := trade.Propose(
		swapRequest,
		market.QuoteAsset,
		mkt.Fee,
		mkt.FixedFee.BaseFee,
		mkt.FixedFee.QuoteFee,
		uint64(blocktime),
		nil,
	); !ok {
		swapFail = trade.SwapFailMessage()
		goto end
	}

	if !isValidTradePrice(swapRequest, tradeType, mkt, marketUnspents, t.priceSlippage) {
		trade.Fail(
			swapRequest.GetId(),
			int(pkgswap.ErrCodeInvalidSwapRequest),
			"bad pricing",
		)
		swapFail = trade.SwapFailMessage()
		goto end
	}

	// derive output and change address for market, and change address for fee account
	outInfo, _ = vault.DeriveNextExternalAddressForAccount(marketAccountIndex)
	changeInfo, _ = vault.DeriveNextInternalAddressForAccount(marketAccountIndex)
	feeChangeInfo, _ = vault.DeriveNextInternalAddressForAccount(domain.FeeAccount)

	mnemonic, _ = vault.GetMnemonicSafe()
	fillProposalResult, err = TradeManager.FillProposal(FillProposalOpts{
		Mnemonic:      mnemonic,
		SwapRequest:   swapRequest,
		MarketUtxos:   marketUnspents.ToUtxos(),
		FeeUtxos:      feeUnspents.ToUtxos(),
		MarketInfo:    marketInfo,
		FeeInfo:       feeInfo,
		OutputInfo:    *outInfo,
		ChangeInfo:    *changeInfo,
		FeeChangeInfo: *feeChangeInfo,
		Network:       t.network,
	})
	if err != nil {
		trade.Fail(
			swapRequest.GetId(),
			int(pkgswap.ErrCodeRejectedSwapRequest),
			"internal error",
		)
		log.WithError(err).Infof("trade with id %s rejected", trade.ID)
		swapFail = trade.SwapFailMessage()
		goto end
	}

	if ok, _ := trade.Accept(
		fillProposalResult.PsetBase64,
		fillProposalResult.InputBlindingKeys,
		fillProposalResult.OutputBlindingKeys,
		uint64(t.expiryDuration.Seconds()),
	); !ok {
		swapFail = trade.SwapFailMessage()
		log.Infof("trade with id %s rejected", trade.ID)
		goto end
	}

	swapAccept = trade.SwapAcceptMessage()
	swapExpiryTime = trade.ExpiryTime

end:
	if swapAccept != nil {
		log.Infof("trade with id %s accepted", trade.ID)
		selectedUnspentKeys := getUnspentKeys(fillProposalResult.SelectedUnspents)

		lockedUnspents, _ := t.repoManager.UnspentRepository().LockUnspents(
			ctx,
			selectedUnspentKeys,
			trade.ID,
		)
		log.Debugf("locked %d unspents", lockedUnspents)

		// set timer for trade expiration
		go func() {
			// admit a tollerance of 1 minute past the expiration time.
			time.Sleep(t.expiryDuration + time.Minute)
			t.checkTradeExpiration(trade.ID, fillProposalResult.SelectedUnspents)
		}()

		t.blockchainListener.StartObserveOutpoints(
			fillProposalResult.SelectedUnspents, trade.ID.String(),
		)
	}

	go func() {
		if _, err := t.repoManager.RunTransaction(
			context.Background(),
			false,
			func(ctx context.Context) (interface{}, error) {
				if _, err := t.repoManager.TradeRepository().GetOrCreateTrade(ctx, &trade.ID); err != nil {
					return nil, err
				}
				if err := t.repoManager.TradeRepository().UpdateTrade(ctx, &trade.ID, func(_ *domain.Trade) (*domain.Trade, error) {
					return trade, nil
				}); err != nil {
					return nil, err
				}

				if swapAccept != nil {
					if err := t.repoManager.VaultRepository().UpdateVault(ctx, func(_ *domain.Vault) (*domain.Vault, error) {
						return vault, nil
					}); err != nil {
						return nil, err
					}
				}
				return nil, nil
			},
		); err != nil {
			log.WithError(err).Warn("unable to persist changes after trade is accepted")
		}
	}()

	return swapAccept, swapFail, swapExpiryTime, nil
}

// TradeComplete is the domain controller for the TradeComplete RPC
func (t *tradeService) TradeComplete(
	ctx context.Context,
	swapComplete domain.SwapComplete,
	swapFail domain.SwapFail,
) (string, domain.SwapFail, error) {
	if swapFail != nil {
		swapFailMsg, err := t.tradeFail(ctx, swapFail)
		if err != nil {
			log.Debugf("error while aborting trade: %s", err)
			return "", nil, ErrServiceUnavailable
		}
		return "", swapFailMsg, nil
	}

	return t.tradeComplete(ctx, swapComplete)
}

func (t *tradeService) tradeComplete(
	ctx context.Context,
	swapComplete domain.SwapComplete,
) (txID string, swapFail domain.SwapFail, err error) {
	swapID := swapComplete.GetAcceptId()
	trade, err := t.repoManager.TradeRepository().GetTradeBySwapAcceptID(ctx, swapID)
	if err != nil {
		return
	}
	if trade == nil {
		err = fmt.Errorf("trade with swap id %s not found", swapComplete.GetAcceptId())
		return
	}

	tx := swapComplete.GetTransaction()

	res, err := trade.Complete(tx)
	if err != nil {
		return
	}

	defer func() {
		if err := t.repoManager.TradeRepository().UpdateTrade(
			ctx,
			&trade.ID,
			func(_ *domain.Trade) (*domain.Trade, error) { return trade, nil },
		); err != nil {
			log.WithError(err).Warn("unable to persist changes to trade with id: ", trade.ID)
		}
	}()

	if !res.OK {
		swapFail = trade.SwapFailMessage()
		return
	}
	log.Infof("trade with id %s completed", trade.ID)

	cb := circuitbreaker.NewCircuitBreaker()
	iTxid, err := cb.Execute(func() (interface{}, error) {
		return t.explorerSvc.BroadcastTransaction(res.TxHex)
	})
	if err != nil {
		trade.Fail(
			swapID, int(pkgswap.ErrCodeFailedToComplete), fmt.Sprintf(
				"failed to broadcast tx: %s", err.Error(),
			))
		log.WithError(err).WithField("hex", res.TxHex).Warn("unable to broadcast trade tx")
		return
	}
	txID = iTxid.(string)
	trade.TxID = txID

	log.Infof("trade with id %s broadcasted: %s", trade.ID, txID)

	go func() {
		_, accountIndex, _ := t.repoManager.MarketRepository().GetMarketByAsset(
			ctx,
			trade.MarketQuoteAsset,
		)
		extractUnspentsFromTxAndUpdateUtxoSet(
			t.repoManager.UnspentRepository(),
			t.repoManager.VaultRepository(),
			t.network,
			res.TxHex,
			accountIndex,
		)
	}()

	return
}

func (t *tradeService) tradeFail(
	ctx context.Context,
	swapFail domain.SwapFail,
) (domain.SwapFail, error) {
	swapID := swapFail.GetMessageId()
	trade, err := t.repoManager.TradeRepository().GetTradeBySwapAcceptID(ctx, swapID)
	if err != nil {
		return nil, err
	}

	tradeID := trade.ID
	if err := t.repoManager.TradeRepository().UpdateTrade(
		ctx,
		&tradeID,
		func(trade *domain.Trade) (*domain.Trade, error) {
			trade.Fail(
				swapID,
				int(pkgswap.ErrCodeFailedToComplete),
				"set failed by counter-party",
			)
			return trade, nil
		},
	); err != nil {
		return nil, err
	}

	go t.unlockUnspentsForTrade(trade)
	go t.blockchainListener.StopObserveTx(trade.TxID)

	return swapFail, nil
}

func (t *tradeService) getInfoAndUnspentsForAccount(
	ctx context.Context,
	account int,
) (domain.AddressesInfo, Unspents, error) {
	info, err := t.repoManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(ctx, account)
	if err != nil {
		return nil, nil, err
	}
	derivedAddresses := info.Addresses()

	unspents, err := t.repoManager.UnspentRepository().GetAvailableUnspentsForAddresses(
		ctx,
		derivedAddresses,
	)
	if err != nil {
		return nil, nil, err
	}

	return info, Unspents(unspents), nil
}

func (t *tradeService) unlockUnspentsForTrade(trade *domain.Trade) {
	p, _ := pset.NewPsetFromBase64(trade.PsetBase64)
	unspentKeys := make([]domain.UnspentKey, 0, len(p.Inputs))

	for _, in := range p.UnsignedTx.Inputs {
		unspentKeys = append(unspentKeys, domain.UnspentKey{
			TxID: bufferutil.TxIDFromBytes(in.Hash),
			VOut: in.Index,
		})
	}

	count, err := t.repoManager.UnspentRepository().UnlockUnspents(
		context.Background(),
		unspentKeys,
	)
	if err != nil {
		log.Warnf(
			"unable to unlock unspents for trade with id %s. You must run "+
				"ReloadUtxo RPC as soon as possible to restore the utxo set of the "+
				"internal wallet. Error: %v", trade.ID, err,
		)
	}

	log.Debugf("unlocked %d unspents", count)
}

func (t *tradeService) checkTradeExpiration(
	tradeID uuid.UUID,
	unspents []explorer.Utxo,
) {
	ctx := context.Background()
	trade, _ := t.repoManager.TradeRepository().GetOrCreateTrade(ctx, &tradeID)

	// If trade has been completed or settled nothing must be done.
	if trade.IsCompleted() || trade.IsSettled() {
		return
	}

	cb := circuitbreaker.NewCircuitBreaker()
	iBlocktime, _ := cb.Execute(func() (interface{}, error) {
		return t.explorerSvc.GetBlockHeight()
	})
	blocktime := iBlocktime.(int)

	if _, err := trade.Expire(uint64(blocktime)); err != nil {
		// If the trade hasn't expired yet, let's wait ideally for the next block
		// (2 minutes) to eventually spend ins of expired trade.
		if err == domain.ErrTradeExpirationDateNotReached {
			go func() {
				log.Infof(
					"no new block fetched in blockchain, postponed check for "+
						"expiration of trade with id: %s", tradeID,
				)
				time.Sleep(2 * time.Minute)
				t.checkTradeExpiration(tradeID, unspents)
			}()
			return
		}
		log.WithError(err).Warnf(
			"an error occured while setting status to expired for trade with id: %s",
			trade.ID,
		)
		return
	}

	t.blockchainListener.StopObserveOutpoints(unspents)

	vault, _ := t.repoManager.VaultRepository().GetOrCreateVault(
		ctx, nil, "", nil,
	)
	mnemonic, _ := vault.GetMnemonicSafe()

	_, mktAccountIndex, _ := t.repoManager.MarketRepository().GetMarketByAsset(
		ctx, trade.MarketQuoteAsset,
	)

	feeInfo, _ := vault.AllDerivedAddressesInfoForAccount(domain.FeeAccount)
	feeUnspents := make([]explorer.Utxo, 0)
	mktUnspents := make([]explorer.Utxo, 0)
	for _, u := range unspents {
		found := false
		for _, info := range feeInfo {
			script := hex.EncodeToString(u.Script())
			if script == info.Script {
				found = true
				break
			}
		}
		if found {
			feeUnspents = append(feeUnspents, u)
		} else {
			mktUnspents = append(mktUnspents, u)
		}
	}

	amountByAsset := make(map[string]uint64)
	for _, u := range mktUnspents {
		amountByAsset[u.Asset()] += u.Value()
	}
	outs := make([]TxOut, 0, len(amountByAsset))
	for asset, amount := range amountByAsset {
		info, _ := vault.DeriveNextExternalAddressForAccount(mktAccountIndex)
		outs = append(outs, TxOut{asset, int64(amount), info.Address})
	}
	outputs, outKeys, _ := parseRequestOutputs(outs)

	feeAccount, _ := vault.AccountByIndex(domain.FeeAccount)
	info, _ := vault.DeriveNextInternalAddressForAccount(domain.FeeAccount)
	feeChangePathByAsset := map[string]string{
		t.network.AssetID: feeAccount.DerivationPathByScript[info.Script],
	}
	mktAccount, _ := vault.AccountByIndex(mktAccountIndex)

	txHex, err := sendToMany(sendToManyOpts{
		mnemonic:              mnemonic,
		unspents:              mktUnspents,
		feeUnspents:           feeUnspents,
		outputs:               outputs,
		outputsBlindingKeys:   outKeys,
		changePathsByAsset:    make(map[string]string),
		feeChangePathByAsset:  feeChangePathByAsset,
		inputPathsByScript:    mktAccount.DerivationPathByScript,
		feeInputPathsByScript: feeAccount.DerivationPathByScript,
		milliSatPerByte:       domain.MinMilliSatPerByte,
		network:               t.network,
	})
	if err != nil {
		log.WithError(err).Warnf(
			"failed to create transaction to spend unspents of expired trade with id: %s",
			trade.ID,
		)
		return
	}

	cb = circuitbreaker.NewCircuitBreaker()
	iTxid, err := cb.Execute(func() (interface{}, error) {
		return t.explorerSvc.BroadcastTransaction(txHex)
	})
	if err != nil {
		log.WithError(err).Warnf(
			"failed to broadcast transactio to spend unspents of expired trade with id: %s",
			trade.ID,
		)
		return
	}
	txid := iTxid.(string)

	log.Infof(
		"spent inputs of expired trade with id: %s in tx: %s", trade.ID, txid,
	)

	unspentKeys := getUnspentKeys(unspents)
	if n, _ := t.repoManager.UnspentRepository().UnlockUnspents(ctx, unspentKeys); n > 0 {
		log.Debugf("unlocked %d unspents", n)
	}
	if _, err := t.repoManager.RunTransaction(ctx, false, func(ctx context.Context) (interface{}, error) {
		if err := t.repoManager.TradeRepository().UpdateTrade(
			ctx, &trade.ID, func(_ *domain.Trade) (*domain.Trade, error) {
				return trade, nil
			}); err != nil {
			return nil, err
		}
		if err := t.repoManager.VaultRepository().UpdateVault(
			ctx, func(_ *domain.Vault) (*domain.Vault, error) {
				return vault, nil
			}); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		log.WithError(err).Warn("failed to persist changes after spending unspents of expired trade with id: %s", tradeID)
	}
	t.blockchainListener.StartObserveTx(txid, trade.MarketQuoteAsset)
}

func fillProposal(opts FillProposalOpts) (*FillProposalResult, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: opts.Mnemonic,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %s", err)
	}

	network := opts.Network
	// fill swap request transaction with daemon's inputs and outputs
	psetBase64, selectedUnspentsForSwap, err := w.UpdateSwapTx(wallet.UpdateSwapTxOpts{
		PsetBase64:           opts.SwapRequest.GetTransaction(),
		Unspents:             opts.MarketUtxos,
		InputAmount:          opts.SwapRequest.GetAmountR(),
		InputAsset:           opts.SwapRequest.GetAssetR(),
		OutputAmount:         opts.SwapRequest.GetAmountP(),
		OutputAsset:          opts.SwapRequest.GetAssetP(),
		OutputDerivationPath: opts.OutputInfo.DerivationPath,
		ChangeDerivationPath: opts.ChangeInfo.DerivationPath,
		Network:              network,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update swap: %s", err)
	}

	// top-up fees using fee account. Note that the fee output is added after
	// blinding the transaction because it's explicit and must not be blinded
	psetWithFeesResult, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:        psetBase64,
		Unspents:          opts.FeeUtxos,
		MilliSatsPerBytes: domain.MinMilliSatPerByte,
		Network:           network,
		ChangePathsByAsset: map[string]string{
			network.AssetID: opts.FeeChangeInfo.DerivationPath,
		},
		WantPrivateBlindKeys: true,
		WantChangeForFees:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to topup for paying fees: %s", err)
	}

	// concat the selected unspents for paying fees with those for completing the
	// swap in order to get the full list of selected inputs
	selectedUnspents := append(selectedUnspentsForSwap, psetWithFeesResult.SelectedUnspents...)

	inputBlindingData, _, _ := wallet.ExtractBlindingDataFromTx(
		opts.SwapRequest.GetTransaction(),
		opts.SwapRequest.GetInputBlindingKey(),
		nil,
	)

	// get the indexes of the inputs of the tx to sign
	existingInputs := len(inputBlindingData)
	inputsToSign := len(selectedUnspents)

	for i, u := range selectedUnspents {
		inputBlindingData[existingInputs+i] = wallet.BlindingData{
			Asset:         u.Asset(),
			Amount:        u.Value(),
			AssetBlinder:  u.AssetBlinder(),
			AmountBlinder: u.ValueBlinder(),
		}
	}

	outputBlindingKeys := opts.SwapRequest.GetOutputBlindingKey()
	outputBlindingKeys[opts.OutputInfo.Script] = opts.OutputInfo.BlindingKey
	outputBlindingKeys[opts.ChangeInfo.Script] = opts.ChangeInfo.BlindingKey
	for script, blindKey := range psetWithFeesResult.ChangeOutputsBlindingKeys {
		outputBlindingKeys[script] = blindKey
	}

	// blind the transaction
	blindedPset, err := w.BlindSwapTransactionWithData(wallet.BlindSwapTransactionWithDataOpts{
		PsetBase64:         psetWithFeesResult.PsetBase64,
		InputBlindingData:  inputBlindingData,
		OutputBlindingKeys: outputBlindingKeys,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to blind: %s", err)
	}

	// add the explicit fee output to the tx
	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    transactionutil.NewFeeOutput(psetWithFeesResult.FeeAmount, network),
		Network:    network,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add explicit fees: %s", err)
	}

	// get the derivation paths of the selected inputs
	allInfo := append(opts.MarketInfo, opts.FeeInfo...)
	selectedInfo := getSelectedInfo(allInfo, selectedUnspents)

	signedPsetBase64 := blindedPlusFees.PsetBase64
	for i := 0; i < inputsToSign; i++ {
		inIndex := existingInputs + i
		signedPsetBase64, err = w.SignInput(wallet.SignInputOpts{
			PsetBase64:     signedPsetBase64,
			InIndex:        uint32(inIndex),
			DerivationPath: selectedInfo[i].DerivationPath,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to sign input %d tx: %s", inIndex, err)
		}
	}

	// get blinding private keys for selected inputs
	inputBlindingKeys := opts.SwapRequest.GetInputBlindingKey()
	for _, info := range selectedInfo {
		inputBlindingKeys[info.Script] = info.BlindingKey
	}

	return &FillProposalResult{
		PsetBase64:         signedPsetBase64,
		SelectedUnspents:   selectedUnspents,
		InputBlindingKeys:  inputBlindingKeys,
		OutputBlindingKeys: outputBlindingKeys,
	}, nil
}

func getSelectedInfo(allInfo domain.AddressesInfo, utxos []explorer.Utxo) domain.AddressesInfo {
	contains := func(script string) *domain.AddressInfo {
		for _, info := range allInfo {
			if info.Script == script {
				return &info
			}
		}
		return nil
	}

	selectedInfo := make(domain.AddressesInfo, 0)
	for _, u := range utxos {
		script := hex.EncodeToString(u.Script())
		if info := contains(script); info != nil {
			selectedInfo = append(selectedInfo, *info)
		}
	}
	return selectedInfo
}

func getUnspentKeys(unspents []explorer.Utxo) []domain.UnspentKey {
	keys := make([]domain.UnspentKey, 0, len(unspents))
	for _, u := range unspents {
		keys = append(keys, domain.UnspentKey{
			TxID: u.Hash(),
			VOut: u.Index(),
		})
	}
	return keys
}

func mergeDerivationPaths(maps ...map[string]string) map[string]string {
	merge := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			merge[k] = v
		}
	}
	return merge
}

// previewForMarket returns the current price and balances of a market, along
// with a preview amount for a BUY or SELL trade based on the strategy type.
func previewForMarket(
	unspents []domain.Unspent,
	market *domain.Market,
	tradeType int,
	amount uint64,
	asset string,
) (*PriceWithFee, error) {
	isBuy := tradeType == TradeBuy
	isBaseAsset := asset == market.BaseAsset

	balances := getBalanceByAsset(unspents)
	marketBalance := Balance{
		BaseAmount:  balances[market.BaseAsset],
		QuoteAmount: balances[market.QuoteAsset],
	}

	preview, err := market.Preview(
		marketBalance.BaseAmount, marketBalance.QuoteAmount, amount,
		isBaseAsset, isBuy,
	)
	if err != nil {
		return nil, err
	}

	return &PriceWithFee{
		Price: Price(preview.Price),
		Fee: Fee{
			BasisPoint:    market.Fee,
			FixedBaseFee:  market.FixedFee.BaseFee,
			FixedQuoteFee: market.FixedFee.QuoteFee,
		},
		Amount:  preview.Amount,
		Asset:   preview.Asset,
		Balance: marketBalance,
	}, nil
}

func getBalanceByAsset(unspents []domain.Unspent) map[string]uint64 {
	balances := map[string]uint64{}
	for _, unspent := range unspents {
		if _, ok := balances[unspent.AssetHash]; !ok {
			balances[unspent.AssetHash] = 0
		}
		balances[unspent.AssetHash] += unspent.Value
	}
	return balances
}

// isValidPrice checks that the amounts of the trade are valid by
// making a preview of each counter amounts of the swap given the
// current price of the market.
// Since the price is variable in time, the predicted amounts are not compared
// against those of the swap, but rather they are used to create a range in
// which the swap amounts must be included to be considered valid.
func isValidTradePrice(
	swapRequest domain.SwapRequest,
	tradeType int,
	market *domain.Market,
	unspents []domain.Unspent,
	slippage decimal.Decimal,
) bool {
	// TODO: parallelize the 2 ways of calculating and validating the preview
	// amount to speed up the process.
	amount := swapRequest.GetAmountR()
	if tradeType == TradeSell {
		amount = swapRequest.GetAmountP()
	}

	preview, _ := previewForMarket(
		unspents,
		market,
		tradeType,
		amount,
		market.BaseAsset,
	)

	if preview != nil {
		if isPriceInRange(swapRequest, tradeType, preview.Amount, true, slippage) {
			return true
		}
	}

	amount = swapRequest.GetAmountP()
	if tradeType == TradeSell {
		amount = swapRequest.GetAmountR()
	}

	preview, _ = previewForMarket(
		unspents,
		market,
		tradeType,
		amount,
		market.QuoteAsset,
	)

	if preview == nil {
		return false
	}

	return isPriceInRange(swapRequest, tradeType, preview.Amount, false, slippage)
}

func isPriceInRange(
	swapRequest domain.SwapRequest,
	tradeType int,
	previewAmount uint64,
	isPreviewForQuoteAsset bool,
	slippage decimal.Decimal,
) bool {
	amountToCheck := decimal.NewFromInt(int64(swapRequest.GetAmountP()))
	if tradeType == TradeSell {
		if isPreviewForQuoteAsset {
			amountToCheck = decimal.NewFromInt(int64(swapRequest.GetAmountR()))
		}
	} else {
		if !isPreviewForQuoteAsset {
			amountToCheck = decimal.NewFromInt(int64(swapRequest.GetAmountR()))
		}
	}

	expectedAmount := decimal.NewFromInt(int64(previewAmount))
	lowerBound := expectedAmount.Mul(decimal.NewFromInt(1).Sub(slippage))
	upperBound := expectedAmount.Mul(decimal.NewFromInt(1).Add(slippage))

	return amountToCheck.GreaterThanOrEqual(lowerBound) && amountToCheck.LessThanOrEqual(upperBound)
}
