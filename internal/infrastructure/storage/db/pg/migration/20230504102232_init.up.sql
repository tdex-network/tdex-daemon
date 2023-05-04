CREATE TABLE market (
    name VARCHAR(50) NOT NULL PRIMARY KEY,
    base_asset VARCHAR(64) NOT NULL,
    quote_asset VARCHAR(64) NOT NULL,
    base_asset_precision INTEGER,
    quote_asset_precision INTEGER,
    tradable BOOLEAN,
    strategy_type INTEGER,
    base_price INTEGER,
    quote_price INTEGER,
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
    trader_pubkey BYTEA,
    status_code INTEGER,
    status_failed BOOLEAN,
    pset_base64 TEXT NOT NULL,
    tx_id VARCHAR(64),
    tx_hex TEXT NOT NULL,
    expiry_time BIGINT,
    settlement_time BIGINT,
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
    account_name VARCHAR(50) NOT NULL,
    tx_id VARCHAR(64) NOT NULL,
    timestamp BIGINT NOT NULL
);

CREATE TABLE transaction_asset_amount (
    id SERIAL PRIMARY KEY,
    fk_transaction_id INTEGER NOT NULL,
    asset VARCHAR(64) NOT NULL,
    amount BIGINT NOT NULL,
    FOREIGN KEY (fk_transaction_id) REFERENCES transaction (id)
);