package application

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

// checkFeeAndMarketBalances helper to check for low balances for fee and
// market accounts and to publish AccountLowBalance topic on pubsub just in case.
func checkFeeAndMarketBalances(
	repoManager ports.RepoManager, pubsub ports.SecurePubSub,
	ctx context.Context, mkt *domain.Market, lbtcAsset string, feeThreshold uint64,
) {
	// Get balance including locked unspents.
	feeBalance, err := getBalanceForFee(repoManager, ctx, lbtcAsset, false)
	if err == nil && feeBalance <= feeThreshold {
		account := map[string]interface{}{
			"type":  "fee",
			"index": domain.FeeAccount,
		}
		if err := publishAccountLowBalanceTopic(
			pubsub, account, feeBalance,
		); err != nil {
			log.Warn(err)
		}
	}

	// Get balance including locked unspents.
	mktBalance, err := getBalanceForMarket(repoManager, ctx, mkt, false)
	if err == nil {
		lowBalance := mktBalance.BaseAmount <= uint64(mkt.FixedFee.BaseFee) ||
			mktBalance.QuoteAmount <= uint64(mkt.FixedFee.QuoteFee)
		if lowBalance {
			account := map[string]interface{}{
				"type":        "market",
				"base_asset":  mkt.BaseAsset,
				"quote_asset": mkt.QuoteAsset,
				"index":       mkt.AccountIndex,
			}
			if err := publishAccountLowBalanceTopic(
				pubsub, account, mktBalance,
			); err != nil {
				log.Warn(err)
			}
		}
	}
}

// publishAccountLowBalanceTopic helper to publish an AccountLowBalance topic
// on the given pubsub service.
func publishAccountLowBalanceTopic(
	pubsub ports.SecurePubSub, account, balance interface{},
) error {
	if pubsub == nil {
		return nil
	}

	topics := pubsub.TopicsByCode()
	topic := topics[AccountLowBalance]
	payload := map[string]interface{}{
		"account": account,
		"balance": balance,
	}
	message, _ := json.Marshal(payload)

	if err := pubsub.Publish(topic.Label(), string(message)); err != nil {
		return fmt.Errorf(
			"an error occured while publishing message for topic %s: %s",
			topic.Label(), err,
		)
	}
	return nil
}

func publishMarketWithdrawTopic(
	pubsub ports.SecurePubSub,
	mkt Market, mktBalance, withdrewBalance Balance,
	destAddress, txid string,
) error {
	if pubsub == nil {
		return nil
	}

	baseBalance := mktBalance.BaseAmount - withdrewBalance.BaseAmount
	quoteBalance := mktBalance.QuoteAmount - withdrewBalance.QuoteAmount

	payload := map[string]interface{}{
		"market": map[string]string{
			"base_asset":  mkt.BaseAsset,
			"quote_asset": mkt.QuoteAsset,
		},
		"amount_withdraw": map[string]interface{}{
			"base_amount":  withdrewBalance.BaseAmount,
			"quote_amount": withdrewBalance.QuoteAmount,
		},
		"receiving_address": destAddress,
		"txid":              txid,
		"balance": map[string]uint64{
			"base_balance":  baseBalance,
			"quote_balance": quoteBalance,
		},
	}
	message, _ := json.Marshal(payload)
	topics := pubsub.TopicsByCode()
	topic := topics[AccountWithdraw]
	if err := pubsub.Publish(topic.Label(), string(message)); err != nil {
		log.WithError(err).Warnf(
			"an error occured while publishing message for topic %s",
			topic.Label(),
		)
	}
	return nil
}

func publishFeeWithdrawTopic(
	pubsub ports.SecurePubSub,
	balance, withdrewBalance uint64,
	destAddress, txid, lbtcAsset string,
) error {
	if pubsub == nil {
		return nil
	}

	lbtcBalance := balance - withdrewBalance

	payload := map[string]interface{}{
		"fee": map[string]string{
			"lbtc_asset": lbtcAsset,
		},
		"amount_withdraw": map[string]interface{}{
			"lbtc_amount": withdrewBalance,
		},
		"receiving_address": destAddress,
		"txid":              txid,
		"balance": map[string]uint64{
			"lbtc_balance": lbtcBalance,
		},
	}
	message, _ := json.Marshal(payload)
	topics := pubsub.TopicsByCode()
	topic := topics[AccountWithdraw]
	if err := pubsub.Publish(topic.Label(), string(message)); err != nil {
		log.WithError(err).Warnf(
			"an error occured while publishing message for topic %s",
			topic.Label(),
		)
	}
	return nil
}
