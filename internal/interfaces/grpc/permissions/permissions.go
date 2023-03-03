package permissions

import (
	"fmt"

	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	"gopkg.in/macaroon-bakery.v2/bakery"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	tdexold "github.com/tdex-network/tdex-protobuf/generated/go/trade"
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
	Reflection      = "reflection"
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

	for _, v := range daemonv1.OperatorService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv1.OperatorService_ServiceDesc.ServiceName, v.MethodName))
	}

	for _, v := range daemonv1.WalletService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv1.WalletService_ServiceDesc.ServiceName, v.MethodName))
	}

	for _, v := range daemonv1.WalletUnlockerService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv1.WalletUnlockerService_ServiceDesc.ServiceName, v.MethodName))
	}

	for _, v := range tdexv1.TradeService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", tdexv1.TradeService_ServiceDesc.ServiceName, v.MethodName))
	}

	for _, v := range tdexv1.TransportService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", tdexv1.TransportService_ServiceDesc.ServiceName, v.MethodName))
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
		fmt.Sprintf("/%s/IsReady", daemonv1.WalletUnlockerService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GenSeed", daemonv1.WalletUnlockerService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "read",
		}},
		fmt.Sprintf("/%s/InitWallet", daemonv1.WalletUnlockerService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UnlockWallet", daemonv1.WalletUnlockerService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ChangePassword", daemonv1.WalletUnlockerService_ServiceDesc.ServiceName): {{
			Entity: EntityUnlocker,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListMarkets", tdexv1.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetMarketBalance", tdexv1.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetMarketPrice", tdexv1.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/PreviewTrade", tdexv1.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ProposeTrade", tdexv1.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "write",
		}},
		fmt.Sprintf("/%s/CompleteTrade", tdexv1.TradeService_ServiceDesc.ServiceName): {{
			Entity: EntityTrade,
			Action: "write",
		}},
		fmt.Sprintf("/%v/SupportedContentTypes", tdexv1.TransportService_ServiceDesc.ServiceName): {{
			Entity: EntityTransport,
			Action: "read",
		}},
		// Tdex old proto
		fmt.Sprintf("/%s/Markets", tdexold.File_trade_proto.Services().Get(0).FullName()): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/Balances", tdexold.File_trade_proto.Services().Get(0).FullName()): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/MarketPrice", tdexold.File_trade_proto.Services().Get(0).FullName()): {{
			Entity: EntityTrade,
			Action: "read",
		}},
		fmt.Sprintf("/%s/TradePropose", tdexold.File_trade_proto.Services().Get(0).FullName()): {{
			Entity: EntityTrade,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ProposeTrade", tdexold.File_trade_proto.Services().Get(0).FullName()): {{
			Entity: EntityTrade,
			Action: "write",
		}},
		fmt.Sprintf("/%s/TradeComplete", tdexold.File_trade_proto.Services().Get(0).FullName()): {{
			Entity: EntityTrade,
			Action: "write",
		}},
		fmt.Sprintf("/%s/CompleteTrade", tdexold.File_trade_proto.Services().Get(0).FullName()): {{
			Entity: EntityTrade,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ServerReflectionInfo", grpc_reflection_v1alpha.File_reflection_grpc_reflection_v1alpha_reflection_proto.Services().Get(0).FullName()): {{
			Entity: Reflection,
			Action: "write",
		}},
	}
}

// AllPermissionsByMethod returns a mapping of the RPC server calls to the
// permissions they require.
func AllPermissionsByMethod() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		fmt.Sprintf("/%s/WalletAddress", daemonv1.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityWallet,
			Action: "write",
		}},
		fmt.Sprintf("/%s/SendToMany", daemonv1.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityWallet,
			Action: "write",
		}},
		fmt.Sprintf("/%s/WalletBalance", daemonv1.WalletService_ServiceDesc.ServiceName): {{
			Entity: EntityWallet,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetInfo", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetFeeAddress", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListFeeAddresses", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetFeeBalance", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ClaimFeeDeposits", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/WithdrawFee", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "write",
		}},
		fmt.Sprintf("/%s/NewMarket", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/GetMarketInfo", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetMarketAddress", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListMarketAddresses", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetMarketBalance", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ClaimMarketDeposits", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/OpenMarket", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/CloseMarket", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/DropMarket", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "write",
		}},
		fmt.Sprintf("/%s/GetMarketCollectedSwapFees", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/WithdrawMarket", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketPercentageFee", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketFixedFee", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketPrice", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityPrice,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketAssetsPrecision", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdateMarketStrategy", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/GetFeeFragmenterAddress", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListFeeFragmenterAddresses", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetFeeFragmenterBalance", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/FeeFragmenterSplitFunds", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/WithdrawFeeFragmenter", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/GetMarketFragmenterAddress", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListMarketFragmenterAddresses", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetMarketFragmenterBalance", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/MarketFragmenterSplitFunds", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/WithdrawMarketFragmenter", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListMarkets", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityPrice,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ListTrades", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ListDeposits", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ListWithdrawals", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ReloadUtxos", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListUtxos", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityOperator,
			Action: "read",
		}},
		fmt.Sprintf("/%s/AddWebhook", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityWebhook,
			Action: "write",
		}},
		fmt.Sprintf("/%s/RemoveWebhook", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityWebhook,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListWebhooks", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityWebhook,
			Action: "read",
		}},
		fmt.Sprintf("/%s/GetMarketReport", daemonv1.OperatorService_ServiceDesc.ServiceName): {{
			Entity: EntityMarket,
			Action: "read",
		}},
	}
}
