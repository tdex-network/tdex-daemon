package permissions

import (
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

const (
	EntityOperator = "operator"
	EntityTrader   = "trader"
)

func MarketPermissions() []bakery.Op {
	return []bakery.Op{
		{
			Entity: macaroons.PermissionEntityCustomURI,
			Action: "/Operator/OpenMarket",
		},
		{
			Entity: macaroons.PermissionEntityCustomURI,
			Action: "/Operator/CloseMarket",
		},
		{
			Entity: macaroons.PermissionEntityCustomURI,
			Action: "/Operator/UpdateMarketStrategy",
		},
	}
}

func PricePermissions() []bakery.Op {
	return []bakery.Op{
		{
			Entity: macaroons.PermissionEntityCustomURI,
			Action: "/Operator/UpdateMarketPrice",
		},
	}
}

func ReadOnlyPermissions() []bakery.Op {
	return []bakery.Op{
		{
			Entity: EntityOperator,
			Action: "read",
		},
	}
}

func AdminPermissions() []bakery.Op {
	return []bakery.Op{
		{
			Entity: EntityOperator,
			Action: "read",
		},
		{
			Entity: EntityOperator,
			Action: "write",
		},
	}
}

func Whitelist() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		"/Wallet/GenSeed": {{
			Entity: EntityOperator,
			Action: "read",
		}},
		"/Wallet/InitWallet": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Wallet/UnlockWallet": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Wallet/ChangePassword": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Trade/Markets": {{
			Entity: EntityTrader,
			Action: "read",
		}},
		"/Trade/Balances": {{
			Entity: EntityTrader,
			Action: "read",
		}},
		"/Trade/MarketPrice": {{
			Entity: EntityTrader,
			Action: "read",
		}},
		"/Trade/TradePropose": {{
			Entity: EntityTrader,
			Action: "write",
		}},
		"/Trade/TradeComplete": {{
			Entity: EntityTrader,
			Action: "write",
		}},
	}
}

// AllPermissionsByMethod returns a mapping of the RPC server calls to the
// permissions they require.
func AllPermissionsByMethod() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		"/Wallet/WalletAddress": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Wallet/SendToMany": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Wallet/WalletBalance": {{
			Entity: EntityOperator,
			Action: "read",
		}},
		"/Operator/DepositMarket": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/DepositFeeAccount": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/ClaimMarketDeposit": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/ClaimFeeDeposit": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/OpenMarket": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/CloseMarket": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/UpdateMarketPercentageFee": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/UpdateMarketFixedFee": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/UpdateMarketPrice": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/UpdateMarketStrategy": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/WithdrawMarket": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/DropMarket": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/ReloadUtxos": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/AddWebhook": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/RemoveWebhook": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/ListMarket": {{
			Entity: EntityOperator,
			Action: "read",
		}},
		"/Operator/ListDepositMarket": {{
			Entity: EntityOperator,
			Action: "read",
		}},
		"/Operator/BalanceFeeAccount": {{
			Entity: EntityOperator,
			Action: "read",
		}},
		"/Operator/ListTrades": {{
			Entity: EntityOperator,
			Action: "read",
		}},
		"/Operator/ListUtxos": {{
			Entity: EntityOperator,
			Action: "read",
		}},
		"/Operator/ListWebhooks": {{
			Entity: EntityOperator,
			Action: "read",
		}},
	}
}
