-- name: CreateUser :one
INSERT INTO accounts.users (
    "id",
    "first_name", 
    "middle_name", 
    "last_name", 
    "date_of_birth", 
    "gender",
    "email", 
    "phone",
    "is_whatsapp_phone", 
    "status",
    "hash_type", 
    "hash_password", 
    "hash_pin",
    "hash_table_seq"
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, $14
)
RETURNING *;


-- name: GetUserHashes :one
SELECT
    "id",
    "status",
    "hash_type",
    "hash_password",
    "hash_pin",
    "hash_table_seq"
FROM "accounts"."users"
WHERE 
    "phone" = $1
    OR "email" = $2
    OR "id" = $3
LIMIT 1;

-- name: GetUserHashesByEmail :one
SELECT
    "id",
    "status",
    "hash_type",
    "hash_password",
    "hash_pin",
    "hash_table_seq"
FROM "accounts"."users" 
WHERE "email" = $1;

-- name: GetUserHashesByID :one
SELECT
    "id",
    "status",
    "hash_type",
    "hash_password",
    "hash_pin",
    "hash_table_seq"
FROM "accounts"."users" 
WHERE "id" = $1;

-- name: GetUserHashesByPhone :one
SELECT
    "id",
    "status",
    "hash_type",
    "hash_password",
    "hash_pin",
    "hash_table_seq"
FROM "accounts"."users" 
WHERE "phone" = $1;



-- name: GetUserStatusByEmail :one
SELECT
 "id",
 "status",
 "phone",
 "email"
FROM "accounts"."users" 
WHERE "email" = $1;

-- name: GetUserStatusByPhone :one
SELECT
 "id",
 "status",
 "phone",
 "email"
FROM "accounts"."users" 
WHERE "phone" = $1;



-- Login by Email with Password
-- name: GetUserByEmailN :one
SELECT 
    "id",
    "email",
    "hash_password",
    "hash_type",
    "hash_table_seq",
    "status",
    "first_name",
    "last_name",
    "is_email_verified",
    "is_phone_verified"
FROM "accounts"."users" 
WHERE "email" = $1 
    AND status != 'deleted'
    LIMIT 1;

-- Login by Email with Password
-- name: GetUserByEmail :one
SELECT 
    "id",
    "email",
    "hash_password",
    "hash_type",
    "hash_table_seq",
    "status",
    "first_name",
    "last_name",
    "is_email_verified",
    "is_phone_verified"
FROM "accounts"."users" 
WHERE email = $1;





-- name: GetUserByPhoneN :one
SELECT 
    "id",
    "email",
    "hash_password",
    "hash_type",
    "hash_table_seq",
    "status",
    "first_name",
    "last_name",
    "is_email_verified",
    "is_phone_verified"
FROM "accounts"."users" 
WHERE "phone" = $1 
    AND "status" != 'deleted'
    LIMIT 1;




-- name: UpdateUserVerification :one
UPDATE "accounts"."users"
SET 
    "is_phone_verified" = $2,
    "is_email_verified" = $3,
    "is_nin_verified" = $4
WHERE 
    "id" = $1
RETURNING "id";


-- name: GetUserVerificationByID :one
SELECT 
    "id",
    "status",
    "is_phone_verified",
    "is_email_verified", 
    "is_nin_verified" 
FROM "accounts"."users"
WHERE 
"id" = $1 
LIMIT 1;

-- name: GetUserVerificationByEmail :one
SELECT 
    "id",
    "status",
    "is_phone_verified",
    "is_email_verified", 
    "is_nin_verified" 
FROM "accounts"."users"
WHERE 
"email" = $1 
LIMIT 1;

-- name: GetUserVerificationByPhone :one
SELECT 
    "id",
    "status",
    "is_phone_verified",
    "is_email_verified", 
    "is_nin_verified" 
FROM "accounts"."users"
WHERE 
"phone" = $1 
LIMIT 1;


-- name: UpdateUserNIN :one
UPDATE "accounts"."users"
SET 
    "is_nin_verified" = TRUE,
    "encrypted_nin" = $2
WHERE
    "id" = $1
RETURNING 
"id", 
"is_nin_verified",
"hash_pin",
"encrypted_nin";