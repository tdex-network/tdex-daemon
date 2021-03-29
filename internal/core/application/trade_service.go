package application

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
	pkgswap "github.com/tdex-network/tdex-daemon/pkg/swap"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/address"
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
		swapComplete *domain.SwapComplete,
		swapFail *domain.SwapFail,
	) (string, domain.SwapFail, error)
	GetMarketBalance(
		ctx context.Context,
		market Market,
	) (*BalanceWithFee, error)
}

type tradeService struct {
	dbManager          ports.DbManager
	explorerSvc        explorer.Service
	blockchainListener BlockchainListener
	marketBaseAsset    string
	expiryDuration     uint64
	priceSlippage      decimal.Decimal
	network            *network.Network
}

func NewTradeService(
	dbManager ports.DbManager,
	explorerSvc explorer.Service,
	bcListener BlockchainListener,
	marketBaseAsset string,
	expiryDuration time.Duration,
	priceSlippage decimal.Decimal,
	net *network.Network,
) TradeService {
	return newTradeService(
		dbManager,
		explorerSvc,
		bcListener,
		marketBaseAsset,
		uint64(expiryDuration),
		priceSlippage,
		net,
	)
}

func newTradeService(
	dbManager ports.DbManager,
	explorerSvc explorer.Service,
	bcListener BlockchainListener,
	marketBaseAsset string,
	expiryDuration uint64,
	priceSlippage decimal.Decimal,
	net *network.Network,
) *tradeService {
	return &tradeService{
		dbManager:          dbManager,
		explorerSvc:        explorerSvc,
		blockchainListener: bcListener,
		marketBaseAsset:    marketBaseAsset,
		expiryDuration:     expiryDuration,
		priceSlippage:      priceSlippage,
		network:            net,
	}
}

// Markets is the domain controller for the Markets RPC
func (t *tradeService) GetTradableMarkets(ctx context.Context) (
	[]MarketWithFee,
	error,
) {
	tradableMarkets, err := t.dbManager.MarketRepository().GetTradableMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketsWithFee := make([]MarketWithFee, 0, len(tradableMarkets))
	for _, mkt := range tradableMarkets {
		marketsWithFee = append(marketsWithFee, MarketWithFee{
			Market: Market{
				BaseAsset:  mkt.BaseAsset,
				QuoteAsset: mkt.QuoteAsset,
			},
			Fee: Fee{
				BasisPoint: mkt.Fee,
			},
		})
	}

	return marketsWithFee, nil
}

// MarketPrice is the domain controller for the MarketPrice RPC.
func (t *tradeService) GetMarketPrice(
	ctx context.Context,
	market Market,
	tradeType int,
	amount uint64,
	asset string,
) (*PriceWithFee, error) {
	// check the asset strings
	if err := validateAssetString(market.BaseAsset); err != nil {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	if err := validateAssetString(market.QuoteAsset); err != nil {
		return nil, domain.ErrMarketInvalidQuoteAsset
	}

	// Checks if base asset is correct
	if market.BaseAsset != t.marketBaseAsset {
		return nil, ErrMarketNotExist
	}

	if err := validateAssetString(asset); err != nil {
		return nil, errors.New("invalid asset")
	}
	if asset != market.BaseAsset && asset != market.QuoteAsset {
		return nil, errors.New("asset must match one of those of the market")
	}

	//Checks if market exist
	mkt, mktAccountIndex, err := t.dbManager.MarketRepository().GetMarketByAsset(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if mktAccountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	if !mkt.IsTradable() {
		return nil, domain.ErrMarketIsClosed
	}

	unspents, _, _, _, err := t.getUnspentsBlindingsAndDerivationPathsForAccount(ctx, mktAccountIndex)
	if err != nil {
		return nil, err
	}

	price, previewAmount, err := getPriceAndPreviewForMarket(unspents, mkt, tradeType, amount, asset)
	if err != nil {
		return nil, err
	}

	previewAsset := market.BaseAsset
	if asset == market.BaseAsset {
		previewAsset = market.QuoteAsset
	}

	return &PriceWithFee{
		Price: price,
		Fee: Fee{
			BasisPoint: mkt.Fee,
		},
		Amount: previewAmount,
		Asset:  previewAsset,
	}, nil
}

// TradePropose is the domain controller for the TradePropose RPC
func (t *tradeService) TradePropose(
	ctx context.Context,
	market Market,
	tradeType int,
	swapRequest domain.SwapRequest,
) (domain.SwapAccept, domain.SwapFail, uint64, error) {
	// check the asset strings
	if err := validateAssetString(market.BaseAsset); err != nil {
		return nil, nil, 0, domain.ErrMarketInvalidBaseAsset
	}

	if err := validateAssetString(market.QuoteAsset); err != nil {
		return nil, nil, 0, domain.ErrMarketInvalidQuoteAsset
	}

	mkt, marketAccountIndex, err := t.dbManager.MarketRepository().GetMarketByAsset(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		return nil, nil, 0, err
	}
	if marketAccountIndex < 0 {
		return nil, nil, 0, ErrMarketNotExist
	}

	// get all unspents for market account (both as []domain.Unspents and as
	// []explorer.Utxo)along with private blinding keys and signing derivation
	// paths for respectively unblinding and signing them later
	marketUnspents, marketUtxos, marketBlindingKeysByScript, marketDerivationPaths, err :=
		t.getUnspentsBlindingsAndDerivationPathsForAccount(ctx, marketAccountIndex)
	if err != nil {
		return nil, nil, 0, err
	}

	// Check we got at least one
	if len(marketUnspents) == 0 || len(marketUtxos) == 0 {
		return nil, nil, 0, ErrMarketNotFunded
	}

	// ... and the same for fee account (we'll need to top-up fees)
	feeUnspents, feeUtxos, feeBlindingKeysByScript, feeDerivationPaths, err :=
		t.getUnspentsBlindingsAndDerivationPathsForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return nil, nil, 0, err
	}
	// Check we got at least one
	if len(feeUnspents) <= 0 {
		return nil, nil, 0, ErrFeeAccountNotFunded
	}

	var mnemonic []string
	var outputBlindingKeysByScript map[string][]byte
	var outputDerivationPath, changeDerivationPath, feeChangeDerivationPath string

	// derive output and change address for market, and change address for fee account
	if err := t.dbManager.VaultRepository().UpdateVault(
		ctx,
		func(v *domain.Vault) (*domain.Vault, error) {
			mnemonic, err = v.GetMnemonicSafe()
			if err != nil {
				return nil, err
			}
			info, err := v.DeriveNextExternalAddressForAccount(marketAccountIndex)
			if err != nil {
				return nil, err
			}
			changeInfo, err := v.DeriveNextInternalAddressForAccount(marketAccountIndex)
			if err != nil {
				return nil, err
			}
			feeChangeInfo, err := v.DeriveNextInternalAddressForAccount(domain.FeeAccount)
			if err != nil {
				return nil, err
			}

			outputBlindingKeysByScript = map[string][]byte{
				info.Script:       info.BlindingKey,
				changeInfo.Script: changeInfo.BlindingKey,
			}

			outputDerivationPath = info.DerivationPath
			changeDerivationPath = changeInfo.DerivationPath
			feeChangeDerivationPath = feeChangeInfo.DerivationPath

			return v, nil
		}); err != nil {
		return nil, nil, 0, err
	}

	var swapAccept domain.SwapAccept
	var swapFail domain.SwapFail
	var swapExpiryTime uint64
	var selectedUnspentKeys []domain.UnspentKey
	var tradeID string
	var lockedUnspentsCount int

	// parse swap proposal and possibly accept
	trade, err := t.dbManager.TradeRepository().GetOrCreateTrade(ctx, nil)
	if err != nil {
		return nil, nil, 0, err
	}

	if err := t.dbManager.TradeRepository().UpdateTrade(
		ctx,
		&trade.ID,
		func(tt *domain.Trade) (*domain.Trade, error) {
			ok, err := tt.Propose(
				swapRequest,
				market.QuoteAsset, mkt.Fee,
				nil,
			)
			if err != nil {
				return nil, err
			}
			if !ok {
				swapFail = tt.SwapFailMessage()
				return tt, nil
			}

			if !isValidTradePrice(swapRequest, tradeType, mkt, marketUnspents, t.priceSlippage) {
				tt.Fail(
					swapRequest.GetId(),
					int(pkgswap.ErrCodeInvalidSwapRequest),
					"bad pricing",
				)
				swapFail = tt.SwapFailMessage()
				return tt, nil
			}

			acceptSwapResult, err := acceptSwap(acceptSwapOpts{
				mnemonic:                   mnemonic,
				swapRequest:                swapRequest,
				marketUnspents:             marketUtxos,
				feeUnspents:                feeUtxos,
				marketBlindingKeysByScript: marketBlindingKeysByScript,
				feeBlindingKeysByScript:    feeBlindingKeysByScript,
				outputBlindingKeysByScript: outputBlindingKeysByScript,
				marketDerivationPaths:      marketDerivationPaths,
				feeDerivationPaths:         feeDerivationPaths,
				outputDerivationPath:       outputDerivationPath,
				changeDerivationPath:       changeDerivationPath,
				feeChangeDerivationPath:    feeChangeDerivationPath,
				network:                    t.network,
			})
			if err != nil {
				return nil, err
			}

			ok, err = tt.Accept(
				acceptSwapResult.psetBase64,
				acceptSwapResult.inputBlindingKeys,
				acceptSwapResult.outputBlindingKeys,
				t.expiryDuration,
			)
			if err != nil {
				return nil, err
			}
			if !ok {
				swapFail = tt.SwapFailMessage()
				return tt, nil
			}

			// Last thing to do is to lock the unspents. In case something goes wrong
			// here, an error is returned and the accepted trade is discarded.
			selectedUnspentKeys = getUnspentKeys(acceptSwapResult.selectedUnspents)
			count, err := t.dbManager.UnspentRepository().LockUnspents(
				ctx,
				selectedUnspentKeys,
				tt.ID,
			)
			if err != nil {
				return nil, err
			}

			swapAccept = tt.SwapAcceptMessage()
			swapExpiryTime = tt.ExpiryTime
			trade = tt
			lockedUnspentsCount = count
			tradeID = trade.ID.String()

			return tt, nil
		}); err != nil {
		return nil, nil, 0, err
	}

	if swapFail == nil {
		go t.blockchainListener.StartObserveTx(trade.TxID)
		go func() {
			// admit a tollerance of 1 minute past the expiration time.
			time.Sleep(time.Duration(t.expiryDuration+60) * time.Second)
			t.checkTradeExpiration(trade.TxID, selectedUnspentKeys)
		}()

		log.Infof("trade with %s accepted", tradeID)
		log.Debugf("locked %d unspents", lockedUnspentsCount)
	}

	return swapAccept, swapFail, swapExpiryTime, nil
}

// TradeComplete is the domain controller for the TradeComplete RPC
func (t *tradeService) TradeComplete(
	ctx context.Context,
	swapComplete *domain.SwapComplete,
	swapFail *domain.SwapFail,
) (string, domain.SwapFail, error) {
	if swapFail != nil {
		swapFailMsg, err := t.tradeFail(ctx, *swapFail)
		return "", swapFailMsg, err
	}

	return t.tradeComplete(ctx, *swapComplete)
}

func (t *tradeService) tradeComplete(
	ctx context.Context,
	swapComplete domain.SwapComplete,
) (txID string, swapFail domain.SwapFail, err error) {
	trade, err := t.dbManager.TradeRepository().GetTradeBySwapAcceptID(ctx, swapComplete.GetAcceptId())
	if err != nil {
		return
	}

	psetBase64 := swapComplete.GetTransaction()

	// here we manipulate the trade to reach the Complete status
	res, err := trade.Complete(psetBase64)
	if err != nil {
		return
	}
	// for domain related errors, we check for swap failures that can happens
	// for tradin related problems or transaction manomission
	if !res.OK {
		swapFail = trade.SwapFailMessage()
		return
	}
	log.Infof("trade with id %s completed", trade.ID)

	// we are going to broadcast the transaction, this will actually tell if the
	//transaction is a valid one to be included in blockcchain
	if _, err = t.explorerSvc.BroadcastTransaction(res.TxHex); err != nil {
		log.WithError(err).WithField("hex", res.TxHex).Warn("unable to broadcast trade tx")
		return
	}

	txID = res.TxID
	log.Infof("trade with id %s broadcasted: %s", trade.ID, txID)

	// we make sure that any problem happening at this point
	// is not influencing the trade therefore we run as goroutine
	// this method will take care to retry to handle potential
	// datastore conflicts (if any) at repository level
	go func() {
		if err := t.dbManager.TradeRepository().UpdateTrade(
			ctx,
			&trade.ID,
			func(previousTrade *domain.Trade) (*domain.Trade, error) { return trade, nil },
		); err != nil {
			log.Error("unable to persist completed trade with id ", trade.ID, " : ", err.Error())
		}

		_, accountIndex, _ := t.dbManager.MarketRepository().GetMarketByAsset(
			ctx,
			trade.MarketQuoteAsset,
		)
		extractUnspentsFromTxAndUpdateUtxoSet(
			t.dbManager.UnspentRepository(),
			t.dbManager.VaultRepository(),
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
	trade, err := t.dbManager.TradeRepository().GetTradeBySwapAcceptID(ctx, swapID)
	if err != nil {
		return nil, err
	}

	tradeID := trade.ID
	if err := t.dbManager.TradeRepository().UpdateTrade(
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

func (t *tradeService) getUnspentsBlindingsAndDerivationPathsForAccount(
	ctx context.Context,
	account int,
) (
	[]domain.Unspent,
	[]explorer.Utxo,
	map[string][]byte,
	map[string]string,
	error,
) {
	info, err := t.dbManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(ctx, account)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	derivedAddresses, blindingKeys := info.AddressesAndKeys()

	unspents, err := t.dbManager.UnspentRepository().GetAvailableUnspentsForAddresses(
		ctx,
		derivedAddresses,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	utxos := make([]explorer.Utxo, 0, len(unspents))
	for _, unspent := range unspents {
		utxos = append(utxos, unspent.ToUtxo())
	}

	scripts := make([]string, 0, len(derivedAddresses))
	for _, addr := range derivedAddresses {
		script, _ := address.ToOutputScript(addr)
		scripts = append(scripts, hex.EncodeToString(script))
	}
	derivationPaths, _ := t.dbManager.VaultRepository().GetDerivationPathByScript(
		ctx,
		account,
		scripts,
	)

	blindingKeysByScript := map[string][]byte{}
	for i, addr := range derivedAddresses {
		script, _ := address.ToOutputScript(addr)
		blindingKeysByScript[hex.EncodeToString(script)] = blindingKeys[i]
	}

	return unspents, utxos, blindingKeysByScript, derivationPaths, nil
}

func (t *tradeService) unlockUnspentsForTrade(trade *domain.Trade) {
	p, _ := pset.NewPsetFromBase64(trade.PsetBase64)
	keyLen := len(p.Inputs)
	unspentKeys := make([]domain.UnspentKey, keyLen, keyLen)

	for i, in := range p.UnsignedTx.Inputs {
		unspentKeys[i] = domain.UnspentKey{
			TxID: bufferutil.TxIDFromBytes(in.Hash),
			VOut: in.Index,
		}
	}

	count, err := t.dbManager.UnspentRepository().UnlockUnspents(
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
	tradeTxID string,
	selectedUnspentKeys []domain.UnspentKey,
) {
	ctx := context.Background()

	// if the trade is expired it's required to unlock the unspents used as input
	// and to bring the trade to failed status
	trade, _ := t.dbManager.TradeRepository().GetTradeByTxID(ctx, tradeTxID)
	if trade.IsExpired() {
		t.blockchainListener.StopObserveTx(trade.TxID)

		count, err := t.dbManager.UnspentRepository().UnlockUnspents(ctx, selectedUnspentKeys)
		if err != nil {
			log.Warnf(
				"trade with id %s has expired but an error occured while "+
					"unlocking its unspents. You must run ReloadUtxo RPC as soon as "+
					"possible to restore the utxo set of the internal wallet. Error: %v",
				trade.ID, err,
			)
			return
		}
		log.Debugf("unlocked %d unspents", count)

		if err := t.dbManager.TradeRepository().UpdateTrade(
			ctx,
			&trade.ID,
			func(tt *domain.Trade) (*domain.Trade, error) {
				if _, err := tt.Expire(); err != nil {
					return nil, err
				}
				return tt, nil
			},
		); err != nil {
			log.Warnf("unable to persist expiration of trade with id %s", trade.ID)
			return
		}
		log.Infof("trade with id %s expired", trade.ID)
		return
	}
}

type acceptSwapOpts struct {
	mnemonic                   []string
	swapRequest                domain.SwapRequest
	marketUnspents             []explorer.Utxo
	feeUnspents                []explorer.Utxo
	marketBlindingKeysByScript map[string][]byte
	feeBlindingKeysByScript    map[string][]byte
	outputBlindingKeysByScript map[string][]byte
	marketDerivationPaths      map[string]string
	feeDerivationPaths         map[string]string
	outputDerivationPath       string
	changeDerivationPath       string
	feeChangeDerivationPath    string
	network                    *network.Network
}

type acceptSwapResult struct {
	psetBase64         string
	selectedUnspents   []explorer.Utxo
	inputBlindingKeys  map[string][]byte
	outputBlindingKeys map[string][]byte
}

func acceptSwap(opts acceptSwapOpts) (res acceptSwapResult, err error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: opts.mnemonic,
	})
	if err != nil {
		return
	}
	network := opts.network
	// fill swap request transaction with daemon's inputs and outputs
	psetBase64, selectedUnspentsForSwap, err := w.UpdateSwapTx(wallet.UpdateSwapTxOpts{
		PsetBase64:           opts.swapRequest.GetTransaction(),
		Unspents:             opts.marketUnspents,
		InputAmount:          opts.swapRequest.GetAmountR(),
		InputAsset:           opts.swapRequest.GetAssetR(),
		OutputAmount:         opts.swapRequest.GetAmountP(),
		OutputAsset:          opts.swapRequest.GetAssetP(),
		OutputDerivationPath: opts.outputDerivationPath,
		ChangeDerivationPath: opts.changeDerivationPath,
		Network:              network,
	})
	if err != nil {
		return
	}

	// top-up fees using fee account. Note that the fee output is added after
	// blinding the transaction because it's explicit and must not be blinded
	psetWithFeesResult, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:        psetBase64,
		Unspents:          opts.feeUnspents,
		MilliSatsPerBytes: domain.MinMilliSatPerByte,
		Network:           network,
		ChangePathsByAsset: map[string]string{
			network.AssetID: opts.feeChangeDerivationPath,
		},
		WantPrivateBlindKeys: true,
		WantChangeForFees:    true,
	})
	if err != nil {
		return
	}

	// concat the selected unspents for paying fees with those for completing the
	// swap in order to get the full list of selected inputs
	selectedUnspents := append(selectedUnspentsForSwap, psetWithFeesResult.SelectedUnspents...)

	// get blinding private keys for selected inputs
	unspentsBlindingKeys := mergeBlindingKeys(opts.marketBlindingKeysByScript, opts.feeBlindingKeysByScript)
	selectedInBlindingKeys := getSelectedBlindingKeys(unspentsBlindingKeys, selectedUnspents)
	// ... and merge with those contained into the swapRequest (trader keys)
	inputBlindingKeys := mergeBlindingKeys(opts.swapRequest.GetInputBlindingKey(), selectedInBlindingKeys)

	// same for output  public blinding keys
	outputBlindingKeys := mergeBlindingKeys(
		opts.outputBlindingKeysByScript,
		psetWithFeesResult.ChangeOutputsBlindingKeys,
		opts.swapRequest.GetOutputBlindingKey(),
	)

	// blind the transaction
	blindedPset, err := w.BlindSwapTransactionWithKeys(wallet.BlindSwapTransactionWithKeysOpts{
		PsetBase64:         psetWithFeesResult.PsetBase64,
		InputBlindingKeys:  inputBlindingKeys,
		OutputBlindingKeys: outputBlindingKeys,
	})
	if err != nil {
		return
	}

	// add the explicit fee output to the tx
	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    transactionutil.NewFeeOutput(psetWithFeesResult.FeeAmount, network),
		Network:    network,
	})
	if err != nil {
		return
	}

	// get the indexes of the inputs of the tx to sign
	inputsToSign := getInputsIndexes(psetWithFeesResult.PsetBase64, selectedUnspents)
	// get the derivation paths of the selected inputs
	unspentsDerivationPaths := mergeDerivationPaths(opts.marketDerivationPaths, opts.feeDerivationPaths)
	derivationPaths := getSelectedDerivationPaths(unspentsDerivationPaths, selectedUnspents)

	signedPsetBase64 := blindedPlusFees.PsetBase64
	for i, inIndex := range inputsToSign {
		signedPsetBase64, err = w.SignInput(wallet.SignInputOpts{
			PsetBase64:     signedPsetBase64,
			InIndex:        inIndex,
			DerivationPath: derivationPaths[i],
		})
	}

	res = acceptSwapResult{
		psetBase64:         signedPsetBase64,
		selectedUnspents:   selectedUnspents,
		inputBlindingKeys:  inputBlindingKeys,
		outputBlindingKeys: outputBlindingKeys,
	}

	return
}

func getInputsIndexes(psetBase64 string, unspents []explorer.Utxo) []uint32 {
	indexes := make([]uint32, 0, len(unspents))

	ptx, _ := pset.NewPsetFromBase64(psetBase64)
	for _, u := range unspents {
		for i, in := range ptx.UnsignedTx.Inputs {
			if u.Hash() == bufferutil.TxIDFromBytes(in.Hash) && u.Index() == in.Index {
				indexes = append(indexes, uint32(i))
				break
			}
		}
	}
	return indexes
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

func mergeBlindingKeys(maps ...map[string][]byte) map[string][]byte {
	merge := make(map[string][]byte, 0)
	for _, m := range maps {
		for k, v := range m {
			merge[k] = v
		}
	}
	return merge
}

func mergeDerivationPaths(maps ...map[string]string) map[string]string {
	merge := make(map[string]string, 0)
	for _, m := range maps {
		for k, v := range m {
			merge[k] = v
		}
	}
	return merge
}

func getSelectedDerivationPaths(derivationPaths map[string]string, unspents []explorer.Utxo) []string {
	selectedPaths := make([]string, 0)
	for _, unspent := range unspents {
		script := hex.EncodeToString(unspent.Script())
		selectedPaths = append(selectedPaths, derivationPaths[script])
	}
	return selectedPaths
}

func getSelectedBlindingKeys(blindingKeys map[string][]byte, unspents []explorer.Utxo) map[string][]byte {
	selectedKeys := map[string][]byte{}
	for _, unspent := range unspents {
		script := hex.EncodeToString(unspent.Script())
		selectedKeys[script] = blindingKeys[script]
	}
	return selectedKeys
}

// getPriceAndPreviewForMarket returns the current price of a market, along
// with a amount preview for a BUY or SELL trade.
// Depending on the strategy set for the market, an input amount might be
// required to calculate the preview amount.
// In the specific, if the market strategy is not pluggable, the preview amount
// is calculated with either InGivenOut or OutGivenIn methods of the
// MakingFormula interface. Otherwise, the price is simply retrieved from the
// domain.Market instance and the preview amount is calculated by applying the
// market fees within the conversion.
// The incoming amount always represents an amount of the market's base asset,
// therefore a preview amount for the correspoing quote asset is returned.
func getPriceAndPreviewForMarket(
	unspents []domain.Unspent,
	market *domain.Market,
	tradeType int,
	amount uint64,
	asset string,
) (Price, uint64, error) {
	balances := getBalanceByAsset(unspents)
	baseAssetBalance := balances[market.BaseAsset]
	quoteAssetBalance := balances[market.QuoteAsset]

	if tradeType == TradeBuy {
		if asset == market.BaseAsset && amount >= baseAssetBalance {
			return Price{}, 0, errors.New("provided amount is too big")
		}
	} else {
		if asset == market.QuoteAsset && amount >= quoteAssetBalance {
			return Price{}, 0, errors.New("provided amount is too big")
		}
	}

	if market.IsStrategyPluggable() {
		previewAmount := calcPreviewAmount(
			market,
			baseAssetBalance,
			quoteAssetBalance,
			tradeType,
			amount,
			asset,
		)

		price := Price{
			BasePrice:  market.BaseAssetPrice(),
			QuotePrice: market.QuoteAssetPrice(),
		}
		return price, previewAmount, nil
	}

	price, previewAmount, err := previewFromFormula(
		market,
		baseAssetBalance,
		quoteAssetBalance,
		tradeType,
		amount,
		asset,
	)
	if err != nil {
		return Price{}, 0, err
	}

	if tradeType == TradeBuy {
		if asset == market.QuoteAsset && previewAmount >= baseAssetBalance {
			return Price{}, 0, errors.New("provided amount is too big")
		}
	} else {
		if asset == market.QuoteAsset && previewAmount >= baseAssetBalance {
			return Price{}, 0, errors.New("provided amount is too big")
		}
	}

	return *price, previewAmount, nil
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

// calcPreviewAmount calculates the amount of a market's quote asset due,
// depending on the trade type and the base asset amount provided.
// The market fees are either added or subtracted to the converted amount
// basing on the trade type.
func calcPreviewAmount(
	market *domain.Market,
	baseAssetBalance, quoteAssetBalance uint64,
	tradeType int,
	amount uint64,
	asset string,
) uint64 {
	price := market.QuoteAssetPrice()
	if asset != market.BaseAsset {
		price = market.BaseAssetPrice()
	}

	chargeFeesOnTheWayIn := asset == market.BaseAsset
	var previewAmount uint64
	if tradeType == TradeBuy {
		previewAmount = calcProposeAmount(amount, market.Fee, price, chargeFeesOnTheWayIn)
	} else {
		previewAmount = calcExpectedAmount(amount, market.Fee, price, chargeFeesOnTheWayIn)
	}

	return previewAmount
}

// calcProposeAmount returns the quote asset amount due for a BUY trade, that,
// remind, expresses a willing of buying a certain amount of the market's base
// asset.
// The market fees can be collected in either base or quote asset, but this is
// not relevant when calculating the preview amount. The reason is explained
// with the following example:
//
// Alice wants to BUY 0.1 LBTC in exchange for USDT (hence LBTC/USDT market).
// Lets assume the provider holds 10 LBTC and 65000 USDT in his reserves, so
// the USDT/LBTC price is 6500.
// Depending on how the fees are collected we have:
// - fee_asset = LBTC
//		feeAmount = lbtcAmount * feePercentage
// 		usdtAmount = (lbtcAmount + feeAmount) * price =
//			= (lbtcAmount + lbtcAmount * feeAmount) * price =
//			= (1 + feeAmount) * lbtcAmount * price = 1.25 * 0.1 * 6500 = 812,5 USDT
// - fee_asset = USDT
//		cAmount = lbtcAmount * price
// 		feeAmount = cAmount * feePercentage
// 		usdtAmount = cAmount + feeAmount =
//			= (lbtcAmount * price) + (lbtcAmount * price * feePercentage)
// 			= lbtcAmount * price * (1 + feePercentage) = 0.1  * 6500 * 1,25 = 812,5 USDT
func calcProposeAmount(
	amount uint64,
	feeAmount int64,
	price decimal.Decimal,
	chargeFeesOnTheWayIn bool,
) uint64 {
	netAmountR := decimal.NewFromInt(int64(amount)).Mul(price).BigInt().Uint64()
	if !chargeFeesOnTheWayIn {
		amountP, _ := mathutil.LessFee(netAmountR, uint64(feeAmount))
		return amountP
	}

	amountP, _ := mathutil.PlusFee(netAmountR, uint64(feeAmount))
	return amountP
}

func calcExpectedAmount(
	amount uint64,
	feeAmount int64,
	price decimal.Decimal,
	chargeFeesOnTheWayIn bool,
) uint64 {
	netAmountP := decimal.NewFromInt(int64(amount)).Mul(price).BigInt().Uint64()

	if !chargeFeesOnTheWayIn {
		amountR, _ := mathutil.PlusFee(netAmountP, uint64(feeAmount))
		return amountR
	}

	amountR, _ := mathutil.LessFee(netAmountP, uint64(feeAmount))
	return amountR
}

func previewFromFormula(
	market *domain.Market,
	baseAssetBalance, quoteAssetBalance uint64,
	tradeType int,
	amount uint64,
	asset string,
) (price *Price, previewAmount uint64, err error) {
	formula := market.Strategy.Formula()

	if tradeType == TradeBuy {
		if asset == market.BaseAsset {
			previewAmount, err = formula.InGivenOut(
				&mm.FormulaOpts{
					BalanceIn:           quoteAssetBalance,
					BalanceOut:          baseAssetBalance,
					Fee:                 uint64(market.Fee),
					ChargeFeeOnTheWayIn: true,
				},
				amount,
			)
		} else {
			previewAmount, err = formula.OutGivenIn(
				&mm.FormulaOpts{
					BalanceIn:           quoteAssetBalance,
					BalanceOut:          baseAssetBalance,
					Fee:                 uint64(market.Fee),
					ChargeFeeOnTheWayIn: true,
				},
				amount,
			)
		}
	} else {
		if asset == market.BaseAsset {
			previewAmount, err = formula.OutGivenIn(
				&mm.FormulaOpts{
					BalanceIn:           baseAssetBalance,
					BalanceOut:          quoteAssetBalance,
					Fee:                 uint64(market.Fee),
					ChargeFeeOnTheWayIn: true,
				},
				amount,
			)
		} else {
			previewAmount, err = formula.InGivenOut(
				&mm.FormulaOpts{
					BalanceIn:           baseAssetBalance,
					BalanceOut:          quoteAssetBalance,
					Fee:                 uint64(market.Fee),
					ChargeFeeOnTheWayIn: true,
				},
				amount,
			)
		}
	}
	if err != nil {
		return
	}

	// we can ignore errors because if the above function calls do not return
	// any, we can assume the following do the same because they all perform the
	// same checks.
	price, _ = priceFromBalances(formula, baseAssetBalance, quoteAssetBalance)

	return price, previewAmount, nil
}

func priceFromBalances(
	formula mm.MakingFormula,
	baseAssetBalance,
	quoteAssetBalance uint64,
) (*Price, error) {
	basePrice, err := formula.SpotPrice(&mm.FormulaOpts{
		BalanceIn:  quoteAssetBalance,
		BalanceOut: baseAssetBalance,
	})
	if err != nil {
		return nil, err
	}
	quotePrice, _ := formula.SpotPrice(&mm.FormulaOpts{
		BalanceIn:  baseAssetBalance,
		BalanceOut: quoteAssetBalance,
	})
	if err != nil {
		return nil, err
	}

	return &Price{
		BasePrice:  basePrice,
		QuotePrice: quotePrice,
	}, nil
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
	// TODO: parallelize the 2 ways of calculating and valifdating the preview
	// amount to speed up the process.
	amount := swapRequest.GetAmountR()
	if tradeType == TradeSell {
		amount = swapRequest.GetAmountP()
	}

	_, previewAmount, _ := getPriceAndPreviewForMarket(
		unspents,
		market,
		tradeType,
		amount,
		market.BaseAsset,
	)

	if isPriceInRange(swapRequest, tradeType, previewAmount, true, slippage) {
		return true
	}

	amount = swapRequest.GetAmountP()
	if tradeType == TradeSell {
		amount = swapRequest.GetAmountR()
	}

	_, previewAmount, _ = getPriceAndPreviewForMarket(
		unspents,
		market,
		tradeType,
		amount,
		market.QuoteAsset,
	)

	return isPriceInRange(swapRequest, tradeType, previewAmount, false, slippage)
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

func (t *tradeService) GetMarketBalance(
	ctx context.Context,
	market Market,
) (*BalanceWithFee, error) {
	// check the asset strings
	err := validateAssetString(market.BaseAsset)
	if err != nil {
		return nil, domain.ErrMarketInvalidBaseAsset
	}

	err = validateAssetString(market.QuoteAsset)
	if err != nil {
		return nil, domain.ErrMarketInvalidQuoteAsset
	}

	m, accountIndex, err := t.dbManager.MarketRepository().GetMarketByAsset(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	info, err := t.dbManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(ctx, m.AccountIndex)
	if err != nil {
		return nil, err
	}
	marketAddresses := info.Addresses()

	baseAssetBalance, err := t.dbManager.UnspentRepository().GetBalance(
		ctx,
		marketAddresses,
		m.BaseAsset,
	)
	if err != nil {
		return nil, err
	}

	quoteAssetBalance, err := t.dbManager.UnspentRepository().GetBalance(
		ctx,
		marketAddresses,
		m.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	return &BalanceWithFee{
		Balance: Balance{
			BaseAmount:  int64(baseAssetBalance),
			QuoteAmount: int64(quoteAssetBalance),
		},
		Fee: Fee{
			BasisPoint: m.Fee,
		},
	}, nil
}
