package config

import (
	"errors"
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
	// TraderListeningPortKey is the port where the gRPC Trader interface will listen on
	TraderListeningPortKey = "TRADER_LISTENING_PORT"
	// OperatorListeningPortKey is the port where the gRPC Operator interface will listen on
	OperatorListeningPortKey = "OPERATOR_LISTENING_PORT"
	// ExplorerEndpointKey is the endpoint where the Electrs (for Liquid) REST API is listening
	ExplorerEndpointKey = "EXPLORER_ENDPOINT"
	// ExplorerRequestTimeoutKey is the max time to wait for HTTP responses before it expires
	ExplorerRequestTimeoutKey = "EXPLORER_REQUEST_TIMEOUT"
	// DataDirPathKey is the local data directory to store the internal state of daemon
	DataDirPathKey = "DATA_DIR_PATH"
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
	// SSLCertPathKey is the path to the SSL certificate
	SSLCertPathKey = "SSL_CERT"
	// SSLKeyPathKey is the path to the SSL private key
	SSLKeyPathKey = "SSL_KEY"
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
)

var vip *viper.Viper
var defaultDataDir = btcutil.AppDataDir("tdex-daemon", false)

func init() {
	vip = viper.New()
	vip.SetEnvPrefix("TDEX")
	vip.AutomaticEnv()

	vip.SetDefault(TraderListeningPortKey, 9945)
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
	vip.SetDefault(DataDirPathKey, defaultDataDir)
	vip.SetDefault(PriceSlippageKey, 0.05)
	vip.SetDefault(EnableProfilerKey, false)
	vip.SetDefault(StatsIntervalKey, 600)
	vip.SetDefault(CrawlLimitKey, 10)
	vip.SetDefault(CrawlTokenBurst, 1)

	validate()

	if err := initDataDir(); err != nil {
		log.WithError(err).Panic("error while init data dir")
	}

	vip.Set(MnemonicKey, "")
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
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

// Validate method of config will panic
func validate() {
	if err := validateDefaultFee(vip.GetFloat64(DefaultFeeKey)); err != nil {
		log.WithError(err).Panic("default fee is not valid")
	}
	if err := validateDefaultNetwork(vip.GetString(NetworkKey)); err != nil {
		log.WithError(err).Panic("default network is not valid")
	}
	path := vip.GetString(DataDirPathKey)
	if path != defaultDataDir {
		if err := validatePath(path); err != nil {
			log.WithError(err).Panic("datadir is not valid")
		}
	}
	certPath, keyPath := vip.GetString(SSLCertPathKey), vip.GetString(SSLKeyPathKey)
	if (certPath != "" && keyPath == "") || (certPath == "" && keyPath != "") {
		log.Fatalln("SSL requires both key and certificate when enabled")
	}

	if rpcEndpoint := vip.GetString(ElementsRPCEndpointKey); rpcEndpoint != "" {
		if err := validateEndpoint(rpcEndpoint); err != nil {
			log.WithError(err).Panic("Elements RPC endpoint is not a valid url")
		}
		// ElementsStartRescanTimestamp can assume the 0 value that means scanning
		// the entire blockchain. This wil be used only in regtest mode
		if vip.IsSet(ElementsStartRescanTimestampKey) {
			rescanTime := vip.GetInt(ElementsStartRescanTimestampKey)
			if rescanTime < 0 {
				log.WithError(
					fmt.Errorf("timestamp must not be a negative number"),
				).Panic("Elements rescan timestamp is not valid")
			}
		}
	} else {
		if err := validateEndpoint(vip.GetString(ExplorerEndpointKey)); err != nil {
			log.WithError(err).Panic("explorer endpoint is not a valid url")
		}
	}
}

func validateDefaultFee(fee float64) error {
	if fee < 0.01 || fee > 99 {
		return errors.New("percentage of the fee on each swap must be > 0.01 and < 99")
	}

	return nil
}

func validateDefaultNetwork(net string) error {
	if net != network.Liquid.Name && net != network.Regtest.Name {
		return fmt.Errorf(
			"network must be either '%s' or '%s'",
			network.Liquid.Name,
			network.Regtest.Name,
		)
	}
	return nil
}

func validatePath(path string) error {
	if path != "" {
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}

		if !stat.IsDir() {
			return errors.New("not a directory")
		}
	}

	return nil
}

func validateEndpoint(endpoint string) error {
	_, err := url.Parse(endpoint)
	return err
}

func initDataDir() error {
	dataDir := GetString(DataDirPathKey)
	if err := makeDirectoryIfNotExists(dataDir); err != nil {
		log.WithError(err).Panic(
			fmt.Sprintf("error while creating %v folder", dataDir),
		)
	}
	if err := makeDirectoryIfNotExists(filepath.Join(dataDir, "db")); err != nil {
		log.WithError(err).Panic("error while creating db folder")
	}
	if err := makeDirectoryIfNotExists(filepath.Join(dataDir, "stats")); err != nil {
		log.WithError(err).Panic("error while creating stats folder")
	}

	return nil
}
