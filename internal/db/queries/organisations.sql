-- Insert into organisations
-- CreateOrganisation
INSERT INTO "accounts"."organisations" (
    "id", 
    "name", 
    "description", 
    "creator_user", 
    "max_co_organisers", 
    "max_active_events"
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;


-- UpdateOrganisation
UPDATE  "accounts"."organisations" 
SET 
    "name" = $2,
    "description" = $3,
    "max_co_organisers" = $4,
    "max_active_events" = $5
WHERE 
    "id" = $1
RETURNING *;



-- Insert into organisation_members
-- OrganisationAddMember
INSERT INTO "accounts"."organisation_members" (
    org_id, 
    position, 
    member, 
    is_admin, 
    inivted_by, 
    date
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;