package config

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
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
	// QuoteAssetKey is the default asset hash to be used as quote asset for all markets. This is mutually exclusive with BaseAssetKey.
	QuoteAssetKey = "QUOTE_ASSET"
	//NativeAssetKey is used to set lbtc hash, used for fee outputs, in regtest network
	NativeAssetKey = "NATIVE_ASSET"
	// CrawlIntervalKey is the interval in milliseconds to be used when watching the blockchain via the explorer
	CrawlIntervalKey = "CRAWL_INTERVAL"
	// FeeAccountBalanceThresholdKey is the treshold of LBTC balance (in satoshis) for the fee account, after wich we alert the operator that it cannot subsidize anymore swaps
	FeeAccountBalanceThresholdKey = "FEE_ACCOUNT_BALANCE_THRESHOLD"
	// TradeExpiryTimeKey is the duration in seconds of lock on unspents we reserve for accpeted trades, before eventually double spending it
	TradeExpiryTimeKey = "TRADE_EXPIRY_TIME"
	// TradeSatsPerByte is the sats per byte ratio to use for paying for trades' network fees
	TradeSatsPerByte = "TRADE_SATS_PER_BYTE"
	// PriceSlippageKey is the percentage of the slipage for accepting trades compared to current spot price
	PriceSlippageKey = "PRICE_SLIPPAGE"
	// TradeTLSKeyKey is the path of the the TLS key for the Trade interface
	TradeTLSKeyKey = "SSL_KEY"
	// TradeTLSCertKey is the path of the the TLS certificate for the Trade interface
	TradeTLSCertKey = "SSL_CERT"
	// MnemonicKey is the mnemonic of the master private key of the daemon's wallet
	MnemonicKey = "MNEMONIC"
	// EnableProfilerKey enables profiler that can be used to investigate performance issues
	EnableProfilerKey = "ENABLE_PROFILER"
	// StatsIntervalKey defines interval for printing basic tdex statistics
	StatsIntervalKey = "STATS_INTERVAL"
	// NoMacaroonsKey is used to start the daemon without using macaroons auth
	// service.
	NoMacaroonsKey = "NO_MACAROONS"
	// OperatorExtraIP is used to add an extra ip address to the self-signed TLS
	// certificate for the Operator gRPC interface.
	OperatorExtraIPKey = "OPERATOR_EXTRA_IP"
	// OperatorExtraDomain is used to add an extra domain to the self signed TLS
	// certificate for the Operator gRPC interface. This is useful to add the
	// onion endpoint in case the daemon is served as a TOR hidden service for
	// example.
	OperatorExtraDomainKey = "OPERATOR_EXTRA_DOMAIN"
	// CBMaxFailingRequestsKey is used in combo with FailingRatio to set the max
	// number of failing request for the circuit breaker service to change its
	// internal state and stop making network calls.
	CBMaxFailingRequestsKey = "MAX_FAILING_REQUESTS"
	// CBFailingRatioKey is used in combo with MaxFailingRequests to set the
	// failing ratio over which the circuit breaker service to change its
	// internal state and stop making network calls.
	CBFailingRatioKey = "FAILING_RATIO"
	// RescanRangeStartKey defines the initial index from where the daemon should
	// start deriving and scanning for addresses of an account during the
	// restoration of the utxos.
	RescanRangeStartKey = "RESCAN_START"
	// RescanGapLimitKey defines the max number of consecutive unused addresses
	// that cause the restoration to stop.
	// For example, if set to 20, the utxo set restoration terminates whenever
	// 20 consecutive unused addresses, or those not involved in any transaction
	// in the blockchain.
	RescanGapLimitKey = "RESCAN_GAP_LIMIT"
	// WalletUnlockPasswordFile defines full path to a file  that contains the
	//password for unlocking the wallet, if provided wallet will be unlocked
	//automatically
	WalletUnlockPasswordFile = "WALLET_UNLOCK_PASSWORD_FILE"
	// RunOnOnePortKey is used to run the daemon on one port as requested by some users
	RunOnOnePortKey = "RUN_ON_ONE_PORT"

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
	vip.SetDefault(TradeSatsPerByte, 0.1)
	vip.SetDefault(DatadirKey, defaultDatadir)
	vip.SetDefault(PriceSlippageKey, 0.05)
	vip.SetDefault(EnableProfilerKey, false)
	vip.SetDefault(StatsIntervalKey, 600)
	vip.SetDefault(NoMacaroonsKey, false)
	vip.SetDefault(CBMaxFailingRequestsKey, 20)
	vip.SetDefault(CBFailingRatioKey, 0.7)
	vip.SetDefault(RescanRangeStartKey, 0)
	vip.SetDefault(RescanGapLimitKey, 50)
	vip.SetDefault(RunOnOnePortKey, false)

	if err := validate(); err != nil {
		log.Fatalf("error while validating config: %s", err)
	}

	if err := initDatadir(); err != nil {
		log.Fatalf("error while creating datadir: %s", err)
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

func GetStringSlice(key string) []string {
	return vip.GetStringSlice(key)
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
func GetNetwork() (*network.Network, error) {
	switch vip.GetString(NetworkKey) {
	case network.Liquid.Name:
		return &network.Liquid, nil
	case network.Testnet.Name:
		return &network.Testnet, nil
	case network.Regtest.Name:
		net := network.Regtest
		if nativeAsset := vip.GetString(NativeAssetKey); nativeAsset != "" {
			net.AssetID = nativeAsset
		}
		return &net, nil
	default:
		return nil, fmt.Errorf("network is unknown")
	}
}

// TODO: attach network name to datadir
func GetDatadir() string {
	return GetString(DatadirKey)
}

// GetExplorer returns the explorer service to be used by the daemon.
func GetExplorer() (explorer.Service, error) {
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
		networkName != network.Testnet.Name &&
		networkName != network.Regtest.Name {
		return fmt.Errorf(
			"network must be either '%s' | '%s' | '%s'",
			network.Liquid.Name,
			network.Testnet.Name,
			network.Regtest.Name,
		)
	}

	// Check native asset only for regtest
	if networkName == network.Regtest.Name {
		if nativeAsset := GetString(NativeAssetKey); nativeAsset != "" {
			if err := validateAssetString(nativeAsset); err != nil {
				return err
			}
		}
	}

	tlsKey, tlsCert := GetString(TradeTLSKeyKey), GetString(TradeTLSCertKey)
	if (tlsKey == "" && tlsCert != "") || (tlsKey != "" && tlsCert == "") {
		return fmt.Errorf(
			"TLS over Trade interface requires both key and certificate when enabled",
		)
	}

	explorerEndpoint := GetString(ExplorerEndpointKey)
	if _, err := url.Parse(explorerEndpoint); err != nil {
		return fmt.Errorf("explorer endpoint is not a valid url: %s", err)
	}

	maxFailingReq := GetString(CBMaxFailingRequestsKey)
	if _, err := strconv.Atoi(maxFailingReq); err != nil {
		return fmt.Errorf("%s must be a valid number", CBMaxFailingRequestsKey)
	}
	failingRatio := GetString(CBFailingRatioKey)
	if _, err := strconv.ParseFloat(failingRatio, 64); err != nil {
		return fmt.Errorf("%s must be a value in range (0, 1)", CBFailingRatioKey)
	}

	start := GetInt(RescanRangeStartKey)
	end := GetInt(RescanGapLimitKey)
	if start < 0 {
		return fmt.Errorf("%s must not be a negative number", RescanRangeStartKey)
	}
	if end < 0 {
		return fmt.Errorf("%s must not be a negative number", RescanGapLimitKey)
	}
	if start >= end {
		return fmt.Errorf("%s must be greater than %s", RescanRangeStartKey, RescanGapLimitKey)
	}

	satsPerByte := GetFloat(TradeSatsPerByte)
	if satsPerByte < 0.1 {
		return fmt.Errorf("%s must be equal or greater than 0.1", TradeSatsPerByte)
	}

	// If quote asset is set, automatically unset the base asset
	if vip.IsSet(QuoteAssetKey) {
		vip.Set(BaseAssetKey, "")
	}

	if baseAsset := vip.GetString(BaseAssetKey); len(baseAsset) > 0 {
		if err := validateAssetString(baseAsset); err != nil {
			return err
		}
	}
	if quoteAsset := vip.GetString(QuoteAssetKey); len(quoteAsset) > 0 {
		if err := validateAssetString(quoteAsset); err != nil {
			return err
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

func validateAssetString(asset string) error {
	b, err := hex.DecodeString(asset)
	if err != nil {
		return fmt.Errorf("asset %s is not an hex string", asset)
	}
	if len(b) != 32 {
		return fmt.Errorf("asset has invalid length")
	}
	return nil
}
