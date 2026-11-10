-- ==============================================================
-- log.unified_audit DML
-- File: log_unified_audit.sql
-- ==============================================================

-- name: CreateUnifiedAuditEntry :one
INSERT INTO log.unified_audit (
    occurred_at, 
    actor_type, 
    actor_id, 
    action_type, 
    table_name, 
    affected_row_id, 
    old_data, 
    new_data, 
    ip_address, 
    user_agent, 
    reason, 
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: ListUnifiedAuditsByActor :many
SELECT * FROM log.unified_audit WHERE actor_type = $1 AND actor_id = $2;

-- name: ListUnifiedAuditsByRecordID :many
SELECT * FROM log.unified_audit WHERE table_name = $1 AND affected_row_id = $2;
