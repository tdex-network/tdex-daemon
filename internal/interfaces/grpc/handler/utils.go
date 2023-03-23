package grpchandler

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/shopspring/decimal"
	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/psetv2"
)

func parsePassword(pwd string) (string, error) {
	if len(pwd) <= 0 {
		return "", errors.New("missing password")
	}
	if !isValidPassword(pwd) {
		return "", errors.New("invalid password")
	}
	return pwd, nil
}

func parseMarket(market *tdexv2.Market) (ports.Market, error) {
	if market == nil {
		return nil, fmt.Errorf("missing market")
	}
	if !isValidAsset(market.GetBaseAsset()) {
		return nil, errors.New("invalid base asset")
	}
	if !isValidAsset(market.GetQuoteAsset()) {
		return nil, errors.New("invalid quote asset")
	}

	return market, nil
}

func parseMarketFee(fee *tdexv2.MarketFee) (int64, int64, error) {
	if fee == nil {
		return 0, 0, nil
	}
	baseFee, quoteFee := fee.GetBaseAsset(), fee.GetQuoteAsset()
	if baseFee < -1 {
		return -1, -1, errors.New("invalid market base fee value")
	}
	if quoteFee < -1 {
		return -1, -1, errors.New("invalid market quote fee value")
	}
	if baseFee == -1 && baseFee == quoteFee {
		return -1, -1, errors.New("invalid market base and quote fee values")
	}
	return baseFee, quoteFee, nil
}

func parseStrategy(strategyType daemonv2.StrategyType) (int, error) {
	switch strategyType {
	case daemonv2.StrategyType_STRATEGY_TYPE_BALANCED:
		return domain.StrategyTypeBalanced, nil
	case daemonv2.StrategyType_STRATEGY_TYPE_PLUGGABLE:
		return domain.StrategyTypePluggable, nil
	case daemonv2.StrategyType_STRATEGY_TYPE_UNSPECIFIED:
		return -1, errors.New("invalid strategy type")
	default:
		return -1, errors.New("unknown strategy type")
	}
}

func parsePrecision(precision uint32) (uint, error) {
	if precision > 8 {
		return 0, fmt.Errorf("asset precision must be in range [0, 8]")
	}
	return uint(precision), nil
}

func parsePrice(price *tdexv2.Price) (*decimal.Decimal, *decimal.Decimal, error) {
	if price == nil {
		return nil, nil, errors.New("missing market price")
	}
	if !isValidPrice(price.GetBasePrice()) {
		return nil, nil, errors.New("invalid base price")
	}
	if !isValidPrice(price.GetQuotePrice()) {
		return nil, nil, errors.New("invalid base price")
	}
	basePrice := decimal.NewFromFloat(price.GetBasePrice())
	quotePrice := decimal.NewFromFloat(price.GetQuotePrice())
	return &basePrice, &quotePrice, nil
}

func parseOutputs(outs []*daemonv2.TxOutput) ([]ports.TxOutput, error) {
	list := make([]ports.TxOutput, 0)
	for i, o := range outs {
		if !isValidAsset(o.GetAsset()) {
			return nil, fmt.Errorf("invalid asset for output %d", i)
		}
		if !isValidScript(o.GetScript()) {
			return nil, fmt.Errorf("invalid address for output %d", i)
		}
		if !isValidBlindKey(o.GetBlindingKey()) {
			return nil, fmt.Errorf("invalid address for output %d", i)
		}
		if !isValidAmount(o.GetAmount()) {
			return nil, fmt.Errorf("invalid amount for outut %d", i)
		}
		list = append(list, o)
	}
	return list, nil
}

func parseAccountName(account string) (string, error) {
	if !isValidAccount(account) {
		return "", errors.New("missing account name")
	}
	return account, nil
}

func parseSwapRequest(
	sr *tdexv2.SwapRequest, feeAsset string, feeAmount uint64,
) (ports.SwapRequest, error) {
	if sr == nil {
		return nil, fmt.Errorf("missing swap request")
	}
	if !isValidAmount(sr.GetAmountP()) {
		return nil, fmt.Errorf("invalid swap request amount proposed")
	}
	if !isValidAsset(sr.GetAssetP()) {
		return nil, fmt.Errorf("invalid swap request asset proposed")
	}
	if !isValidAmount(sr.GetAmountR()) {
		return nil, fmt.Errorf("invalid swap request amount received")
	}
	if !isValidAsset(sr.GetAssetR()) {
		return nil, fmt.Errorf("invalid swap request asset received")
	}
	if !isValidTransaction(sr.GetTransaction()) {
		return nil, fmt.Errorf("invalid swap request transaction")
	}
	if !isValidUnblindedInputList(sr.GetUnblindedInputs()) {
		return nil, fmt.Errorf("invalid unblinded input(s)")
	}
	if !isValidAsset(feeAsset) {
		return nil, fmt.Errorf("invalid fee asset")
	}
	if !isValidAmount(feeAmount) {
		return nil, fmt.Errorf("invalid fee amount")
	}
	return swapRequestInfo{sr, feeAsset, feeAmount}, nil
}

func parseTradeType(tradeType tdexv2.TradeType) (ports.TradeType, error) {
	if tradeType != tdexv2.TradeType_TRADE_TYPE_BUY &&
		tradeType != tdexv2.TradeType_TRADE_TYPE_SELL {
		return nil, fmt.Errorf("unknown trade type")
	}
	return tradeTypeInfo(tradeType), nil
}

func parseAmount(amount uint64) (uint64, error) {
	if !isValidAmount(amount) {
		return 0, fmt.Errorf("invalid amount")
	}
	return amount, nil
}

func parseAsset(asset string) (string, error) {
	if len(asset) <= 0 {
		return "", fmt.Errorf("missing asset")
	}
	if !isValidAsset(asset) {
		return "", fmt.Errorf("invalid asset")
	}
	return asset, nil
}

func parseNumOfAddresses(num int64) (int, error) {
	if num < 0 {
		return -1, errors.New("invalid number of derived addresses")
	}
	if num == 0 {
		return 1, nil
	}
	return int(num), nil
}

func parseMaxFragments(num uint32) (uint32, error) {
	if int(num) < 0 {
		return 0, errors.New("invalid max number of fragments")
	}
	return num, nil
}

func parseMillisatsPerByte(msatsPerByte uint64) (uint64, error) {
	if int(msatsPerByte) < 0 {
		return 0, errors.New("invalid mSats/vByte value")
	}
	return msatsPerByte, nil
}

func parsePercentageFee(bp uint32) (uint32, error) {
	if int(bp) < 0 {
		return 0, errors.New("invalid percentage fee")
	}
	return bp, nil
}

func parseMnemonic(mnemonic []string) ([]string, error) {
	if len(mnemonic) <= 0 {
		return nil, fmt.Errorf("missing mnemonic")
	}
	return mnemonic, nil
}

func parsePage(page *daemonv2.Page) (ports.Page, error) {
	if page == nil {
		return nil, nil
	}
	if page.GetNumber() < 0 {
		return nil, errors.New("invalid page number")
	}
	if page.GetSize() <= 0 {
		return nil, errors.New("invalid page size")
	}
	return pageInfo{page}, nil
}

func parseWebhook(hook *daemonv2.AddWebhookRequest) (ports.Webhook, error) {
	if _, err := parseWebhookActionType(hook.GetAction()); err != nil {
		return nil, err
	}
	if len(hook.GetEndpoint()) <= 0 {
		return nil, errors.New("missing webhook endpoint")
	}
	return webhookInfo{hook}, nil
}

func parseWebhookActionType(actionType daemonv2.ActionType) (int, error) {
	if actionType <= daemonv2.ActionType_ACTION_TYPE_UNSPECIFIED {
		return -1, errors.New("invalid action type")
	}
	return int(actionType), nil
}

func parseTimeRange(timeRange *daemonv2.TimeRange) (ports.TimeRange, error) {
	pp := timeRange.GetPredefinedPeriod()
	cp := timeRange.GetCustomPeriod()
	if pp == daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_UNSPECIFIED && cp == nil {
		return nil, errors.New("missing predefined or custom period")
	}
	if pp < daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_UNSPECIFIED ||
		pp > daemonv2.PredefinedPeriod_PREDEFINED_PERIOD_ALL {
		return nil, errors.New("unknown predefined period")
	}
	if cp != nil {
		startDate, endDate := cp.GetStartDate(), cp.GetEndDate()
		if startDate == "" && endDate == "" {
			return nil, errors.New("missing custom start and end dates")
		}
		sd, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid custom start date: %s", err)
		}
		ed, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid custom end date: %s", err)
		}
		if ed.Equal(sd) || ed.Before(sd) {
			return nil, fmt.Errorf(
				"invalid custom period: end date is equal or before start date",
			)
		}
	}

	return timeRangeInfo{timeRange}, nil
}

func parseTimeFrame(timeFrame daemonv2.TimeFrame) (int, error) {
	switch timeFrame {
	case daemonv2.TimeFrame_TIME_FRAME_UNSPECIFIED, daemonv2.TimeFrame_TIME_FRAME_HOUR:
		return 1, nil
	case daemonv2.TimeFrame_TIME_FRAME_FOUR_HOURS:
		return 4, nil
	case daemonv2.TimeFrame_TIME_FRAME_DAY:
		return 24, nil
	case daemonv2.TimeFrame_TIME_FRAME_WEEK:
		return 24 * 7, nil
	case daemonv2.TimeFrame_TIME_FRAME_MONTH:
		year, month, _ := time.Now().Date()
		numOfDaysForCurrentMont := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		return numOfDaysForCurrentMont, nil
	default:
		return -1, fmt.Errorf("unknown time frame")
	}
}

func isValidPassword(pwd string) bool {
	return len(pwd) >= 8
}

func isValidAsset(asset string) bool {
	b, err := hex.DecodeString(asset)
	if err != nil {
		return false
	}
	return len(b) == 32
}

func isValidPrice(price float64) bool {
	return price > 0
}

func isValidScript(script string) bool {
	if len(script) <= 0 {
		return true
	}
	buf, err := hex.DecodeString(script)
	if err != nil {
		return false
	}
	_, err = address.ParseScript(buf)
	return err == nil
}

func isValidBlindKey(blindKey string) bool {
	if len(blindKey) <= 0 {
		return true
	}

	buf, err := hex.DecodeString(blindKey)
	if err != nil {
		return false
	}
	_, err = btcec.ParsePubKey(buf)
	return err == nil
}

func isValidAmount(amount uint64) bool {
	return int64(amount) >= 0
}

func isValidAccount(account string) bool {
	return len(account) > 0
}

func isValidTransaction(tx string) bool {
	_, err := psetv2.NewPsetFromBase64(tx)
	return err == nil
}

func isValidUnblindedInputList(list []*tdexv2.UnblindedInput) bool {
	for _, in := range list {
		if !isValidIndex(in.GetIndex()) || !isValidAsset(in.GetAsset()) ||
			!isValidAmount(in.GetAmount()) || !isValidAsset(in.GetAssetBlinder()) ||
			!isValidAsset(in.GetAmountBlinder()) {
			return false
		}
	}
	return true
}

func isValidIndex(index uint32) bool {
	return int(index) >= 0
}

func timestampToString(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	return time.Unix(timestamp, 0).Format(time.RFC3339)
}
