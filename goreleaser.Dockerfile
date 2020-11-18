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
    TDEX_UNSPENT_TTL=

WORKDIR /tdex-daemon

COPY tdexd-linux /
COPY tdex /

RUN install /tdex /bin

# expose trader and operator interface ports
EXPOSE 9945
EXPOSE 9000

CMD /tdexd-linux

