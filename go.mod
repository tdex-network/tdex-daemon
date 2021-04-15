module github.com/tdex-network/tdex-daemon

go 1.14

replace github.com/tdex-network/tdex-daemon/pkg/explorer => ./pkg/explorer

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/gogo/protobuf v1.2.1
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/improbable-eng/grpc-web v0.13.0
	github.com/prometheus/client_golang v0.9.3
	github.com/rs/cors v1.7.0 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.7.0
	github.com/soheilhy/cmux v0.1.4
	github.com/sony/gobreaker v0.4.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tdex-network/tdex-daemon/pkg/explorer v0.0.0-20210322164638-77a31ea9e66d
	github.com/tdex-network/tdex-protobuf v0.0.0-20210324125236-78f6c6c0618b
	github.com/thanhpk/randstr v1.0.4
	github.com/timshannon/badgerhold/v2 v2.0.0-20201016201833-94bc303c76d4
	github.com/urfave/cli/v2 v2.3.0
	github.com/vulpemventures/go-bip39 v1.0.2
	github.com/vulpemventures/go-elements v0.2.1-0.20210409173614-5b1acc1d1e95
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
)
