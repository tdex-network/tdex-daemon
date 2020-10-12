package application

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/uow"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	pbswap "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/pset"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TraderService interface {
	GetTradableMarkets(ctx context.Context) ([]MarketWithFee, error)
	GetMarketPrice(
		ctx context.Context,
		market Market,
		tradeType int,
		amount uint64,
	) (*PriceWithFee, error)
	TradePropose(
		req *pbtrade.TradeProposeRequest,
		stream pbtrade.Trade_TradeProposeServer,
	) error
	TradeComplete(req *pbtrade.TradeCompleteRequest,
		stream pbtrade.Trade_TradeCompleteServer) error
}

type traderService struct {
	marketRepository  domain.MarketRepository
	tradeRepository   domain.TradeRepository
	vaultRepository   domain.VaultRepository
	unspentRepository domain.UnspentRepository
	explorerSvc       explorer.Service
}

func NewTraderService(
	marketRepository domain.MarketRepository,
	tradeRepository domain.TradeRepository,
	vaultRepository domain.VaultRepository,
	unspentRepository domain.UnspentRepository,
	explorerSvc explorer.Service,
) TraderService {
	return &traderService{
		marketRepository:  marketRepository,
		tradeRepository:   tradeRepository,
		vaultRepository:   vaultRepository,
		unspentRepository: unspentRepository,
		explorerSvc:       explorerSvc,
	}
}

// Markets is the domain controller for the Markets RPC
func (t *traderService) GetTradableMarkets(ctx context.Context) (
	[]MarketWithFee,
	error,
) {
	tradableMarkets, err := t.marketRepository.GetTradableMarkets(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	marketsWithFee := make([]MarketWithFee, 0)
	for _, mkt := range tradableMarkets {
		marketsWithFee = append(marketsWithFee, MarketWithFee{
			Market: Market{
				BaseAsset:  mkt.BaseAssetHash(),
				QuoteAsset: mkt.QuoteAssetHash(),
			},
			Fee: Fee{
				FeeAsset:   mkt.FeeAsset(),
				BasisPoint: mkt.Fee(),
			},
		})
	}

	return marketsWithFee, nil
}

// MarketPrice is the domain controller for the MarketPrice RPC.
func (t *traderService) GetMarketPrice(
	ctx context.Context,
	market Market,
	tradeType int,
	amount uint64,
) (*PriceWithFee, error) {

	// Checks if base asset is correct
	if market.BaseAsset != config.GetString(config.BaseAssetKey) {
		return nil, domain.ErrMarketNotExist
	}
	//Checks if market exist
	mkt, _, err := t.marketRepository.GetMarketByAsset(
		ctx,
		market.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	if !mkt.IsTradable() {
		return nil, domain.ErrMarketIsClosed
	}

	price, previewAmount, err := getPriceAndPreviewForMarket(
		ctx, t.vaultRepository, t.unspentRepository,
		mkt, tradeType, amount,
	)
	if err != nil {
		return nil, err
	}

	return &PriceWithFee{
		Price: price,
		Fee: Fee{
			FeeAsset:   mkt.FeeAsset(),
			BasisPoint: mkt.Fee(),
		},
		Amount: previewAmount,
	}, nil
}

// TradePropose is the domain controller for the TradePropose RPC
func (t *traderService) TradePropose(
	req *pbtrade.TradeProposeRequest,
	stream pbtrade.Trade_TradeProposeServer,
) error {
	_, marketAccountIndex, err := t.marketRepository.GetMarketByAsset(
		context.Background(),
		req.GetMarket().GetQuoteAsset(),
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// get all unspents for market account along with private blinding keys and
	// signing derivation paths for respectively unblinding and signing them later
	marketUnspents, marketBlindingKeysByScript, marketDerivationPaths,
		err := t.getUnspentsBlindingsAndDerivationPathsForAccount(marketAccountIndex)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// ... and the same for fee account (we'll need to top-up fees)
	feeUnspents, feeBlindingKeysByScript, feeDerivationPaths, err :=
		t.getUnspentsBlindingsAndDerivationPathsForAccount(domain.FeeAccount)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	var reply *pbtrade.TradeProposeReply

	// try to accept the incoming proposal in a transactional way, by committing
	// changes to different storages only if the trade is accepted at the very end.
	// This process causes changes that affect different domains so we need to
	// update all or none of them in case any errors occur.
	unit := uow.NewUnitOfWork(t.tradeRepository, t.unspentRepository)

	if err := unit.Run(func(u uow.Contextual) error {
		var mnemonic []string
		var tradeID uuid.UUID
		var selectedUnspents []explorer.Utxo
		var outputBlindingKeyByScript map[string][]byte
		var outputDerivationPath, changeDerivationPath, feeChangeDerivationPath string

		// derive output and change address for market, and change address for fee account
		vaultCtx := u.Context(t.vaultRepository)
		if err := t.vaultRepository.UpdateVault(
			vaultCtx,
			nil,
			"",
			func(v *domain.Vault) (*domain.Vault, error) {
				mnemonic, err = v.Mnemonic()
				if err != nil {
					return nil, err
				}
				outputAddress, outputScript, _, err := v.DeriveNextExternalAddressForAccount(marketAccountIndex)
				if err != nil {
					return nil, err
				}
				_, changeScript, _, err := v.DeriveNextInternalAddressForAccount(marketAccountIndex)
				if err != nil {
					return nil, err
				}
				_, feeChangeScript, _,
					err := v.DeriveNextInternalAddressForAccount(domain.FeeAccount)
				if err != nil {
					return nil, err
				}
				marketAccount, _ := v.AccountByIndex(marketAccountIndex)
				feeAccount, _ := v.AccountByIndex(domain.FeeAccount)

				outputBlindingKeyByScript = blindingKeyByScriptFromCTAddress(outputAddress)
				outputDerivationPath, _ = marketAccount.DerivationPathByScript(outputScript)
				changeDerivationPath, _ = marketAccount.DerivationPathByScript(changeScript)
				feeChangeDerivationPath, _ = feeAccount.DerivationPathByScript(feeChangeScript)

				return v, nil
			}); err != nil {
			return err
		}

		tradeCtx := u.Context(t.tradeRepository)
		// parse swap proposal and possibly accept
		if err := t.tradeRepository.UpdateTrade(
			tradeCtx,
			nil,
			func(t *domain.Trade) (*domain.Trade, error) {
				if err := t.Propose(req.GetSwapRequest(), req.GetMarket().GetQuoteAsset(), nil); err != nil {
					return nil, err
				}
				tradeID = t.ID()

				acceptSwapResult, err := acceptSwap(acceptSwapOpts{
					mnemonic:                   mnemonic,
					swapRequest:                req.GetSwapRequest(),
					marketUnspents:             marketUnspents,
					feeUnspents:                feeUnspents,
					marketBlindingKeysByScript: marketBlindingKeysByScript,
					feeBlindingKeysByScript:    feeBlindingKeysByScript,
					outputBlindingKeyByScript:  outputBlindingKeyByScript,
					marketDerivationPaths:      marketDerivationPaths,
					feeDerivationPaths:         feeDerivationPaths,
					outputDerivationPath:       outputDerivationPath,
					changeDerivationPath:       changeDerivationPath,
					feeChangeDerivationPath:    feeChangeDerivationPath,
				})
				if err != nil {
					return nil, err
				}

				if err := t.Accept(
					acceptSwapResult.psetBase64,
					acceptSwapResult.inputBlindingKeys,
					acceptSwapResult.outputBlindingKeys,
				); err != nil {
					return nil, err
				}

				reply = &pbtrade.TradeProposeReply{
					SwapAccept:     t.SwapAcceptMessage(),
					ExpiryTimeUnix: t.SwapExpiryTime(),
				}

				return t, nil
			}); err != nil {
			return err
		}

		selectedUnspentKeys := getUnspentKeys(selectedUnspents)
		unspentCtx := u.Context(t.unspentRepository)
		if err := t.unspentRepository.LockUnspents(
			unspentCtx,
			selectedUnspentKeys,
			tradeID,
		); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err := stream.Send(reply); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

// TradeComplete is the domain controller for the TradeComplete RPC
func (t *traderService) TradeComplete(req *pbtrade.TradeCompleteRequest,
	stream pbtrade.Trade_TradeCompleteServer) error {
	ctx := context.Background()
	currentTrade, err := t.tradeRepository.GetTradeBySwapAcceptID(
		ctx,
		req.GetSwapComplete().GetAcceptId(),
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	var reply *pbtrade.TradeCompleteReply
	tradeID := currentTrade.ID()
	if err := t.tradeRepository.UpdateTrade(
		ctx,
		&tradeID,
		func(tr *domain.Trade) (*domain.Trade, error) {
			psetBase64 := req.GetSwapComplete().GetTransaction()
			opts := wallet.FinalizeAndExtractTransactionOpts{
				PsetBase64: psetBase64,
			}
			txHex, txID, err := wallet.FinalizeAndExtractTransaction(opts)
			if err != nil {
				return nil, err
			}

			if err := tr.Complete(psetBase64, txID); err != nil {
				return nil, err
			}

			if _, err := t.explorerSvc.BroadcastTransaction(txHex); err != nil {
				return nil, err
			}

			reply = &pbtrade.TradeCompleteReply{
				Txid: txID,
			}
			return tr, nil
		}); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err := stream.Send(reply); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (t *traderService) getUnspentsBlindingsAndDerivationPathsForAccount(
	account int,
) (
	[]explorer.Utxo,
	map[string][]byte,
	map[string]string,
	error,
) {
	derivedAddresses, blindingKeys, err := t.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(context.Background(), account)
	if err != nil {
		return nil, nil, nil, err
	}

	scripts := make([]string, 0, len(derivedAddresses))
	for _, addr := range derivedAddresses {
		script, _ := address.ToOutputScript(addr, *config.GetNetwork())
		scripts = append(scripts, hex.EncodeToString(script))
	}
	derivationPaths, _ := t.vaultRepository.GetDerivationPathByScript(
		context.Background(),
		account,
		scripts,
	)

	unspents := t.unspentRepository.GetAvailableUnspentsForAddresses(
		context.Background(),
		derivedAddresses,
	)

	utxos, err := t.getUtxos(derivedAddresses, blindingKeys)
	if err != nil {
		return nil, nil, nil, err
	}

	availableUtxos := make([]explorer.Utxo, 0, len(unspents))
	for _, unspent := range unspents {
		unspentKey := unspent.Key()
		for _, utxo := range utxos {
			if unspentKey.TxID == utxo.Hash() && unspentKey.VOut == utxo.Index() {
				availableUtxos = append(availableUtxos, utxo)
				break
			}
		}
	}

	blindingKeysByScript := map[string][]byte{}
	for i, addr := range derivedAddresses {
		script, _ := address.ToOutputScript(addr, *config.GetNetwork())
		blindingKeysByScript[hex.EncodeToString(script)] = blindingKeys[i]
	}

	return availableUtxos, blindingKeysByScript, derivationPaths, nil
}

func (t *traderService) getUtxos(
	addresses []string,
	blindingKeys [][]byte,
) ([]explorer.Utxo, error) {
	chUnspents := make(chan []explorer.Utxo)
	chErr := make(chan error, 1)
	unspents := make([]explorer.Utxo, 0)

	for _, addr := range addresses {
		go t.getUtxosForAddress(addr, blindingKeys, chUnspents, chErr)

		select {
		case err := <-chErr:
			close(chErr)
			close(chUnspents)
			return nil, err
		case unspentsForAddress := <-chUnspents:
			unspents = append(unspents, unspentsForAddress...)
		}
	}

	return unspents, nil
}

func (t *traderService) getUtxosForAddress(addr string, blindingKeys [][]byte,
	chUnspents chan []explorer.Utxo, chErr chan error) {
	unspents, err := t.explorerSvc.GetUnspents(addr, blindingKeys)
	if err != nil {
		chErr <- err
		return
	}
	chUnspents <- unspents
}

func blindingKeyByScriptFromCTAddress(addr string) map[string][]byte {
	script, _ := address.ToOutputScript(addr, *config.GetNetwork())
	blech32, _ := address.FromBlech32(addr)
	return map[string][]byte{
		hex.EncodeToString(script): blech32.PublicKey,
	}
}

type acceptSwapOpts struct {
	mnemonic                   []string
	swapRequest                *pbswap.SwapRequest
	marketUnspents             []explorer.Utxo
	feeUnspents                []explorer.Utxo
	marketBlindingKeysByScript map[string][]byte
	feeBlindingKeysByScript    map[string][]byte
	outputBlindingKeyByScript  map[string][]byte
	marketDerivationPaths      map[string]string
	feeDerivationPaths         map[string]string
	outputDerivationPath       string
	changeDerivationPath       string
	feeChangeDerivationPath    string
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
	network := config.GetNetwork()

	// fill swap request transaction with daemon's inputs and outputs
	psetBase64, selectedUnspentsForSwap, err := w.UpdateSwapTx(wallet.UpdateSwapTxOpts{
		PsetBase64:           opts.swapRequest.GetTransaction(),
		Unspents:             opts.marketUnspents,
		InputAmount:          opts.swapRequest.GetAmountP(),
		InputAsset:           opts.swapRequest.GetAssetP(),
		OutputAmount:         opts.swapRequest.GetAmountR(),
		OutputAsset:          opts.swapRequest.GetAssetR(),
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
		opts.outputBlindingKeyByScript,
		psetWithFeesResult.ChangeOutputsBlindingKeys,
		opts.swapRequest.GetOutputBlindingKey(),
	)

	// blind the transaction
	blindedPset, err := w.BlindSwapTransaction(wallet.BlindSwapTransactionOpts{
		PsetBase64:         psetWithFeesResult.PsetBase64,
		InputBlindingKeys:  inputBlindingKeys,
		OutputBlindingKeys: outputBlindingKeys,
	})

	// add the explicit fee output to the tx
	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    transactionutil.NewFeeOutput(psetWithFeesResult.FeeAmount),
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
	ctx context.Context,
	vaultRepo domain.VaultRepository,
	unspentRepo domain.UnspentRepository,
	market *domain.Market,
	tradeType int,
	amount uint64,
) (
	price Price,
	previewAmount uint64,
	err error,
) {
	if market.IsStrategyPluggable() {
		previewAmount = calcPreviewAmount(market, tradeType, amount)
		price = Price{
			BasePrice:  market.BaseAssetPrice(),
			QuotePrice: market.QuoteAssetPrice(),
		}
		return
	}

	addresses, _, err := vaultRepo.GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, market.AccountIndex())
	if err != nil {
		return
	}

	unspents := unspentRepo.GetAvailableUnspentsForAddresses(ctx, addresses)
	if len(unspents) == 0 {
		err = errors.New("no available funds for market")
		return
	}

	price, previewAmount = previewFromFormula(unspents, market, tradeType, amount)
	return
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
func calcPreviewAmount(market *domain.Market, tradeType int, amount uint64) uint64 {
	if tradeType == TradeBuy {
		return calcProposeAmount(
			amount,
			market.Fee(),
			market.QuoteAssetPrice(),
		)
	}

	return calcExpectedAmount(
		amount,
		market.Fee(),
		market.QuoteAssetPrice(),
	)
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
) uint64 {
	feePercentage := decimal.NewFromInt(feeAmount).Div(decimal.NewFromInt(100))
	amountR := decimal.NewFromInt(int64(amount))

	// amountP = amountR * price * (1 + feePercentage)
	amountP := amountR.Mul(price).Mul(decimal.NewFromInt(1).Add(feePercentage))
	return amountP.BigInt().Uint64()
}

func calcExpectedAmount(
	amount uint64,
	feeAmount int64,
	price decimal.Decimal,
) uint64 {
	feePercentage := decimal.NewFromInt(feeAmount).Div(decimal.NewFromInt(100))
	amountP := decimal.NewFromInt(int64(amount))

	// amountR = amountP + price * (1 - feePercentage)
	amountR := amountP.Mul(price).Mul(decimal.NewFromInt(1).Sub(feePercentage))
	return amountR.BigInt().Uint64()
}

func previewFromFormula(unspents []domain.Unspent, market *domain.Market, tradeType int, amount uint64) (Price, uint64) {
	balances := getBalanceByAsset(unspents)
	baseBalanceAvailable := balances[market.BaseAssetHash()]
	quoteBalanceAvailable := balances[market.QuoteAssetHash()]
	formula := market.Strategy().Formula()

	price := Price{
		BasePrice: formula.SpotPrice(&mm.FormulaOpts{
			BalanceIn:  quoteBalanceAvailable,
			BalanceOut: baseBalanceAvailable,
		}),
		QuotePrice: formula.SpotPrice(&mm.FormulaOpts{
			BalanceIn:  baseBalanceAvailable,
			BalanceOut: quoteBalanceAvailable,
		}),
	}

	if tradeType == TradeBuy {
		previewAmount := formula.InGivenOut(
			&mm.FormulaOpts{
				BalanceIn:           quoteBalanceAvailable,
				BalanceOut:          baseBalanceAvailable,
				Fee:                 uint64(market.Fee()),
				ChargeFeeOnTheWayIn: market.FeeAsset() == market.BaseAssetHash(),
			},
			amount,
		)
		return price, previewAmount
	}
	previewAmount := formula.OutGivenIn(
		&mm.FormulaOpts{
			BalanceIn:           baseBalanceAvailable,
			BalanceOut:          quoteBalanceAvailable,
			Fee:                 uint64(market.Fee()),
			ChargeFeeOnTheWayIn: market.FeeAsset() == market.QuoteAssetHash(),
		},
		amount,
	)
	return price, previewAmount
}
