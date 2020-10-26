package interceptor

import (
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

// UnaryInterceptor returns the unary interceptor
func UnaryInterceptor(
	dbManager *dbbadger.DbManager,
	macaroonService *macaroons.Service,
) grpc.ServerOption {
	return grpc.UnaryInterceptor(
		middleware.ChainUnaryServer(
			macaroonService.UnaryServerInterceptor(MainRPCServerPermissions()),
			unaryLogger,
			unaryTransactionHandler(dbManager),
		),
	)
}

// StreamInterceptor returns the stream interceptor with a logrus log
func StreamInterceptor(
	dbManager *dbbadger.DbManager,
	macaroonService *macaroons.Service,
) grpc.ServerOption {
	return grpc.StreamInterceptor(
		middleware.ChainStreamServer(
			macaroonService.StreamServerInterceptor(MainRPCServerPermissions()),
			streamLogger,
			streamTransactionHandler(dbManager),
		),
	)
}

// MainRPCServerPermissions returns a mapping of the main RPC server calls to
// the permissions they require.
func MainRPCServerPermissions() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		"/Wallet/GenSeed": {{
			Entity: "wallet",
			Action: "genseed",
		}},
		"/Wallet/InitWallet": {{
			Entity: "wallet",
			Action: "initwallet",
		}},
		"/Wallet/UnlockWallet": {{
			Entity: "wallet",
			Action: "unlockwallet",
		}},
		"/Wallet/ChangePassword": {{
			Entity: "wallet",
			Action: "changepassword",
		}},
		"/Wallet/WalletAddress": {{
			Entity: "wallet",
			Action: "walletaddress",
		}},
		"/Wallet/WalletBalance": {{
			Entity: "wallet",
			Action: "walletbalance",
		}},
		"/Wallet/SendToMany": {{
			Entity: "wallet",
			Action: "sendtomany",
		}},
		"/Operator/DepositMarket": {{
			Entity: "operator",
			Action: "depositmarket",
		}},
		"/Operator/ListDepositMarket": {{
			Entity: "operator",
			Action: "listdepositmarket",
		}},
		"/Operator/DepositFeeAccount": {{
			Entity: "operator",
			Action: "depositfeeaccount",
		}},
		"/Operator/BalanceFeeAccount": {{
			Entity: "operator",
			Action: "balancefeeaccount",
		}},
		"/Operator/OpenMarket": {{
			Entity: "operator",
			Action: "openmarket",
		}},
		"/Operator/CloseMarket": {{
			Entity: "operator",
			Action: "closemarket",
		}},
		"/Operator/ListMarket": {{
			Entity: "operator",
			Action: "listmarket",
		}},
		"/Operator/UpdateMarketFee": {{
			Entity: "operator",
			Action: "updatemarketfee",
		}},
		"/Operator/UpdateMarketPrice": {{
			Entity: "operator",
			Action: "updatemarketprice",
		}},
		"/Operator/UpdateMarketStrategy": {{
			Entity: "operator",
			Action: "updatemarketstrategy",
		}},
		"/Operator/WithdrawMarket": {{
			Entity: "operator",
			Action: "withdrawmarket",
		}},
		"/Operator/ListSwaps": {{
			Entity: "operator",
			Action: "listswaps",
		}},
		"/Operator/ReportMarketFee": {{
			Entity: "operator",
			Action: "reportmarketfee",
		}},
		"/Trade/Markets": {{
			Entity: "trade",
			Action: "markets",
		}},
		"/Trade/Balances": {{
			Entity: "trade",
			Action: "balances",
		}},
		"/Trade/MarketPrice": {{
			Entity: "trade",
			Action: "marketprice",
		}},
		"/Trade/TradePropose": {{
			Entity: "trade",
			Action: "tradepropose",
		}},
		"/Trade/TradeComplete": {{
			Entity: "trade",
			Action: "tradecomplete",
		}},
	}
}
