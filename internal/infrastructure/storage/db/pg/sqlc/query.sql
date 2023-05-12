/** Market **/

-- name: InsertMarket :one
INSERT INTO market (name,base_asset,quote_asset,base_asset_precision,
quote_asset_precision,tradable,strategy_type,base_price,quote_price, active) VALUES
($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: GetMarketByName :many
SELECT * from market m inner join market_fee f on m.name = f.fk_market_name where name = $1 and active = true;

-- name: GetMarketByBaseAndQuoteAsset :many
SELECT * from market m inner join market_fee f on m.name = f.fk_market_name where
base_asset = $1 and quote_asset = $2 and active = true;

-- name: GetTradableMarkets :many
SELECT * FROM market m inner join market_fee f on m.name = f.fk_market_name
where tradable = true and active = true;

-- name: GetAllMarkets :many
SELECT * FROM market m inner join market_fee f on m.name = f.fk_market_name and active = true;;

-- name: UpdateMarket :one
UPDATE market SET base_asset_precision = $1, quote_asset_precision = $2,
tradable = $3, strategy_type = $4, base_price = $5, quote_price = $6 WHERE
name = $7 RETURNING *;

-- name: InactivateMarket :exec
UPDATE market SET active=false WHERE name = $1;

-- name: UpdateMarketPrice :one
UPDATE market SET base_price = $1, quote_price = $2 WHERE name = $3 RETURNING *;

/** MarketFee **/

-- name: InsertMarketFee :one
INSERT INTO market_fee (base_asset_fee, quote_asset_fee, type, fk_market_name) VALUES
($1, $2, $3, $4) RETURNING *;

-- name: GetFeeByMarketName :one
SELECT * FROM market_fee WHERE fk_market_name = $1;

-- name: UpdateMarketFee :one
UPDATE market_fee SET base_asset_fee = $1, quote_asset_fee = $2 WHERE
fk_market_name = $3 and type = $4 RETURNING *;

/** Trade **/

-- name: InsertTrade :one
INSERT INTO trade (id, type, fee_asset, fee_amount, trader_pubkey, status_code,
status_failed, pset_base64, tx_id, tx_hex, expiry_time, settlement_time, base_price, quote_price,
fk_market_name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: GetAllTrades :many
SELECT t.*, m.name, m.base_asset, m.quote_asset,
       s.id as swap_id, s.message, s.timestamp, s.type as swap_type, f.base_asset_fee,
       f.quote_asset_fee, f.type as fee_type
FROM trade t left join swap s on t.id = s.fk_trade_id inner join
     market m on m.name = t.fk_market_name left join trade_fee f on t.id = f.fk_trade_id
order by t.id DESC LIMIT $1 OFFSET $2;

-- name: GetTradeById :many
SELECT t.*, m.name, m.base_asset, m.quote_asset,
s.id as swap_id, s.message, s.timestamp, s.type as swap_type, f.base_asset_fee,
       f.quote_asset_fee, f.type as fee_type
FROM trade t left join swap s on t.id = s.fk_trade_id inner join
market m on m.name = t.fk_market_name left join trade_fee f on t.id = f.fk_trade_id
WHERE t.id = $1;

-- name: GetAllTradesByMarket :many
SELECT t.*, m.name, m.base_asset, m.quote_asset,
       s.id as swap_id, s.message, s.timestamp, s.type as swap_type, f.base_asset_fee,
        f.quote_asset_fee, f.type as fee_type
FROM trade t left join swap s on t.id = s.fk_trade_id inner join
     market m on m.name = t.fk_market_name left join trade_fee f on t.id = f.fk_trade_id
WHERE m.name = $1 order by t.id DESC LIMIT $2 OFFSET $3;

-- name: GetTradeBySwapAcceptId :many
SELECT t.*, m.name, m.base_asset, m.quote_asset,
       s.id as swap_id, s.message, s.timestamp, s.type as swap_type, f.base_asset_fee,
       f.quote_asset_fee, f.type as fee_type
FROM trade t left join swap s on t.id = s.fk_trade_id inner join
     market m on m.name = t.fk_market_name left join trade_fee f on t.id = f.fk_trade_id
WHERE t.id in (select t.id from trade t inner join swap s on t.id = s.fk_trade_id where s.id = $1);

-- name: GetTradeByTxId :many
SELECT t.*, m.name, m.base_asset, m.quote_asset,
       s.id as swap_id, s.message, s.timestamp, s.type as swap_type, f.base_asset_fee,
       f.quote_asset_fee, f.type as fee_type
FROM trade t left join swap s on t.id = s.fk_trade_id inner join
     market m on m.name = t.fk_market_name left join trade_fee f on t.id = f.fk_trade_id
WHERE t.tx_id = $1;

-- name: GetTradesByMarketAndStatus :many
SELECT t.*, m.name, m.base_asset, m.quote_asset,
       s.id as swap_id, s.message, s.timestamp, s.type as swap_type, f.base_asset_fee,
       f.quote_asset_fee, f.type as fee_type
FROM trade t left join swap s on t.id = s.fk_trade_id inner join
     market m on m.name = t.fk_market_name left join trade_fee f on t.id = f.fk_trade_id
WHERE m.name = $1 AND t.status_code = $2 and t.status_failed = $3 order by t.id DESC LIMIT $4 OFFSET $5;

-- name: UpdateTrade :one
UPDATE trade SET type = $1, fee_asset = $2, fee_amount = $3, trader_pubkey = $4,
status_code = $5, status_failed = $6, pset_base64 = $7, tx_id = $8, tx_hex = $9,
expiry_time = $10, settlement_time = $11 WHERE id = $12 RETURNING *;

/** TradeFee **/

-- name: InsertTradeFee :one
INSERT INTO trade_fee (base_asset_fee, quote_asset_fee, type, fk_trade_id) VALUES
    ($1, $2, $3, $4) RETURNING *;


/** Swap **/

-- name: InsertSwap :one
INSERT INTO swap (id, message, timestamp, type, fk_trade_id) VALUES
($1, $2, $3, $4, $5) RETURNING *;

-- name: DeleteSwapsByTradeId :exec
DELETE FROM swap WHERE fk_trade_id = $1;

/** Transaction **/

-- name: InsertTransaction :one
INSERT INTO transaction (type, account_name, tx_id, timestamp) VALUES
($1, $2, $3, $4) RETURNING *;

-- name: InsertTransactionAssetAmount :one
INSERT INTO transaction_asset_amount (fk_transaction_id, asset, amount) VALUES
($1, $2, $3) RETURNING *;

-- name: GetAllTransactions :many
SELECT t.type, t.account_name, t.tx_id, t.timestamp, taa.asset, taa.amount
FROM transaction t left join transaction_asset_amount taa on t.id = taa.fk_transaction_id
WHERE t.type = $1 ORDER BY timestamp DESC LIMIT $2 OFFSET $3;

-- name: GetAllTransactionsForAccountNameAndPage :many
SELECT t.type, t.account_name, t.tx_id, t.timestamp, taa.asset, taa.amount
FROM transaction t left join transaction_asset_amount taa on t.id = taa.fk_transaction_id
WHERE t.type = $1 AND account_name = $2 ORDER BY timestamp DESC LIMIT $3 OFFSET $4;