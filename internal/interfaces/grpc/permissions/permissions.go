package permissions

import (
	"fmt"

	"gopkg.in/macaroon-bakery.v2/bakery"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
)

const (
	EntityOperator  = "operator"
	EntityTrade     = "trade"
	EntityMarket    = "market"
	EntityPrice     = "price"
	EntityUnlocker  = "unlocker"
	EntityWallet    = "wallet"
	EntityWebhook   = "webhook"
	EntityTransport = "transport"
)

func Validate() error {
	methodsThatNeedsAuth := AllPermissionsByMethod()
	publicRoutes := Whitelist()

	unhandledMethods := findUnhandledMethods(publicRoutes, methodsThatNeedsAuth)
	if len(unhandledMethods) > 0 {
		return fmt.Errorf("unhandled permissions for following methods: %v", unhandledMethods)
	}

	return nil
}

// findUnhandledMethods returns RPC methods that are not included in public routes
// nor in routes for which invocation some kind of auth is needed
// purpose of this check is to prevent forgetting adding of new rpc methods to public/auth map
func findUnhandledMethods(publicRoutes, methodsThatNeedsAuth map[string][]bakery.Op) []string {
	result := make([]string, 0)
	allMethods := make([]string, 0)

	for _, v := range daemonv2.OperatorService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv2.OperatorService_ServiceDesc.ServiceName, v.MethodName))
	}

	for _, v := range daemonv2.WalletService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv2.WalletService_ServiceDesc.ServiceName, v.MethodName))
	}

	for _, v := range tdexv2.TradeService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", tdexv2.TradeService_ServiceDesc.ServiceName, v.MethodName))
	}

	for _, v := range tdexv2.TransportService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", tdexv2.TransportService_ServiceDesc.ServiceName, v.MethodName))
	}

	for _, v := range allMethods {
		_, ok := publicRoutes[v]
		if ok {
			continue
		}

		_, ok = methodsThatNeedsAuth[v]
		if ok {
			continue
		}

		result = append(result, v)
	}

	return result
}

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
			Action: "read",
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
			Action: "read",
		},
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
			Entity: EntityPrice,
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
			Action: "read",
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
		fmt.Sprintf("/%s/GetStatus", daemonv2.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetInfo", daemonv2.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GenSeed", daemonv2.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "read",
		}},
		fmt.Sprintf("/%s/InitWallet", daemonv2.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UnlockWallet", daemonv2.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		fmt.Sprintf("/%s/LockWallet", daemonv2.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ChangePassword", daemonv2.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListMarkets", tdexv2.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetMarketBalance", tdexv2.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetMarketPrice", tdexv2.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/PreviewTrade", tdexv2.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ProposeTrade", tdexv2.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "write",
		}},
		fmt.Sprintf("/%s/CompleteTrade", tdexv2.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "write",
		}},
		fmt.Sprintf("/%v/SupportedContentTypes", tdexv2.TransportService_ServiceDesc.ServiceName): {{
			Entity: EntityTransport,
			Action: "read",
		}},
	}
}

// AllPermissionsByMethod returns a mapping of the RPC server calls to the
// permissions they require.
func AllPermissionsByMethod() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		fmt.Sprintf("/%s/DeriveFeeAddresses", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListFeeAddresses", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetFeeBalance", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/WithdrawFee", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "write",
		}},
		fmt.Sprintf("/%s/NewMarket", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/GetMarketInfo", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/DeriveMarketAddresses", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListMarketAddresses", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/OpenMarket", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/CloseMarket", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/DropMarket", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "write",
		}},
		fmt.Sprintf("/%s/GetMarketReport", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/WithdrawMarket", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketPercentageFee", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketFixedFee", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketPrice", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityPrice,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketAssetsPrecision", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketStrategy", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/DeriveFeeFragmenterAddresses", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListFeeFragmenterAddresses", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetFeeFragmenterBalance", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/FeeFragmenterSplitFunds", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/WithdrawFeeFragmenter", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/DeriveMarketFragmenterAddresses", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListMarketFragmenterAddresses", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetMarketFragmenterBalance", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/MarketFragmenterSplitFunds", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/WithdrawMarketFragmenter", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListMarkets", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityPrice,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ListTrades", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ListDeposits", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ListWithdrawals", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ListUtxos", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "read",
		}},
		fmt.Sprintf("/%s/AddWebhook", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityWebhook,
			Action: "write",
		}},
		fmt.Sprintf("/%s/RemoveWebhook", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityWebhook,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListWebhooks", daemonv2.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityWebhook,
			Action: "read",
		}},
	}
}
