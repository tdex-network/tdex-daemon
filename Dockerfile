# first image used to build the sources
FROM golang:1.19-buster AS builder

ARG VERSION
ARG COMMIT
ARG DATE
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o bin/tdexd cmd/tdexd/*
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o bin/tdex cmd/tdex/*
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o bin/tdex-migration cmd/migration/main.go

# Second image, running the tdexd executable
FROM debian:buster-slim

# $USER name, and data $DIR to be used in the `final` image
ARG USER=tdex
ARG DIR=/home/tdex

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

# NOTE: Default GID == UID == 1000
RUN adduser --disabled-password \
            --home "$DIR/" \
            --gecos "" \
            "$USER"
USER $USER

COPY --from=builder /app/bin/* /usr/local/bin/
COPY web/layout.html web/layout.html

# Prevents `VOLUME $DIR/.tdex-daemon/` being created as owned by `root`
RUN mkdir -p "$DIR/.tdex-daemon/"

# Expose volume containing all `tdexd` data
VOLUME $DIR/.tdex-daemon/

# expose trader and operator interface ports
EXPOSE 9945
EXPOSE 9000

ENTRYPOINT [ "tdexd" ]

