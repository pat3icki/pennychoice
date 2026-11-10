-- ==============================================================
-- admin.users DML
-- File: adm_users.sql
-- ==============================================================

-- name: CreateAdminUser :one
INSERT INTO admin.users (
    id, first_name, last_name, email, username, role, status, hash_password
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateAdminUserStatus :one
UPDATE admin.users
SET 
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateAdminUserRole :one
UPDATE admin.users
SET 
    role = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetAdminUserByID :one
SELECT * FROM admin.users WHERE id = $1 LIMIT 1;

-- name: GetAdminUserByEmail :one
SELECT * FROM admin.users WHERE email = $1 LIMIT 1;

-- name: GetAdminUserByUsername :one
SELECT * FROM admin.users WHERE username = $1 LIMIT 1;