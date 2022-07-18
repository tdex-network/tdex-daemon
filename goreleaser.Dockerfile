FROM debian:buster-slim

WORKDIR /app

COPY . .

ARG OS
ARG ARCH

ENV TDEXD="tdexd-${OS}-${ARCH}"

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

# Prevents `VOLUME $DIR/.tdex-daemon/` being created as owned by `root`
RUN mkdir -p "$DIR/.tdex-daemon/"

# Expose volume containing all `tdexd` data
VOLUME $DIR/.tdex-daemon/

# expose trader and operator interface ports
EXPOSE 9945
EXPOSE 9000

ENTRYPOINT ./$TDEXD
