package v091webhook

const (
	TradeSettled      WebhookAction = iota //1
	AccountLowBalance                      //2
	AccountWithdraw                        //3
	AllActions                             //5
)

type WebhookAction int

type Webhook struct {
	ID         string        `json:"id"`
	ActionType WebhookAction `json:"action_type"`
	Endpoint   string        `json:"endpoint"`
	Secret     string        `json:"secret"`
}
