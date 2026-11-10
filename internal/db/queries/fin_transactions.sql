-- ==============================================================
-- finance.campaign_invoices
-- finance.campaign_transactions
-- finance.organisation_pending_transactions
-- finance.organisation_invoices DML
-- File: fin_transactions.sql
-- ==============================================================

-- ========================================
-- finance.campaign_invoices
-- ========================================

-- name: CreateCampaignPendingTransaction :one
INSERT INTO finance.campaign_invoices (
    campaign_id, contestant_id, transaction_reference, payment_method, amount, currency, paid_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateCampaignPendingTransactionStatus :one
UPDATE finance.campaign_invoices
SET invoice_status = $2
WHERE id = $1
RETURNING *;

-- name: GetCampaignPendingTransactionByRef :one
SELECT * FROM finance.campaign_invoices WHERE transaction_reference = $1 LIMIT 1;

-- name: ListPendingTransactionsByCampaign :many
SELECT * FROM finance.campaign_invoices WHERE campaign_id = $1;

-- ========================================
-- finance.campaign_transactions
-- ========================================

-- name: FinaliseCampaignTransaction :one
INSERT INTO finance.campaign_transactions (
    campaign_id, contestant_id, transaction_reference, payment_method, nuban_from_account_number, nuban_from_bank_code, nuban_from_account_name, nuban_to_account_number, nuban_to_bank_code, nuban_to_account_name, payment_provider, status, campaign_balance_before, campaign_balance_after, amount, currency, entry_type, purpose, paid_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
) RETURNING *;

-- name: GetCampaignTransactionByRef :one
SELECT * FROM finance.campaign_transactions WHERE transaction_reference = $1 LIMIT 1;

-- name: ListCampaignTransactions :many
SELECT * FROM finance.campaign_transactions WHERE campaign_id = $1;

-- name: ListContestantTransactions :many
SELECT * FROM finance.campaign_transactions WHERE contestant_id = $1;

-- ========================================
-- finance.organisation_pending_transactions
-- ========================================

-- name: CreateOrganisationPendingTransaction :one
INSERT INTO finance.organisation_pending_transactions (
    organisation_id, transaction_reference, entry_type, amount, currency, nuban_target_account_number, nuban_target_bank_code, nuban_target_account_name, purpose, paid_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: UpdateOrganisationPendingTransactionStatus :one
UPDATE finance.organisation_pending_transactions
SET status = $2
WHERE id = $1
RETURNING *;

-- name: GetOrgPendingTransactionByRef :one
SELECT * FROM finance.organisation_pending_transactions WHERE transaction_reference = $1 LIMIT 1;

-- name: ListOrgPendingTransactions :many
SELECT * FROM finance.organisation_pending_transactions WHERE organisation_id = $1;

-- ========================================
-- finance.organisation_invoices
-- ========================================

-- name: FinaliseOrganisationTransaction :one
INSERT INTO finance.organisation_invoices (
    organisation_id, transaction_reference, entry_type, currency, amount, nuban_target_account_number, nuban_target_bank_code, nuban_target_account_name, purpose, paid_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetOrganisationTransactionByRef :one
SELECT * FROM finance.organisation_invoices WHERE transaction_reference = $1 LIMIT 1;

-- name: ListOrganisationTransactions :many
SELECT * FROM finance.organisation_invoices WHERE organisation_id = $1;