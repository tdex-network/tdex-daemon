package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/tdex-network/tdex-daemon/internal/core/application"

	"github.com/spf13/viper"
)

const (
	// TradeListeningPortKey is the port where the gRPC Trader interface will listen on
	TradeListeningPortKey = "TRADE_LISTENING_PORT"
	// OperatorListeningPortKey is the port where the gRPC Operator interface will listen on
	OperatorListeningPortKey = "OPERATOR_LISTENING_PORT"
	// DatadirKey is the local data directory to store the internal state of daemon
	DatadirKey = "DATADIR"
	// LogLevelKey are the different logging levels. For reference on the values https://godoc.org/github.com/sirupsen/logrus#Level
	LogLevelKey = "LOG_LEVEL"
	// FeeAccountBalanceThresholdKey is the treshold of LBTC balance (in satoshis) for the fee account, after which we alert the operator that it cannot subsidize anymore swaps
	FeeAccountBalanceThresholdKey = "FEE_ACCOUNT_BALANCE_THRESHOLD"
	// TradeExpiryTimeKey is the duration in seconds of lock on unspents we reserve for accepted trades, before eventually double spending it
	TradeExpiryTimeKey = "TRADE_EXPIRY_TIME"
	// TradeSatsPerByte is the sats per byte ratio to use for paying for trades' network fees
	TradeSatsPerByte = "TRADE_SATS_PER_BYTE"
	// PriceSlippageKey is the percentage of the slipage for accepting trades compared to current spot price
	PriceSlippageKey = "PRICE_SLIPPAGE"
	// TradeTLSKeyKey is the path of the the TLS key for the Trade interface
	TradeTLSKeyKey = "TRADE_TLS_KEY"
	// TradeTLSCertKey is the path of the the TLS certificate for the Trade interface
	TradeTLSCertKey = "TRADE_TLS_CERT"
	// EnableProfilerKey enables profiler that can be used to investigate performance issues
	EnableProfilerKey = "ENABLE_PROFILER"
	// StatsIntervalKey defines interval for printing basic tdex statistics
	StatsIntervalKey = "STATS_INTERVAL"
	// NoMacaroonsKey is used to start the daemon without using macaroons auth
	// service.
	NoMacaroonsKey = "NO_MACAROONS"
	// NoOperatorTlsKey is used to start the daemon without using TLS for the operator service
	NoOperatorTlsKey = "NO_OPERATOR_TLS"
	// OperatorExtraIP is used to add an extra ip address to the self-signed TLS
	// certificate for the Operator gRPC interface.
	OperatorExtraIPKey = "OPERATOR_EXTRA_IP"
	// OperatorExtraDomain is used to add an extra domain to the self signed TLS
	// certificate for the Operator gRPC interface. This is useful to add the
	// onion endpoint in case the daemon is served as a TOR hidden service for
	// example.
	OperatorExtraDomainKey = "OPERATOR_EXTRA_DOMAIN"
	// WalletUnlockPasswordFile defines full path to a file  that contains the
	//password for unlocking the wallet, if provided wallet will be unlocked
	//automatically
	WalletUnlockPasswordFile = "WALLET_UNLOCK_PASSWORD_FILE"
	// ConnectAddrKey is the address <host:port> of the tdex-daemon service to connect to, ie. myservice:9000
	ConnectAddrKey = "CONNECT_ADDR"
	// ConnectUrlProtoKey is the http/https protocol of the tdex-daemon service to connect to
	// used by tdexconnect to create connection url
	ConnectProtoKey = "CONNECT_PROTO"
	// OceanWalletAddrKey is the address for connecting to the ocean wallet.
	OceanWalletAddrKey = "WALLET_ADDR"
	// DBTypeKey is used to switch database type between those supported
	DBTypeKey = "DB_TYPE"
	// PgConnectAddr is postgres connection string in postgresql://user:password@host:port/name format
	PgConnectAddr = "PG_CONNECT_ADDR"
	// PgMigrationSource is the path to the migration files for postgres
	PgMigrationSource = "PG_MIGRATION_SOURCE"

	DbLocation        = "db"
	TLSLocation       = "tls"
	MacaroonsLocation = "macaroons"
	ProfilerLocation  = "stats"

	httpsProtocol = "https"
)

var vip *viper.Viper
var defaultDatadir = btcutil.AppDataDir("tdex-daemon", false)

func InitConfig() error {
	vip = viper.New()
	vip.SetEnvPrefix("TDEX")
	vip.AutomaticEnv()

	vip.SetDefault(TradeListeningPortKey, 9945)
	vip.SetDefault(OperatorListeningPortKey, 9000)
	vip.SetDefault(LogLevelKey, 4)
	vip.SetDefault(FeeAccountBalanceThresholdKey, 5000)
	vip.SetDefault(TradeExpiryTimeKey, 120)
	vip.SetDefault(TradeSatsPerByte, 0.1)
	vip.SetDefault(DatadirKey, defaultDatadir)
	vip.SetDefault(PriceSlippageKey, 0.05)
	vip.SetDefault(EnableProfilerKey, false)
	vip.SetDefault(StatsIntervalKey, 600)
	vip.SetDefault(NoMacaroonsKey, false)
	vip.SetDefault(NoOperatorTlsKey, false)
	vip.SetDefault(ConnectProtoKey, httpsProtocol)
	vip.SetDefault(DBTypeKey, application.DBPostgres)
	vip.SetDefault(PgMigrationSource, "file://internal/infrastructure/storage/db/pg/migration/")

	if err := validate(); err != nil {
		return fmt.Errorf("error while validating config: %s", err)
	}

	if err := initDatadir(); err != nil {
		return fmt.Errorf("error while creating datadir: %s", err)
	}

	return nil
}

func GetString(key string) string {
	return vip.GetString(key)
}

func GetInt(key string) int {
	return vip.GetInt(key)
}

func GetFloat(key string) float64 {
	return vip.GetFloat64(key)
}

func GetStringSlice(key string) []string {
	return vip.GetStringSlice(key)
}

func GetDuration(key string) time.Duration {
	return vip.GetDuration(key)
}

func GetBool(key string) bool {
	return vip.GetBool(key)
}

func GetDatadir() string {
	return GetString(DatadirKey)
}

func validate() error {
	datadir := GetString(DatadirKey)
	if len(datadir) <= 0 {
		return fmt.Errorf("missing datadir")
	}

	tlsKey, tlsCert := GetString(TradeTLSKeyKey), GetString(TradeTLSCertKey)
	if (tlsKey == "" && tlsCert != "") || (tlsKey != "" && tlsCert == "") {
		return fmt.Errorf(
			"TLS for Trade interface requires both key and certificate when enabled",
		)
	}

	satsPerByte := GetFloat(TradeSatsPerByte)
	if satsPerByte < 0.1 {
		return fmt.Errorf("%s must be equal or greater than 0.1", TradeSatsPerByte)
	}

	if !vip.IsSet(OceanWalletAddrKey) {
		return fmt.Errorf("missing wallet address")
	}

	if !validatePgConnectionString(vip.GetString(PgConnectAddr)) {
		return fmt.Errorf("please provide a valid postgres connection string" +
			" in the format: postgres://user:password@host:port/dbname")
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
	}
	noOperatorTls := GetBool(NoOperatorTlsKey)
	if !noOperatorTls {
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

func validatePgConnectionString(connectionString string) bool {
	pattern := `^postgresql:\/\/([^:]+):([^@]+)@([^:]+):(\d+)\/(.+)$`
	matched, _ := regexp.MatchString(pattern, connectionString)

	return matched
}
