CREATE TABLE IF NOT EXISTS wallets (
                                       id          UUID           PRIMARY KEY,
                                       balance     NUMERIC(20, 2) NOT NULL DEFAULT 0,
                                       created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
                                       updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
                                       CONSTRAINT balance_non_negative CHECK (balance >= 0)
);