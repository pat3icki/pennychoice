-- ==============================================================
-- users DML
-- File: acc_users.sql
-- ==============================================================

-- name: ExistsUserPhone :one
-- Case-insensitive check (if needed)
SELECT EXISTS (
    SELECT 1 
    FROM users 
    WHERE LOWER(phone) = LOWER($1)
) AS phone_exists;


-- name: CreateUserProfile :one
INSERT INTO users (
    id, first_name, middle_name, last_name, date_of_birth, gender, status, email, country, signup_stages
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: CreateUserProfileFULL :one
INSERT INTO users (
    id, first_name, middle_name, last_name, date_of_birth, gender, status, email, phone, country, is_whatsapp_phone, hash_password, hash_pin, signup_stages
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: CreateUserProfileEssential :one
INSERT INTO users (
    id, first_name, middle_name, last_name, gender, email, country, hash_password
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateVerificationAttributes :one
UPDATE users
SET 
    is_phone_verified = COALESCE(sqlc.narg('is_phone_verified'), is_phone_verified),
    is_email_verified = COALESCE(sqlc.narg('is_email_verified'), is_email_verified),
    is_identity_verified = COALESCE(sqlc.narg('is_identity_verified'), is_identity_verified),
    identity_document_type = COALESCE(sqlc.narg('identity_document_type'), identity_document_type),
    encrypted_identity_data = COALESCE(sqlc.narg('encrypted_identity_data'), encrypted_identity_data),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, is_email_verified, is_phone_verified, is_identity_verified, identity_document_type;

-- name: UpdateUserSecurities :one
UPDATE users
SET 
    hash_password = $2,
    hash_pin = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUserIdentityData :one
UPDATE users
SET 
    is_identity_verified = $2,
    identity_document_type = $3,
    encrypted_identity_data = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUserPhone :one
UPDATE users
SET 
    phone = $2,
    is_whatsapp_phone = $3,
    is_phone_verified = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUserSoftDelete :one
UPDATE users
SET 
    status = $2,
    deactivated_at = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUserSignupStages :one
UPDATE users
SET 
    signup_stages = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = $1 LIMIT 1;

-- name: ListUsersByStatus :many
SELECT * FROM users WHERE status = $1;

-- name: GetUserVerificationAttributesByID :one
SELECT id, is_email_verified, is_phone_verified, is_identity_verified, identity_document_type, encrypted_identity_data 
FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserSecurities :one
SELECT id, hash_password, hash_pin 
FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserSignupStages :one
SELECT id, signup_stages 
FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserCredentialByEmail :one
SELECT id, email, phone, hash_password, signup_stages
FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserCredentialByPhone :one
SELECT id, email, phone, hash_password, signup_stages
FROM users WHERE phone = $1 LIMIT 1;



-- name: GetUserIDByEmail :one
SELECT id 
FROM users 
WHERE email = $1;

-- name: GetUserIDByPhone :one
SELECT id 
FROM users 
WHERE phone = $1;

-- name: GetUserIDByEmailOrPhone :one
SELECT id 
FROM users 
WHERE email = $1 OR phone = $2;

-- name: GetUserIDByEmailOrPhoneNullable :one
SELECT id 
FROM users 
WHERE email = $1 OR phone = $1;


-- name: CheckUserExistsByEmail :one
SELECT EXISTS(
    SELECT 1 
    FROM users 
    WHERE email = $1
) AS exists;

-- name: CheckUserExistsByPhone :one
SELECT EXISTS(
    SELECT 1 
    FROM users 
    WHERE phone = $1
) AS exists;

-- name: CheckUserExistsByEmailOrPhone :one
SELECT EXISTS(
    SELECT 1 
    FROM users 
    WHERE email = $1 OR phone = $1
) AS exists;

-- name: GetUserIDByEmailForUpdate :one
SELECT id 
FROM users 
WHERE email = $1
FOR UPDATE;

-- name: GetUserIDByPhoneForUpdate :one
SELECT id 
FROM users 
WHERE phone = $1
FOR UPDATE;



-- name: CreateUserContestantProfile :one
INSERT INTO user_contestant_profiles (
    user_id,
    profile_name,
    display_name,
    bio,
    photo_root,
    cover_photo,
    social_links,
    website_url,
    location,
    organisation_name,
    category,
    is_active,
    profile_completed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetUserContestantProfileByID :one
SELECT * FROM user_contestant_profiles
WHERE id = $1 AND is_active = true;

-- name: GetUserContestantProfileByUserID :one
SELECT * FROM user_contestant_profiles
WHERE user_id = $1 AND is_active = true;


-- name: GetUserContestantProfileByUserAndName :one
SELECT * FROM user_contestant_profiles
WHERE user_id = $1 AND profile_name = $2 AND is_active = true;

-- name: GetUserContestantProfilesByUser :many
SELECT * FROM user_contestant_profiles
WHERE user_id = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: GetUserContestantProfilesByCategory :many
SELECT * FROM user_contestant_profiles
WHERE category = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetActiveUserContestantProfiles :many
SELECT * FROM user_contestant_profiles
WHERE is_active = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUserContestantProfile :one
UPDATE user_contestant_profiles
SET 
    profile_name = COALESCE($1, profile_name),
    display_name = COALESCE($2, display_name),
    bio = COALESCE($3, bio),
    photo_root = COALESCE($4, photo_root),
    cover_photo = COALESCE($5, cover_photo),
    social_links = COALESCE($6, social_links),
    website_url = COALESCE($7, website_url),
    location = COALESCE($8, location),
    organisation_name = COALESCE($9, organisation_name),
    category = COALESCE($10, category),
    is_active = COALESCE($11, is_active),
    profile_completed_at = COALESCE($12, profile_completed_at)
WHERE id = $13
RETURNING *;

-- name: UpdateUserContestantProfilePartial :one
UPDATE user_contestant_profiles
SET 
    display_name = COALESCE(sqlc.narg('display_name'), display_name),
    bio = COALESCE(sqlc.narg('bio'), bio),
    photo_root = COALESCE(sqlc.narg('photo_root'), photo_root),
    cover_photo = COALESCE(sqlc.narg('cover_photo'), cover_photo),
    social_links = COALESCE(sqlc.narg('social_links'), social_links),
    website_url = COALESCE(sqlc.narg('website_url'), website_url),
    location = COALESCE(sqlc.narg('location'), location),
    organisation_name = COALESCE(sqlc.narg('organisation_name'), organisation_name),
    category = COALESCE(sqlc.narg('category'), category)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: MarkUserContestantProfileCompleted :one
UPDATE user_contestant_profiles
SET 
    profile_completed_at = CURRENT_TIMESTAMP,
    is_active = true
WHERE id = $1
RETURNING *;

-- name: DeactivateUserContestantProfile :exec
UPDATE user_contestant_profiles
SET is_active = false
WHERE id = $1;

-- name: DeleteUserContestantProfile :exec
DELETE FROM user_contestant_profiles
WHERE id = $1;

-- name: DeleteUserContestantProfilesByUser :exec
DELETE FROM user_contestant_profiles
WHERE user_id = $1;

-- name: CountUserContestantProfilesByUser :one
SELECT COUNT(*) FROM user_contestant_profiles
WHERE user_id = $1 AND is_active = true;

-- name: CountUserContestantProfilesByCategory :one
SELECT COUNT(*) FROM user_contestant_profiles
WHERE category = $1 AND is_active = true;

-- name: CheckUserContestantProfileExists :one
SELECT EXISTS (
    SELECT 1 FROM user_contestant_profiles
    WHERE user_id = $1 AND profile_name = $2 AND is_active = true
);

-- name: GetUserContestantProfilesWithPagination :many
SELECT * FROM user_contestant_profiles
WHERE is_active = true
AND (profile_name ILIKE '%' || $1 || '%' 
     OR display_name ILIKE '%' || $1 || '%'
     OR organisation_name ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: BulkCreateUserContestantProfiles :copyfrom
INSERT INTO user_contestant_profiles (
    user_id,
    profile_name,
    display_name,
    bio,
    photo_root,
    cover_photo,
    social_links,
    website_url,
    location,
    organisation_name,
    category,
    is_active,
    profile_completed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
);

-- name: UpsertUserContestantProfile :one
INSERT INTO user_contestant_profiles (
    user_id,
    profile_name,
    display_name,
    bio,
    photo_root,
    cover_photo,
    social_links,
    website_url,
    location,
    organisation_name,
    category,
    is_active,
    profile_completed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (user_id, profile_name) 
DO UPDATE SET
    display_name = EXCLUDED.display_name,
    bio = EXCLUDED.bio,
    photo_root = EXCLUDED.photo_root,
    cover_photo = EXCLUDED.cover_photo,
    social_links = EXCLUDED.social_links,
    website_url = EXCLUDED.website_url,
    location = EXCLUDED.location,
    organisation_name = EXCLUDED.organisation_name,
    category = EXCLUDED.category,
    is_active = EXCLUDED.is_active,
    profile_completed_at = EXCLUDED.profile_completed_at
RETURNING *;