-- ==============================================================
-- organisations
-- organisations_members
-- organisation_decisions DML
-- File: acc_organisations.sql
-- ==============================================================

-- ========================================
-- organisations
-- ========================================

-- name: CreateOrganisation :one
INSERT INTO organisations (
    id, name, description, creator_user_id, permissions
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateOrganisationFinancials :one
UPDATE organisations
SET 
    balance = $2,
    total_contributions = $3
WHERE id = $1
RETURNING *;

-- name: UpdateOrganisationStats :one
UPDATE organisations
SET 
    total_members = $2,
    total_campaign = $3,
    active_campaign = $4
WHERE id = $1
RETURNING *;

-- name: GetOrganisationPermissions :one 
SELECT permissions FROM organisations WHERE id = $1;


-- name: UpdateOrganisationPermissions :exec 
UPDATE organisations SET permissions = $2 WHERE id = $1;

-- name: GetOrganisationByMember :one
SELECT o.*
FROM organisations o
INNER JOIN organisations_members om ON o.id = om.org_id
WHERE om.member_id = $2
  AND om.org_id = $1 LIMIT 1;

-- name: GetOrganisationByName :one
SELECT * FROM organisations WHERE name = $1 LIMIT 1;

-- name: GetOrganisationByID :one
SELECT * FROM organisations WHERE id = $1 LIMIT 1;

-- name: ListOrganisationByUser :many
SELECT o.*
FROM organisations o
INNER JOIN organisations_members om ON o.id = om.org_id
WHERE om.member_id = $1
LIMIT $2;

-- name: ListOrganisationsByCreator :many
SELECT * FROM organisations WHERE creator_user_id = $1;

-- ========================================
-- organisations_members
-- ========================================

-- name: AddOrganisationMember :one
INSERT INTO organisations_members (
    org_id, member_id, position, invited_by
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: UpdateOrganisationMemberPosition :one
UPDATE organisations_members
SET position = $3
WHERE org_id = $1 AND member_id = $2
RETURNING *;

-- name: GetOrganisationMember :one
SELECT * FROM organisations_members WHERE org_id = $1 AND member_id = $2 LIMIT 1;


-- name: ListOrganisationMembers :many
SELECT * FROM organisations_members WHERE org_id = $1;

-- name: ListUserOrganisations :many
SELECT * FROM organisations_members WHERE member_id = $1;

-- ========================================
-- organisation_decisions
-- ========================================

-- name: CreateOrganisationDecision :one
INSERT INTO organisation_decisions (
    organisation_id, campaign_id, action_type, message, admin_initiator_id, deadline
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateOrganisationDecisionStatus :one
UPDATE organisation_decisions
SET status = $2
WHERE id = $1
RETURNING *;

-- name: GetOrganisationDecisionByID :one
SELECT * FROM organisation_decisions WHERE id = $1 LIMIT 1;

-- name: ListDecisionsByOrganisation :many
SELECT * FROM organisation_decisions WHERE organisation_id = $1;

-- name: ListDecisionsByCampaign :many
SELECT * FROM organisation_decisions WHERE campaign_id = $1;


-- name: AddTotalCampaign :exec
UPDATE organisations 
SET total_campaign = total_campaign + 1
WHERE id = $1;

-- name: AddActiveCampaignInOrganisation :exec
UPDATE organisations 
SET active_campaign = active_campaign + 1
WHERE id = $1
AND active_campaign < total_campaign
AND active_campaign <= max_active_campaign;  -- Optional: ensures constraint is maintained