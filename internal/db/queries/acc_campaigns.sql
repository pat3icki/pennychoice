-- ==============================================================
-- campaigns
-- contestants DML
-- File: acc_campaigns.sql
-- ==============================================================

-- ========================================
-- campaigns
-- ========================================

-- name: CreateCampaign :one
INSERT INTO campaigns (
    id, organisation_id, creator_user_id, 
    name, description, status, bank_name_suffix, tags, 
    vote_currency, vote_per_one_amount, vote_metrics_visibility, 
    allow_refunds, verdict_partial_vote, platform_fee_percentage, 
    votes_starts_at, votes_ends_at, registration_fee, activation_method
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, 
    $10, $11, $12, $13, $14, $15, $16, $17, $18
) RETURNING *;

-- name: UpdateCampaignMetrics :one
UPDATE campaigns
SET 
    amount_raised = $2,
    total_votes = $3,
    total_contestants = $4
WHERE id = $1
RETURNING *;

-- name: UpdateCampaignStatus :one
UPDATE campaigns
SET status = $2
WHERE id = $1
RETURNING id;

-- name: GetCampaignByID :one
SELECT * FROM campaigns WHERE id = $1 LIMIT 1;

-- name: ListCampaignsByOrganisationID :many
SELECT * FROM campaigns WHERE organisation_id = $1;

-- name: ListCampaignsByStatus :many
SELECT * FROM campaigns WHERE status = $1;

-- name: ListCampaignsByDateRange :many
SELECT * FROM campaigns WHERE votes_starts_at >= $1 AND votes_ends_at <= $2;

-- ========================================
-- contestants
-- ========================================

-- name: AddContestant :one
INSERT INTO contestants (
    id, campaign_id, user_id, display_name, photo_root, social_links, checkout
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateContestantVotes :one
UPDATE contestants
SET 
    vote_count = $2,
    hv_per_transact = $3
WHERE id = $1
RETURNING *;


-- name: GetContestantRequirementData :one
SELECT requirement_data FROM request_campaign_contestants WHERE request_contestant_user_id = $1 LIMIT 1;


-- name: GetContestantByID :one
SELECT * FROM contestants WHERE id = $1 LIMIT 1;

-- name: GetContestantByUserAndCampaign :one
SELECT * FROM contestants WHERE campaign_id = $1 AND user_id = $2 LIMIT 1;

-- name: ListContestantsByCampaign :many
SELECT * FROM contestants WHERE campaign_id = $1;

-- name: ListContestantsByUser :many
SELECT * FROM contestants WHERE user_id = $1;


-- name: UpdateCampaignContestantCount :exec
UPDATE campaigns
SET 
    total_contestants = $2
WHERE id = $1;

-- name: ExistCampaignReturn :one
SELECT c.*
FROM campaigns c
WHERE c.id = $3
  AND c.organisation_id = $2
  AND EXISTS (
      SELECT 1
      FROM organisations_members om
      WHERE om.org_id = c.organisation_id
        AND om.member_id = $1
  );

-- name: AdjustVoteDuration :exec
UPDATE campaigns
SET votes_starts_at = $1,
    votes_ends_at   = $2
WHERE id = $3
  AND organisation_id = $4
RETURNING *;

