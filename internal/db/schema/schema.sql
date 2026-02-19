CREATE SCHEMA IF NOT EXISTS "accounts";
CREATE SCHEMA IF NOT EXISTS "timescaledb";
CREATE SCHEMA IF NOT EXISTS "finance";

CREATE TABLE "accounts"."users" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- Profile
  "first_name" VARCHAR(255) NOT NULL,
  "middle_name" VARCHAR(255),
  "last_name" VARCHAR(255) NOT NULL,
  "date_of_birth" DATE,
  "gender" CHAR(1) CHECK (gender IN ('F', 'M', 'O')),
  -- active, delete, suspended 
  "status" VARCHAR(12) NOT NULL CHECK (status IN ('active', 'deactivited', 'suspended')) DEFAULT 'active', 
  "email" VARCHAR(255)  NOT NULL,
  "phone" VARCHAR(20),
  "country" VARCHAR(55) NOT NULL,
  "is_whatsapp_phone" BOOLEAN DEFAULT FALSE,
  
  -- Verification
  "is_phone_verified" BOOLEAN NOT NULL DEFAULT FALSE,
  "is_email_verified" BOOLEAN NOT NULL DEFAULT FALSE,
  "is_nin_verified" BOOLEAN NOT NULL DEFAULT FALSE,
  
  -- Hashing And Security 
  "hash_type" VARCHAR(12) NOT NULL,
  "hash_password" VARCHAR(255) NOT NULL,
  "hash_pin" VARCHAR(255) NOT NULL, 
  "hash_table_seq" SMALLINT DEFAULT 0, -- 0 means the MD5 hash PIN
    -- encrypted_nin is encrypted using the user pin (4 number) 
    -- but using the hash_table_seq to generate a good AES 128
  "encrypted_nin" VARCHAR(255) DEFAULT NULL,

    -- Soft delete fields

  "expected_to_delete" TIMESTAMP DEFAULT NULL CHECK ((status = 'deactivited') = ("deleted_at" IS NOT NULL)),

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
  "source" VARCHAR(50) NOT NULL CHECK (source IN ('organisation', 'campaign', 'system')),
  "source_target_id" TEXT,
  "read_at" TIMESTAMP,

  PRIMARY KEY ("user_id", "msg_id"),

  CONSTRAINT fk_notifications_user 
    FOREIGN KEY ("user_id") 
    REFERENCES "accounts"."users" ("id") 
    ON DELETE CASCADE
); 

-- Organisations Table
CREATE TABLE "accounts"."organisations" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" VARCHAR(255) UNIQUE NOT NULL,
  "description" VARCHAR(516),
  "creator_user_id" UUID NOT NULL,
  "max_co_organisers" SMALLINT NOT NULL DEFAULT 10,
  "max_active_campaign" SMALLINT NOT NULL DEFAULT 12,
  "total_members" SMALLINT NOT NULL DEFAULT 0,
  "total_events" SMALLINT NOT NULL DEFAULT 0,
  "active_events" SMALLINT DEFAULT 0 CHECK ("active_events" <= "total_events"),
  "created_at" TIMESTAMP DEFAULT NOW(),
  "total_contributions" BIGINT NOT NULL,
  "permissions" JSONB,

  CONSTRAINT fk_users_org_creator 
      FOREIGN KEY ("creator_user_id") 
      REFERENCES "accounts"."users" ("id")
      ON DELETE CASCADE
);


-- Organisation Decison (votes or decision) Table
CREATE TABLE "accounts"."organisation_decisons" (
  "id" UUID PRIMARY KEY,
  "campaign_id" UUID,
  "action_type" VARCHAR(255), -- withdraw_funds, invite_member, remove_member,
                              -- create, update, delete events 
  "campaign_id" UUID,
  "organisation_id" UUID NOT NULL,
  "message" TEXT NOT NULL,
  "status" VARCHAR(12) NOT NULL, -- create, excution, concluded
  "initator_id" UUID NOT NULL,
  "deadline" TIMESTAMP,
  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Organisation Members Table
CREATE TABLE "accounts"."organisations_members" (
  "org_id" UUID NOT NULL,
  "member_id" UUID NOT NULL,
  "position" INT,
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


CREATE TABLE "accounts"."campaign" (
  "id" uuid PRIMARY KEY,
  "organisation_id" uuid,
  "name" VARCHAR(55),
  "description" text,
  "status" VARCHAR(6),
  "bank_name_suffix" VARCHAR(12),
  "tags" TEXT,
  "amount_raised" BIGINT,
  "total_votes" DOUBLE PRECISION,
  "total_contestants" smallint,
  "vote_currency" VARCHAR(3),
  "vote_per_one_amount" int,
  "vote_metrics_visibility" VARCHAR(20), -- private, contestants, public
  "allow_refunds" BOOLEAN,
  "vedict_partial_vote" BOOLEAN,
  "platform_fee_percentage" numeric(5,2),
  "votes_starts_at" TIMESTAMP,
  "votes_ends_at" TIMESTAMP,
  "creator_user_id" UUID,
  "create_at" TIMESTAMP
);

CREATE TABLE "timescaledb"."campaign_metrics" (
  "time" TIMESTAMP PRIMARY KEY,
  "user_id" UUID,
  "campaign_id" UUID,
  "vote" DOUBLE PRECISION,
  "amount" BIGINT
);

CREATE TABLE "accounts"."contestants" (
  "id" UUID PRIMARY KEY,
  "campaign_id" UUID,
  "user_id" UUID,
  "display_name" VARCHAR(255),
  "photo_root" text,
  "social_links" jsonb,
  "organisation_name" varchar(255),
  "nuban_account_number" varchar(11),
  "nuban_account_name" varchar(11),
  "nuban_bank_code" varchar(10),
  "nuban_hash" varchar(255) UNIQUE,
  "is_nuban_enabled" BOOLEAN,
  "vote_count" DOUBLE PRECISION,
  "hv_per_transact" DOUBLE PRECISION -- Highest Vote Per Transaction
);

CREATE TABLE "finance"."transactions" (
  "id" bigint PRIMARY KEY,
  "event_id" uuid NOT NULL,
  "transaction_refernce" VARCHAR(255) NOT NULL UNIQUE,
  "payment_method" VARCHAR(16) NOT NULL,
  "payment_service_provider" VARCHAR(12) NOT NULL,
  "nuban_dest_account_num" VARCHAR(11),
  "nuban_dest_bank_code" VARCHAR(10),
  "nuban_source_account_number" VARCHAR(11),
  "nuban_source_bank_code" VARCHAR(8),
  "psp_metadata" jsonb,
  "status" VARCHAR(8) NOT NULL,
  "event_balance_before" INT NOT NULL,
  "event_balance_after" INT NOT NULL,
  "amount" BIGINT NOT NULL,
  "currency" VARCHAR(3) NOT NULL,
  "entry_type" VARCHAR NOT NULL, -- DEBIT OR CREDIT
  "purpose" VARCHAR(20) NOT NULL,
  "paid_at" TIMESTAMP NOT NULL,
  "created_at" TIMESTAMP NOT NULL

);
