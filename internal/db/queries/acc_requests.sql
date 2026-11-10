-- ==============================================================
-- request_organisation_member
-- request_campaign_contestants DML
-- File: acc_requests.sql
-- ==============================================================

-- ========================================
-- request_organisation_member
-- ========================================

-- name: CreateRequestOrganisationMember :one
INSERT INTO request_organisation_member (
    recipient_user_id, from_user_id, org_id
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: ResolveRequestOrganisationMember :one
UPDATE request_organisation_member
SET 
    verdict = $2,
    verdict_timestamp = $3
WHERE request_id = $1
RETURNING *;

-- name: GetOrganisationMemberRequestByID :one
SELECT * FROM request_organisation_member WHERE request_id = $1 LIMIT 1;

-- name: ListPendingRequestsByRecipient :many
SELECT * FROM request_organisation_member WHERE recipient_user_id = $1 AND verdict = $2;

-- name: ListPendingRequestsByOrganisation :many
SELECT * FROM request_organisation_member WHERE org_id = $1 AND verdict = $2;

-- ========================================
-- request_campaign_contestants
-- ========================================

-- name: CreateContestantJoinRequest :one
INSERT INTO request_campaign_contestants (
    request_contestant_user_id, 
    campaign_id,  
    org_id, 
    initiator_position, 
    initiated_by
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: CreateContestantJoinRequestHasRequirement :one
INSERT INTO request_campaign_contestants (
    request_contestant_user_id, 
    campaign_id,  
    org_id, 
    initiator_position, 
    initiated_by
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: ResolveContestantJoinRequest :one
UPDATE request_campaign_contestants
SET 
    verdict = $2,
    verdict_timestamp = $3
WHERE request_id = $1
RETURNING *;

-- name: GetContestantJoinRequestForUser :one
SELECT * FROM request_campaign_contestants WHERE 
request_id = $1 AND 
request_contestant_user_id = $2  LIMIT 1;


-- name: GetContestantJoinRequestForOrganisation :one
SELECT * FROM request_campaign_contestants 
WHERE request_id = $1
AND org_id = $2 LIMIT 1;



-- name: GetContestantJoinRequestByID :one
SELECT * FROM request_campaign_contestants WHERE request_id = $1 LIMIT 1;


-- name: ListPendingContestantRequestsByCampaign :many
SELECT * FROM request_campaign_contestants WHERE campaign_id = $1 
    AND org_id = $2 AND verdict = $3;