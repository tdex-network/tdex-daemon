package permissions

import (
	"gopkg.in/macaroon-bakery.v2/bakery"
)

const (
	EntityOperator = "operator"
	EntityTrade    = "trade"
	EntityMarket   = "market"
	EntityPrice    = "price"
	EntityUnlocker = "unlocker"
	EntityWallet   = "wallet"
	EntityWebhook  = "webhook"
)

// MarketPermissions returns the permissions of the macaroon market.macaroon.
// This grants access to all actions for the market and price entities.
func MarketPermissions() []bakery.Op {
	return []bakery.Op{
		{
			Entity: EntityMarket,
			Action: "read",
		},
		{
			Entity: EntityMarket,
			Action: "write",
		},
		{
			Entity: EntityPrice,
			Action: "write",
		},
	}
}

// PricePermissions returns the permissions of the macaroon price.macaroon.
// This grants access to all actions for the price entity.
func PricePermissions() []bakery.Op {
	return []bakery.Op{
		{
			Entity: EntityPrice,
			Action: "write",
		},
	}
}

// ReadOnlyPermissions returns the permissions of the macaroon readonly.macaroon.
// This grants access to the read action for all entities.
func ReadOnlyPermissions() []bakery.Op {
	return []bakery.Op{
		{
			Entity: EntityOperator,
			Action: "read",
		},
		{
			Entity: EntityMarket,
			Action: "read",
		},
		{
			Entity: EntityWallet,
			Action: "read",
		},
		{
			Entity: EntityWebhook,
			Action: "read",
		},
	}
}

// WalletPermissions returns the permissions of the macaroon wallet.macaroon.
// This grants access to the all actions for the wallet entity.
func WalletPermissions() []bakery.Op {
	return []bakery.Op{
		{
			Entity: EntityWallet,
			Action: "read",
		},
		{
			Entity: EntityWallet,
			Action: "write",
		},
	}
}

// WebhookPermissions returns the permissions of the macaroon webhook.macaroon.
// This grants access to the all actions for the webhook entity.
func WebhookPermissions() []bakery.Op {
	return []bakery.Op{
		{
			Entity: EntityWebhook,
			Action: "read",
		},
		{
			Entity: EntityWebhook,
			Action: "write",
		},
	}
}

// AdminPermissions returns the permissions of the macaroon admin.macaroon.
// This grants access to the all actions for all entities.
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
		{
			Entity: EntityMarket,
			Action: "read",
		},
		{
			Entity: EntityMarket,
			Action: "write",
		},
		{
			Entity: EntityPrice,
			Action: "write",
		},
		{
			Entity: EntityWebhook,
			Action: "read",
		},
		{
			Entity: EntityWebhook,
			Action: "write",
		},
		{
			Entity: EntityWallet,
			Action: "read",
		},
		{
			Entity: EntityWallet,
			Action: "write",
		},
	}
}

// Whitelist returns the list of all whitelisted methods with the relative
// entity and action.
func Whitelist() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		"/WalletUnlocker/IsReady": {{
			Entity: EntityUnlocker,
			Action: "read",
		}},
		"/WalletUnlocker/GenSeed": {{
			Entity: EntityUnlocker,
			Action: "read",
		}},
		"/WalletUnlocker/InitWallet": {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		"/WalletUnlocker/UnlockWallet": {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		"/WalletUnlocker/ChangePassword": {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		"/Trade/Markets": {{
			Entity: EntityTrade,
			Action: "read",
		}},
		"/Trade/Balances": {{
			Entity: EntityTrade,
			Action: "read",
		}},
		"/Trade/MarketPrice": {{
			Entity: EntityTrade,
			Action: "read",
		}},
		"/Trade/TradePropose": {{
			Entity: EntityTrade,
			Action: "write",
		}},
		"/Trade/TradeComplete": {{
			Entity: EntityTrade,
			Action: "write",
		}},
	}
}

// AllPermissionsByMethod returns a mapping of the RPC server calls to the
// permissions they require.
func AllPermissionsByMethod() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		"/Wallet/WalletAddress": {{
			Entity: EntityWallet,
			Action: "write",
		}},
		"/Wallet/SendToMany": {{
			Entity: EntityWallet,
			Action: "write",
		}},
		"/Wallet/WalletBalance": {{
			Entity: EntityWallet,
			Action: "read",
		}},
		"/Operator/DepositMarket": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/DepositFeeAccount": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/ClaimMarketDeposit": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/ClaimFeeDeposit": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/OpenMarket": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/CloseMarket": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/UpdateMarketPercentageFee": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/UpdateMarketFixedFee": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/UpdateMarketStrategy": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/ListMarket": {{
			Entity: EntityMarket,
			Action: "read",
		}},
		"/Operator/ListDepositMarket": {{
			Entity: EntityMarket,
			Action: "read",
		}},
		"/Operator/BalanceFeeAccount": {{
			Entity: EntityMarket,
			Action: "read",
		}},
		"/Operator/DropMarket": {{
			Entity: EntityMarket,
			Action: "write",
		}},
		"/Operator/UpdateMarketPrice": {{
			Entity: EntityPrice,
			Action: "write",
		}},
		"/Operator/WithdrawMarket": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/ReloadUtxos": {{
			Entity: EntityOperator,
			Action: "write",
		}},
		"/Operator/ListTrades": {{
			Entity: EntityOperator,
			Action: "read",
		}},
		"/Operator/ListUtxos": {{
			Entity: EntityOperator,
			Action: "read",
		}},
		"/Operator/AddWebhook": {{
			Entity: EntityWebhook,
			Action: "write",
		}},
		"/Operator/RemoveWebhook": {{
			Entity: EntityWebhook,
			Action: "write",
		}},
		"/Operator/ListWebhooks": {{
			Entity: EntityWebhook,
			Action: "read",
		}},
	}
}
