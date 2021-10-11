module github.com/tdex-network/tdex-daemon

go 1.16

replace github.com/tdex-network/tdex-daemon/pkg/explorer => ./pkg/explorer

replace github.com/tdex-network/tdex-daemon/pkg/securestore => ./pkg/securestore

replace github.com/tdex-network/tdex-daemon/pkg/macaroons => ./pkg/macaroons

require (
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/gogo/protobuf v1.2.1
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/improbable-eng/grpc-web v0.13.0
	github.com/magiconair/properties v1.8.4
	github.com/mdp/qrterminal/v3 v3.0.0
	github.com/prometheus/client_golang v1.0.0
	github.com/rs/cors v1.7.0 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/soheilhy/cmux v0.1.5
	github.com/sony/gobreaker v0.4.1
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tdex-network/tdex-daemon/pkg/explorer v0.0.0-20211001103242-a11e4485705a
	github.com/tdex-network/tdex-daemon/pkg/macaroons v0.0.0-20210813140257-70d50a8b72a4
	github.com/tdex-network/tdex-daemon/pkg/securestore v0.0.0-20210813140257-70d50a8b72a4
	github.com/tdex-network/tdex-protobuf v0.0.0-20210507104156-d509331cccdb
	github.com/thanhpk/randstr v1.0.4
	github.com/timshannon/badgerhold/v2 v2.0.0-20201016201833-94bc303c76d4
	github.com/urfave/cli/v2 v2.3.0
	github.com/vulpemventures/go-bip39 v1.0.2
	github.com/vulpemventures/go-elements v0.3.3
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20210908233432-aa78b53d3365 // indirect
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/macaroon-bakery.v2 v2.3.0
	gopkg.in/macaroon.v2 v2.1.0
)
