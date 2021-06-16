# first image used to build the sources
FROM golang:1.16-buster AS builder

ARG VERSION
ARG COMMIT
ARG DATE


ENV GO111MODULE=on \
    GOOS=linux \
    CGO_ENABLED=1 \
    GOARCH=amd64

WORKDIR /tdex-daemon

COPY . .
RUN go mod download

RUN go build -ldflags="-s -w " -o tdexd-linux cmd/tdexd/main.go
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o tdex cmd/tdex/*

WORKDIR /build

RUN cp /tdex-daemon/tdexd-linux .
RUN cp /tdex-daemon/tdex .

# Second image, running the tdexd executable
FROM debian:buster

# TDEX environment variables 
# default data directory path is overwrite
# others ENV variables are initialized to empty values: viper will initialize them.
ENV TDEX_DATA_DIR_PATH="/.tdex-daemon" \
    TDEX_EXPLORER_ENDPOINT= \
    TDEX_LOG_LEVEL= \
    TDEX_DEFAULT_FEE= \
    TDEX_NETWORK= \
    TDEX_BASE_ASSET= \
    TDEX_CRAWL_INTERVAL= \
    TDEX_FEE_ACCOUNT_BALANCE_TRESHOLD= \
    TDEX_TRADE_EXPIRY_TIME= \
    TDEX_PRICE_SLIPPAGE= \
    TDEX_MNEMONIC= \
    TDEX_UNSPENT_TTL= \
    TDEX_SSL_KEY= \
    TDEX_SSL_CERT=

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

COPY --from=builder /build/tdexd-linux /
COPY --from=builder /build/tdex /

RUN install /tdex /bin

# expose trader and operator interface ports
EXPOSE 9945
EXPOSE 9000

CMD /tdexd-linux

