# first image used to build the sources
FROM golang:1.15.5-buster AS builder

ENV GO111MODULE=on \
    GOOS=linux \
    CGO_ENABLED=1 \
    GOARCH=amd64

WORKDIR /tdex-daemon

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o tdexd-linux cmd/tdexd/main.go
RUN go build -ldflags="-s -w" -o tdex cmd/tdex/*

WORKDIR /build

RUN cp /tdex-daemon/tdexd-linux .
RUN cp /tdex-daemon/tdex .

# Second image, running the tdexd executable
FROM debian:buster

COPY --from=builder /build/tdexd-linux /
COPY --from=builder /build/tdex /

RUN install /tdex /bin

# expose trader and operator interface ports
EXPOSE 9945
EXPOSE 9000

CMD /tdexd-linux

