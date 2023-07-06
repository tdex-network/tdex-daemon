package v0webhook

// https://github.com/tdex-network/tdex-daemon/blob/master/api-spec/protobuf/tdex-daemon/v1/types.proto#L22-L27
const (
	TradeSettled = iota
	AccountLowBalance
	AccountWithdraw
	AllActions
)

type Webhook struct {
	ID         string `json:"id"`
	ActionType int    `json:"action_type"`
	Endpoint   string `json:"endpoint"`
	Secret     string `json:"secret"`
}
