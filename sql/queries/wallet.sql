-- name: GetWallet :one
SELECT id, balance, created_at, updated_at
FROM wallets
WHERE id = $1;

-- name: GetWalletForUpdate :one
SELECT id, balance, created_at, updated_at
FROM wallets
WHERE id = $1
    FOR UPDATE;

-- name: CreateWallet :one
INSERT INTO wallets (id, balance)
VALUES ($1, 0)
RETURNING id, balance, created_at, updated_at;

-- name: UpdateWalletBalance :one
UPDATE wallets
SET balance   = $1,
    updated_at = NOW()
WHERE id = $2
RETURNING id, balance, created_at, updated_at;
