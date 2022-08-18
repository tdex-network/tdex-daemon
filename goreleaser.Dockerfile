FROM debian:buster-slim

ARG TARGETPLATFORM


WORKDIR /app

COPY . .

RUN set -ex \
  && if [ "${TARGETPLATFORM}" = "linux/amd64" ]; then export TARGETPLATFORM=amd64; fi \
  && if [ "${TARGETPLATFORM}" = "linux/arm64" ]; then export TARGETPLATFORM=arm64; fi \
  && mv tdex /usr/local/bin/tdex \
  && mv tdexdconnect /usr/local/bin/tdexdconnect \
  && mv "tdexd-linux-$TARGETPLATFORM" /usr/local/bin/tdexd


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

COPY web/layout.html web/layout.html

# Prevents `VOLUME $DIR/.tdex-daemon/` being created as owned by `root`
RUN mkdir -p "$DIR/.tdex-daemon/"

# Expose volume containing all `tdexd` data
VOLUME $DIR/.tdex-daemon/

# expose trader and operator interface ports
EXPOSE 9945
EXPOSE 9000

ENTRYPOINT ["tdexd"]
