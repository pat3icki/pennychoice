CREATE SCHEMA IF NOT EXISTS "accounts";


CREATE TABLE "accounts"."users" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- Profile
  "first_name" VARCHAR(255) NOT NULL,
  "middle_name" VARCHAR(255),
  "last_name" VARCHAR(255) NOT NULL,
  "date_of_birth" DATE,
  "gender" CHAR(1) CHECK (gender IN ('F', 'M', 'O')),
  -- active, delete, suspended 
  "status" VARCHAR(12)  CHECK (status IN ('active', 'deleted', 'suspended')) DEFAULT 'active', 
  "email" VARCHAR(255)  NOT NULL,
  "phone" VARCHAR(20),
  "is_whatsapp_phone" BOOLEAN DEFAULT FALSE,
  
  -- Verification
  "is_phone_verified" BOOLEAN DEFAULT FALSE,
  "is_email_verified" BOOLEAN DEFAULT FALSE,
  "is_nin_verified" BOOLEAN DEFAULT FALSE,
  
  -- Hashing And Security 
  "hash_type" VARCHAR(12) NOT NULL,
  "hash_password" VARCHAR(255) NOT NULL,
  "hash_pin" VARCHAR(255) NOT NULL, 
  "hash_table_seq" SMALLINT DEFAULT 0, -- 0 means the MD5 hash PIN
    -- encrypted_nin is encrypted using the user pin (4 number) 
    -- but using the hash_table_seq to generate a good AES 128
  "encrypted_nin" VARCHAR(255) DEFAULT NULL,

    -- Soft delete fields
    -- Check Constraint - check if 
    -- deleted_at IS NOT NULL 
    -- status EQUALS 'deleted'
    -- AND BOTH ARE TRUE 
  "deleted_at" TIMESTAMP DEFAULT NULL CHECK ((status = 'deleted') = ("deleted_at" IS NOT NULL)),

  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT "unique_email_active" 
    UNIQUE NULLS NOT DISTINCT ("email", "deleted_at"),

  CONSTRAINT "unique_phone_active" 
    UNIQUE NULLS NOT DISTINCT ("phone", "deleted_at")

);

CREATE TABLE "accounts"."notifications" (
  "msg_id" BIGSERIAL PRIMARY KEY,
  "user_id" UUID NOT NULL,
  "title" VARCHAR(255) NOT NULL,
  "description" TEXT NOT NULL,
  "status" VARCHAR(20) DEFAULT 'unread' CHECK (status IN ('unread', 'read', 'archived')),
  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  "source" VARCHAR(50) NOT NULL CHECK (source IN ('organisation', 'event', 'system')),
  "source_target_id" TEXT,
  "read_at" TIMESTAMP,

  CONSTRAINT fk_notifications_user 
    FOREIGN KEY ("user_id") 
    REFERENCES "accounts"."users" ("id") 
    ON DELETE CASCADE
); 

-- Organisations Table
CREATE TABLE "accounts"."organisations" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" VARCHAR(255) UNIQUE,
  "description" VARCHAR(500),
  "creator_user" UUID NOT NULL,
  "max_co_organisers" SMALLINT DEFAULT 10,
  "max_active_events" SMALLINT DEFAULT 12,
  "total_members" SMALLINT DEFAULT 0,
  "total_events" SMALLINT DEFAULT 0,
  "active_events" SMALLINT DEFAULT 0 CHECK ("active_events" <= "total_events"),
  "created_at" TIMESTAMP DEFAULT NOW(),
  "permissions" JSONB,

  CONSTRAINT fk_users_org_creator 
      FOREIGN KEY ("creator_user") 
      REFERENCES "accounts"."users" ("id")
      ON DELETE CASCADE
);


-- Members Table
CREATE TABLE "accounts"."organisations_members" (
  "org_id" UUID NOT NULL,
  "member_id" UUID NOT NULL,
  "position" INT,
  "is_admin" BOOLEAN NOT NULL DEFAULT FALSE,
  "invited_by" UUID,
  "joined_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY ("org_id", "member_id"),
  UNIQUE ("org_id", "position"), -- Position is unique per organization
  
  CONSTRAINT fk_orgs_members_id
    FOREIGN KEY ("org_id") REFERENCES "accounts"."organisations" ("id") ON DELETE CASCADE,
  CONSTRAINT fk_user_members_id
    FOREIGN KEY ("member_id") REFERENCES "accounts"."users" ("id") ON DELETE CASCADE
);


--  Request Table
CREATE TABLE "accounts"."request_organisation_member" (
  "request_id" SERIAL PRIMARY KEY,
  "recipient_user_id" UUID,
  "from_user_id" UUID NOT NULL,
  "from_user_name" VARCHAR(255) NOT NULL,
  "org_id" UUID NOT NULL,
  "org_name" VARCHAR(255) NOT NULL,
  "verdict" VARCHAR(16) DEFAULT 'pending' CHECK (verdict IN ('accept', 'decline', 'pending')),
  "verdict_timestamp" TIMESTAMP,
  "invited_timestamp" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_orgs_request_org 
    FOREIGN KEY ("org_id") REFERENCES "accounts"."organisations" ("id") ON DELETE CASCADE
);