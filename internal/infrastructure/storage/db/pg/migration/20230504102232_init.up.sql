CREATE TABLE market (
    name VARCHAR(50) NOT NULL PRIMARY KEY,
    base_asset VARCHAR(255) NOT NULL,
    quote_asset VARCHAR(255) NOT NULL,
    base_asset_precision INTEGER NOT NULL DEFAULT 8,
    quote_asset_precision INTEGER NOT NULL DEFAULT 8,
    tradeable BOOLEAN NOT NULL DEFAULT FALSE,
    strategy_type INTEGER NOT NULL DEFAULT 0,
    base_price INTEGER NOT NULL DEFAULT 0,
    quote_price INTEGER NOT NULL DEFAULT 0,
    UNIQUE (base_asset, quote_asset)
);

CREATE TABLE fee (
    id SERIAL PRIMARY KEY,
    base_asset_fee BIGINT NOT NULL,
    quote_asset_fee BIGINT NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('percentage', 'fixed')),
    fk_market_name VARCHAR(50) NOT NULL,
    FOREIGN KEY (fk_market_name) REFERENCES market (name)
);

CREATE TABLE trade (
    id VARCHAR(50) PRIMARY KEY,
    type INTEGER NOT NULL,
    fee_asset VARCHAR(255) NOT NULL,
    fee_amount BIGINT NOT NULL,
    trader_pubkey BYTEA NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 0,
    status_failed BOOLEAN NOT NULL DEFAULT FALSE,
    pset_base64 TEXT NOT NULL,
    tx_id VARCHAR(255) NOT NULL,
    tx_hex TEXT NOT NULL,
    expiry_time BIGINT NOT NULL,
    settlement_time BIGINT NOT NULL,
    fk_market_name VARCHAR(50) NOT NULL,
    FOREIGN KEY (fk_market_name) REFERENCES market (name)
);

CREATE TABLE swap (
    id TEXT PRIMARY KEY,
    message BYTEA NOT NULL,
    timestamp BIGINT NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('request', 'accept', 'complete', 'fail')),
    fk_trade_id VARCHAR(50) NOT NULL,
    FOREIGN KEY (fk_trade_id) REFERENCES trade (id)
);

CREATE TABLE transaction (
    id SERIAL PRIMARY KEY,
    type VARCHAR(10) NOT NULL CHECK (type IN ('withdrawal', 'deposit')),
    account_name TEXT NOT NULL,
    tx_id TEXT NOT NULL,
    timestamp BIGINT NOT NULL
);

CREATE TABLE transaction_asset_amount (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL,
    asset TEXT NOT NULL,
    amount BIGINT NOT NULL,
    FOREIGN KEY (transaction_id) REFERENCES transaction (id)
);