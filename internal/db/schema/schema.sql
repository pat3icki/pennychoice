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

CREATE TABLE "accounts"."organisations" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) UNIQUE,
  "description" VARCHAR(500),
  "creator_user" uuid NOT NULL,
  "max_co_organisers" SMALLINT DEFAULT 10,
  "max_active_events" SMALLINT DEFAULT 12,
  "total_members" SMALLINT,
  "total_events" SMALLINT DEFAULT 0,
  "active_events" SMALLINT DEFAULT 0 CHECK ( "active_events" <= "total_events"),
  "created_at" TIMESTAMP DEFAULT N0W(),

  CONSTRAINT fk_users_org_creator 
      FOREIGN KEY ("creator_userid") 
      REFERENCES "accounts"."users" ("id")
      ON DELETE CASCADE
);

CREATE TABLE "accounts"."organisation_members" (
  "org_id" UUID NOT NULL,
  "position" int,
  "member" UUID,
  "is_admin" BOOLEAN NOT NULL,
  "inivted_by" UUID,
  "date" DATETIME,

  PRIMARY KEY (org_id, member),
  UNIQUE (org_id, position),
  
  CONSTRAINT fk_orgs_org_members_id
  FOREIGN KEY ("org_id") 
  REFERENCES "accounts"."organisations" ("id")
  ON DELETE CASCADE
);


