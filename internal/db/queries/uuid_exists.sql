-- name: ExistsUser :one
SELECT EXISTS(
    SELECT 1 
    FROM users 
    WHERE id = $1
) AS exists;

-- name: ExistsOrganisation :one
SELECT EXISTS(
    SELECT 1 
    FROM organisations 
    WHERE id = $1
) AS exists;

-- name: ExistsCampaign :one
SELECT EXISTS(
    SELECT 1 
    FROM campaigns 
    WHERE id = $1
) AS exists;

-- name: ExistsContestant :one
SELECT EXISTS(
    SELECT 1 
    FROM contestants 
    WHERE id = $1
) AS exists;

-- name: ExistsCampaignContestant :one
SELECT EXISTS (
    SELECT 1
    FROM contestants
    WHERE campaign_id = $1   
      AND user_id = $2        
) AS is_contestant;

-- name: ExistsTransaction :one
SELECT EXISTS(
    SELECT 1 
    FROM finance.campaign_transactions 
    WHERE id = $1
) AS exists;

-- name: ExistsOrganisationMember :one
SELECT EXISTS(
    SELECT 1 
    FROM organisations_members 
    WHERE org_id = $1 AND member_id = $2
) AS exists;

-- name: ExistsPendingTransaction :one
SELECT EXISTS(
    SELECT 1 
    FROM finance.campaign_invoices 
    WHERE transaction_reference = $1
) AS exists;

-- Generic UUID existence check with table name parameter
-- name: ExistsUUIDInTable :one
SELECT EXISTS(
    SELECT 1 
    FROM information_schema.columns 
    WHERE table_schema = $1 
    AND table_name = $2 
    AND column_name = $3
) AND EXISTS(
    SELECT 1 
    FROM (
        SELECT 
            CASE 
                WHEN $1 = 'accounts' AND $2 = 'users' THEN (SELECT 1 FROM users WHERE id = $4::UUID)
                WHEN $1 = 'accounts' AND $2 = 'organisations' THEN (SELECT 1 FROM organisations WHERE id = $4::UUID)
                WHEN $1 = 'accounts' AND $2 = 'campaigns' THEN (SELECT 1 FROM campaigns WHERE id = $4::UUID)
                WHEN $1 = 'accounts' AND $2 = 'contestants' THEN (SELECT 1 FROM contestants WHERE id = $4::UUID)
                WHEN $1 = 'finance' AND $2 = 'campaign_transactions' THEN (SELECT 1 FROM finance.campaign_transactions WHERE id = $4::UUID)
                WHEN $1 = 'admin' AND $2 = 'users' THEN (SELECT 1 FROM admin.users WHERE id = $4::UUID)
                ELSE NULL
            END AS result
    ) t
    WHERE result = 1
) AS exists;