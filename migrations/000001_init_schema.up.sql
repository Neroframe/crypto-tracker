CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE currencies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  symbol VARCHAR(10) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE currency_prices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  currency_id UUID NOT NULL REFERENCES currencies(id) ON DELETE CASCADE,
  timestamp BIGINT NOT NULL, -- Unix seconds
  price NUMERIC(20,10) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(currency_id, timestamp)
);

CREATE INDEX idx_currency_prices_currency_timestamp
  ON currency_prices (currency_id, timestamp);