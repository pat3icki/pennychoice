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
