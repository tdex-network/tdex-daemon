package config

import (
	"errors"
	"time"

	"github.com/btcsuite/btcutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// TraderListeningPortKey ...
	TraderListeningPortKey = "TRADER_LISTENING_PORT"
	// OperatorListeningPortKey ...
	OperatorListeningPortKey = "OPERATOR_LISTENING_PORT"
	// ExplorerEndpointKey ...
	ExplorerEndpointKey = "EXPLORER_ENDPOINT"
	// DataDirPathKey ...
	DataDirPathKey = "DATA_DIR_PATH"
	// LogLevelKey ...
	LogLevelKey = "LOG_LEVEL"
	// DefaultFeeKey ...
	DefaultFeeKey = "DEFAULT_FEE"
	// BaseAssetKey ...
	BaseAssetKey = "BASE_ASSET"
)

var vip *viper.Viper

func init() {
	vip = viper.New()
	vip.SetEnvPrefix("TDEX")
	vip.AutomaticEnv()

	vip.SetDefault(TraderListeningPortKey, 9945)
	vip.SetDefault(OperatorListeningPortKey, 9000)
	vip.SetDefault(ExplorerEndpointKey, "http://127.0.0.1:3001")
	vip.SetDefault(DataDirPathKey, btcutil.AppDataDir("tdex-daemon", false))
	vip.SetDefault(LogLevelKey, 5)
	vip.SetDefault(DefaultFeeKey, 0.25)
	vip.SetDefault(BaseAssetKey, "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225")

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

// Validate method of config will panic
func Validate() {
	if err := validateDefaultFee(vip.GetFloat64(DefaultFeeKey)); err != nil {
		log.Fatalln(err)
	}
}

func validateDefaultFee(fee float64) error {
	if fee < 0.01 || fee > 99 {
		return errors.New("percentage of the fee on each swap must be > 0.01 and < 99")
	}

	return nil
}
