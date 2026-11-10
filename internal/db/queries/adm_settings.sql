-- ==============================================================
-- admin.settings DML
-- File: adm_settings.sql
-- ==============================================================

-- name: UpsertAdminSetting :one
INSERT INTO admin.settings (
    key, value, description, updated_by
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (key) DO UPDATE SET 
    value = EXCLUDED.value,
    description = EXCLUDED.description,
    updated_by = EXCLUDED.updated_by,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetAdminSetting :one
SELECT * FROM admin.settings WHERE key = $1 LIMIT 1;

-- name: ListAdminSettings :many
SELECT * FROM admin.settings;