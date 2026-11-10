-- ==============================================================
-- finance.webhook DML
-- File: fin_webhook.sql
-- ==============================================================

-- name: LogWebhook :one
INSERT INTO finance.webhook (
    event_type, payload
) VALUES (
    $1, $2
) RETURNING *;

-- name: ListWebhooksByEvent :many
SELECT * FROM finance.webhook WHERE event_type = $1;