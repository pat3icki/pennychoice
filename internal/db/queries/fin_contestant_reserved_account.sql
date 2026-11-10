-- ==============================================================
-- finance.contestant_reserved_account DML
-- File: fin_contestant_reserved_account.sql
-- ==============================================================

-- name: CreateContestantReservedAccount :one
INSERT INTO finance.contestant_reserved_account (
    contestant_id, account_reference, account_name, account_number, currency_code, payment_service_payload
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateContestantReservedAccount :one
UPDATE finance.contestant_reserved_account
SET 
    account_reference = $2,
    account_name = $3,
    account_number = $4,
    payment_service_payload = $5
WHERE contestant_id = $1
RETURNING *;

-- name: GetReservedAccountByContestant :one
SELECT * FROM finance.contestant_reserved_account WHERE contestant_id = $1 LIMIT 1;