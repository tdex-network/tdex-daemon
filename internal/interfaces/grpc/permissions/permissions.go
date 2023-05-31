package permissions

import (
	"fmt"

	"gopkg.in/macaroon-bakery.v2/bakery"

	reflectionv1 "github.com/tdex-network/reflection/api-spec/protobuf/gen/reflection/v1"
	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	EntityOperator   = "operator"
	EntityTrade      = "trade"
	EntityMarket     = "market"
	EntityPrice      = "price"
	EntityUnlocker   = "unlocker"
	EntityWallet     = "wallet"
	EntityWebhook    = "webhook"
	EntityTransport  = "transport"
	EntityReflection = "reflection"
	EntityHealth     = "health"
	EntityFeeder     = "feeder"
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
		{
			Entity: EntityFeeder,
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
		{
			Entity: EntityFeeder,
			Action: "read",
		},
		{
			Entity: EntityFeeder,
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
		fmt.Sprintf("/%s/GetInfo", reflectionv1.ReflectionService_ServiceDesc.ServiceName): {{
			Entity: EntityReflection,
			Action: "read",
		}},
		fmt.Sprintf("/%s/Check", grpchealth.Health_ServiceDesc.ServiceName): {{
			Entity: EntityHealth,
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
		fmt.Sprintf("/%s/AddWebhook", daemonv2.WebhookService_ServiceDesc.ServiceName): {{
			Entity: EntityWebhook,
			Action: "write",
		}},
		fmt.Sprintf("/%s/RemoveWebhook", daemonv2.WebhookService_ServiceDesc.ServiceName): {{
			Entity: EntityWebhook,
			Action: "write",
		}},
		fmt.Sprintf("/%s/ListWebhooks", daemonv2.WebhookService_ServiceDesc.ServiceName): {{
			Entity: EntityWebhook,
			Action: "read",
		}},
		fmt.Sprintf("/%s/AddPriceFeed", daemonv2.FeederService_ServiceDesc.ServiceName): {{
			Entity: EntityFeeder,
			Action: "write",
		}},
		fmt.Sprintf("/%s/StartPriceFeed", daemonv2.FeederService_ServiceDesc.ServiceName): {{
			Entity: EntityFeeder,
			Action: "write",
		}},
		fmt.Sprintf("/%s/StopPriceFeed", daemonv2.FeederService_ServiceDesc.ServiceName): {{
			Entity: EntityFeeder,
			Action: "write",
		}},
		fmt.Sprintf("/%s/UpdatePriceFeed", daemonv2.FeederService_ServiceDesc.ServiceName): {{
			Entity: EntityFeeder,
			Action: "write",
		}},
		fmt.Sprintf("/%s/RemovePriceFeed", daemonv2.FeederService_ServiceDesc.ServiceName): {{
			Entity: EntityFeeder,
			Action: "write",
		}},
		fmt.Sprintf("/%s/GetPriceFeed", daemonv2.FeederService_ServiceDesc.ServiceName): {{
			Entity: EntityFeeder,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ListPriceFeeds", daemonv2.FeederService_ServiceDesc.ServiceName): {{
			Entity: EntityFeeder,
			Action: "read",
		}},
		fmt.Sprintf("/%s/ListSupportedPriceSources", daemonv2.FeederService_ServiceDesc.ServiceName): {{
			Entity: EntityFeeder,
			Action: "read",
		}},
	}
}
