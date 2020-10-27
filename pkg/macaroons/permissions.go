package macaroons

import "gopkg.in/macaroon-bakery.v2/bakery"

var (
	pricePermissions = []bakery.Op{
		{
			Entity: "market",
			Action: "updatemarketprice",
		},
	}
	marketPermissions = []bakery.Op{
		{
			Entity: "market",
			Action: "openmarket",
		},
		{
			Entity: "market",
			Action: "closemarket",
		},
		{
			Entity: "market",
			Action: "updatemarketstrategy",
		},
	}
	//TODO check permissions
	readonlyPermissions = []bakery.Op{}

	adminPermissions = []bakery.Op{
		{
			Entity: "wallet",
			Action: "genseed",
		},
		{
			Entity: "wallet",
			Action: "initwallet",
		},
		{
			Entity: "wallet",
			Action: "unlockwallet",
		},
		{
			Entity: "wallet",
			Action: "changepassword",
		},
		{
			Entity: "wallet",
			Action: "walletaddress",
		},
		{
			Entity: "wallet",
			Action: "walletbalance",
		},
		{
			Entity: "wallet",
			Action: "sendtomany",
		},
		{
			Entity: "operator",
			Action: "depositmarket",
		},
		{
			Entity: "operator",
			Action: "listdepositmarket",
		},
		{
			Entity: "operator",
			Action: "depositfeeaccount",
		},
		{
			Entity: "operator",
			Action: "balancefeeaccount",
		},
		{
			Entity: "operator",
			Action: "openmarket",
		},
		{
			Entity: "operator",
			Action: "closemarket",
		},
		{
			Entity: "operator",
			Action: "listmarket",
		},
		{
			Entity: "operator",
			Action: "updatemarketfee",
		},
		{
			Entity: "operator",
			Action: "updatemarketprice",
		},
		{
			Entity: "operator",
			Action: "updatemarketstrategy",
		},
		{
			Entity: "operator",
			Action: "withdrawmarket",
		},
		{
			Entity: "operator",
			Action: "listswaps",
		},
		{
			Entity: "operator",
			Action: "reportmarketfee",
		},
		{
			Entity: "trade",
			Action: "markets",
		},
		{
			Entity: "trade",
			Action: "balances",
		},
		{
			Entity: "trade",
			Action: "marketprice",
		},
		{
			Entity: "trade",
			Action: "tradepropose",
		},
		{
			Entity: "trade",
			Action: "tradecomplete",
		},
	}
)

// RPCServerPermissions returns a mapping of the main RPC server calls to
// the permissions they require.
func RPCServerPermissions() map[string][]bakery.Op {
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
