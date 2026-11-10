-- ==============================================================
-- notification.users DML
-- File: acc_notifications.sql
-- ==============================================================

-- name: CreateNotification :one
INSERT INTO notification.users (
    user_id, title, description, source, source_target_id
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateNotificationStatus :one
UPDATE notification.users
SET 
    status = $2,
    read_at = $3
WHERE id = $1
RETURNING *;

-- name: GetNotificationByID :one
SELECT * FROM notification.users WHERE id = $1 LIMIT 1;

-- name: ListNotificationsByUser :many
SELECT * FROM notification.users WHERE user_id = $1;

-- name: ListNotificationsByUserAndStatus :many
SELECT * FROM notification.users WHERE user_id = $1 AND status = $2;