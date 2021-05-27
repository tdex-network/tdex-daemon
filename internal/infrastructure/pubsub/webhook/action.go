package webhookpubsub

// webhook action types
const (
	TradeSettled WebhookAction = iota
	AccountLowBalance
	AccountWithdraw
	AllActions
)

var (
	actionToString = map[WebhookAction]string{
		TradeSettled:      "TRADE_SETTLED",
		AccountLowBalance: "ACCOUNT_LOW_BALANCE",
		AccountWithdraw:   "ACCOUNT_WITHDRAW",
		AllActions:        "*",
	}
	stringToAction = map[string]WebhookAction{
		"TRADE_SETTLED":       TradeSettled,
		"ACCOUNT_LOW_BALANCE": AccountLowBalance,
		"ACCOUNT_WITHDRAW":    AccountWithdraw,
		"*":                   AllActions,
	}
)

type WebhookAction int

func WebhookActionFromString(actionStr string) (WebhookAction, bool) {
	action, ok := stringToAction[actionStr]
	return action, ok
}

func (wa WebhookAction) String() string {
	actionStr, ok := actionToString[wa]
	if !ok {
		actionStr = "UNKNOWN"
	}
	return actionStr
}

func (wa WebhookAction) Code() int {
	return int(wa)
}

func (wa WebhookAction) Label() string {
	return wa.String()
}
