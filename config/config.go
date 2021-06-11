package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/elements"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vulpemventures/go-elements/network"
)

const (
	// TradeListeningPortKey is the port where the gRPC Trader interface will listen on
	TradeListeningPortKey = "TRADE_LISTENING_PORT"
	// OperatorListeningPortKey is the port where the gRPC Operator interface will listen on
	OperatorListeningPortKey = "OPERATOR_LISTENING_PORT"
	// ExplorerEndpointKey is the endpoint where the Electrs (for Liquid) REST API is listening
	ExplorerEndpointKey = "EXPLORER_ENDPOINT"
	// ExplorerRequestTimeoutKey are the milliseconds to wait for HTTP responses before timeouts
	ExplorerRequestTimeoutKey = "EXPLORER_REQUEST_TIMEOUT"
	// DatadirKey is the local data directory to store the internal state of daemon
	DatadirKey = "DATA_DIR_PATH"
	// LogLevelKey are the different logging levels. For reference on the values https://godoc.org/github.com/sirupsen/logrus#Level
	LogLevelKey = "LOG_LEVEL"
	// DefaultFeeKey is the default swap fee when creating a market
	DefaultFeeKey = "DEFAULT_FEE"
	// NetworkKey is the network to use. Either "liquid" or "regtest"
	NetworkKey = "NETWORK"
	// BaseAssetKey is the default asset hash to be used as base asset for all markets. Default is LBTC
	BaseAssetKey = "BASE_ASSET"
	// CrawlIntervalKey is the interval in milliseconds to be used when watching the blockchain via the explorer
	CrawlIntervalKey = "CRAWL_INTERVAL"
	// FeeAccountBalanceThresholdKey is the treshold of LBTC balance (in satoshis) for the fee account, after wich we alert the operator that it cannot subsidize anymore swaps
	FeeAccountBalanceThresholdKey = "FEE_ACCOUNT_BALANCE_THRESHOLD"
	// TradeExpiryTimeKey is the duration in seconds of lock on unspents we reserve for accpeted trades, before eventually double spending it
	TradeExpiryTimeKey = "TRADE_EXPIRY_TIME"
	// PriceSlippageKey is the percentage of the slipage for accepting trades compared to current spot price
	PriceSlippageKey = "PRICE_SLIPPAGE"
	// TradeTLSKeyKey is the path of the the TLS key for the Trade interface
	TradeTLSKeyKey = "SSL_KEY"
	// TradeTLSCertKey is the path of the the TLS certificate for the Trade interface
	TradeTLSCertKey = "SSL_CERT"
	// MnemonicKey is the mnemonic of the master private key of the daemon's wallet
	MnemonicKey = "MNEMONIC"
	// EnableProfilerKey nables profiler that can be used to investigate performance issues
	EnableProfilerKey = "ENABLE_PROFILER"
	// StatsIntervalKey defines interval for printing basic tdex statistics
	StatsIntervalKey = "STATS_INTERVAL"
	// ElementsRPCEndpointKey is the url for the RPC interface of the Elements
	// node in the form protocol://user:password@host:port
	ElementsRPCEndpointKey = "ELEMENTS_RPC_ENDPOINT"
	// ElementsStartRescanTimestampKey is the date in Unix seconds of the block
	// from where the node should start rescanning addresses
	ElementsStartRescanTimestampKey = "ELEMENTS_START_RESCAN_TIMESTAMP"
	// CrawlLimitKey represents number of requests per second that crawler
	//makes to explorer
	CrawlLimitKey = "CRAWL_LIMIT"
	// CrawlTokenBurst represents number of bursts tokens permitted from
	//crawler to explorer
	CrawlTokenBurst = "CRAWL_TOKEN"
	// NoMacaroonsKey is used to start the daemon without using macaroons auth
	// service.
	NoMacaroonsKey = "NO_MACAROONS"

	DbLocation        = "db"
	TLSLocation       = "tls"
	MacaroonsLocation = "macaroons"
	ProfilerLocation  = "stats"

	MinDefaultPercentageFee = 0.01
	MaxDefaultPercentageFee = float64(99)
)

var vip *viper.Viper
var defaultDatadir = btcutil.AppDataDir("tdex-daemon", false)

func init() {
	vip = viper.New()
	vip.SetEnvPrefix("TDEX")
	vip.AutomaticEnv()

	vip.SetDefault(TradeListeningPortKey, 9945)
	vip.SetDefault(OperatorListeningPortKey, 9000)
	vip.SetDefault(ExplorerEndpointKey, "https://blockstream.info/liquid/api")
	vip.SetDefault(ExplorerRequestTimeoutKey, 15000)
	vip.SetDefault(LogLevelKey, 4)
	vip.SetDefault(DefaultFeeKey, 0.25)
	vip.SetDefault(CrawlIntervalKey, 5000)
	vip.SetDefault(FeeAccountBalanceThresholdKey, 5000)
	vip.SetDefault(NetworkKey, network.Liquid.Name)
	vip.SetDefault(BaseAssetKey, network.Liquid.AssetID)
	vip.SetDefault(TradeExpiryTimeKey, 120)
	vip.SetDefault(DatadirKey, defaultDatadir)
	vip.SetDefault(PriceSlippageKey, 0.05)
	vip.SetDefault(EnableProfilerKey, false)
	vip.SetDefault(StatsIntervalKey, 600)
	vip.SetDefault(CrawlLimitKey, 10)
	vip.SetDefault(CrawlTokenBurst, 1)
	vip.SetDefault(NoMacaroonsKey, false)

	if err := validate(); err != nil {
		log.WithError(err).Panic("error while validating config")
	}

	if err := initDatadir(); err != nil {
		log.WithError(err).Panic("error while creating datadir")
	}

	vip.Set(MnemonicKey, "")
}

//GetString ...
func GetString(key string) string {
	return vip.GetString(key)
}

//GetInt ...
func GetInt(key string) int {
	return vip.GetInt(key)
}

//GetFloat ...
func GetFloat(key string) float64 {
	return vip.GetFloat64(key)
}

//GetDuration ...
func GetDuration(key string) time.Duration {
	return vip.GetDuration(key)
}

//GetBool ...
func GetBool(key string) bool {
	return vip.GetBool(key)
}

//GetNetwork ...
func GetNetwork() *network.Network {
	if vip.GetString(NetworkKey) == network.Regtest.Name {
		return &network.Regtest
	}
	return &network.Liquid
}

// TODO: attach network name to datadir
func GetDatadir() string {
	return GetString(DatadirKey)
}

//GetExplorer ...
func GetExplorer() (explorer.Service, error) {
	if rpcEndpoint := GetString(ElementsRPCEndpointKey); rpcEndpoint != "" {
		var rescanTime interface{}
		if vip.IsSet(ElementsStartRescanTimestampKey) {
			rescanTime = vip.GetInt(ElementsStartRescanTimestampKey)
		}
		return elements.NewService(rpcEndpoint, rescanTime)
	}

	endpoint := GetString(ExplorerEndpointKey)
	reqTimeout := GetInt(ExplorerRequestTimeoutKey)
	return esplora.NewService(endpoint, reqTimeout)
}

// Set a value for the given key
func Set(key string, value interface{}) {
	vip.Set(key, value)
}

// IsSet returns whether the give key is set
func IsSet(key string) bool {
	return vip.IsSet(key)
}

// GetMnemonic returns the current set mnemonic
func GetMnemonic() []string {
	var mnemonic []string
	if vip.GetString(MnemonicKey) != "" {
		mnemonic = strings.Split(vip.GetString(MnemonicKey), " ")
	}

	return mnemonic
}

func validate() error {
	datadir := GetString(DatadirKey)
	if len(datadir) <= 0 {
		return fmt.Errorf("datadir must not be null")
	}

	percentageFee := GetFloat(DefaultFeeKey)
	if percentageFee < MinDefaultPercentageFee ||
		percentageFee > MaxDefaultPercentageFee {
		return fmt.Errorf(
			"percentage of the fee on each swap must be in range [%.2f, %.2f]",
			MinDefaultPercentageFee, MaxDefaultPercentageFee,
		)
	}

	networkName := GetString(NetworkKey)
	if networkName != network.Liquid.Name &&
		networkName != network.Regtest.Name {
		return fmt.Errorf(
			"network must be either '%s' or '%s'",
			network.Liquid.Name,
			network.Regtest.Name,
		)
	}

	tlsKey, tlsCert := GetString(TradeTLSKeyKey), GetString(TradeTLSCertKey)
	if (tlsKey == "" && tlsCert != "") || (tlsKey != "" && tlsCert == "") {
		return fmt.Errorf(
			"TLS over Trade interface requires both key and certificate when enabled",
		)
	}

	elementsRpcEndpoint := GetString(ElementsRPCEndpointKey)
	if elementsRpcEndpoint != "" {
		if _, err := url.Parse(elementsRpcEndpoint); err != nil {
			return fmt.Errorf("Elements RPC endpoint is not a valid url: %s", err)
		}
		// ElementsStartRescanTimestamp can assume the 0 value that means scanning
		// the entire blockchain. This wil be used only in regtest mode
		if vip.IsSet(ElementsStartRescanTimestampKey) {
			rescanTime := vip.GetInt(ElementsStartRescanTimestampKey)
			if rescanTime < 0 {
				return fmt.Errorf("timestamp must not be a negative number")
			}
		}
	} else {
		exploreEndpoint := GetString(ExplorerEndpointKey)
		if _, err := url.Parse(exploreEndpoint); err != nil {
			return fmt.Errorf("explorer endpoint is not a valid url: %s", err)
		}
	}
	return nil
}

func initDatadir() error {
	datadir := GetDatadir()
	if err := makeDirectoryIfNotExists(filepath.Join(datadir, DbLocation)); err != nil {
		return err
	}

	profilerEnabled := GetBool(EnableProfilerKey)
	if profilerEnabled {
		if err := makeDirectoryIfNotExists(filepath.Join(datadir, ProfilerLocation)); err != nil {
			return err
		}
	}

	// if macaroons is enabled, the daemon automatically enables TLS encryption
	// on the operator interface
	noMacaroons := GetBool(NoMacaroonsKey)
	if !noMacaroons {
		if err := makeDirectoryIfNotExists(filepath.Join(datadir, MacaroonsLocation)); err != nil {
			return err
		}
		if err := makeDirectoryIfNotExists(filepath.Join(datadir, TLSLocation)); err != nil {
			return err
		}
	}
	return nil
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModeDir|0755)
	}
	return nil
}
