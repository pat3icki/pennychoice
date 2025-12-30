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
  "encrypted_nin" VARCHAR(255),

    -- Soft delete fields
  "deleted_at" TIMESTAMP DEFAULT NULL,

  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT "unique_email_active" 
    UNIQUE NULLS NOT DISTINCT ("email", "deleted_at"),

  CONSTRAINT "unique_phone_active" 
    UNIQUE NULLS NOT DISTINCT ("phone", "deleted_at")
);



-- CREATE TABLE "accounts"."users" (
--   "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
--   -- Profile
--   "first_name" VARCHAR(255) NOT NULL,
--   "middle_name" VARCHAR(255) DEFAULT NULL,
--   "last_name" VARCHAR(255) NOT NULL,
--   "date_of_birth" DATE,
--   "gender" CHAR(1) CHECK (gender IN ('F', 'M', 'O')),

--   -- Status & Contact
--   "status" VARCHAR(12) CHECK (status IN ('active', 'deleted', 'suspended')) DEFAULT 'active', 
--   "email" VARCHAR(255) NOT NULL,
--   "phone" VARCHAR(20) UNIQUE,
--   "is_whatsapp_phone" BOOLEAN DEFAULT FALSE,
  
--   -- Verification
--   "is_phone_verified" BOOLEAN DEFAULT FALSE,
--   "is_email_verified" BOOLEAN DEFAULT FALSE,
--   "is_nin_verified" BOOLEAN DEFAULT FALSE,
  
--   -- Hashing And Security 
--   "hash_type" VARCHAR(12) NOT NULL, -- e.g., 'argon2id', 'bcrypt'
--   "hash_password" VARCHAR(255) NOT NULL,
--   "hash_pin" VARCHAR(255) NOT NULL, 
--   "hash_table_seq" SMALLINT DEFAULT 0,
--   "encrypted_nin" VARCHAR(255),

--   -- Timestamps
--   "deleted_at" TIMESTAMPTZ DEFAULT NULL,
--   "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

--   -- Ensures only one 'Active' email (where deleted_at is NULL)
--   CONSTRAINT "unique_email_active" 
--   UNIQUE NULLS NOT DISTINCT ("email", "deleted_at")
-- );



-- CREATE TABLE "accounts"."users" (
--   "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
--   -- Profile
--   "first_name" VARCHAR(255) NOT NULL,
--   "middle_name" VARCHAR(255),
--   "last_name" VARCHAR(255) NOT NULL,
--   "date_of_birth" DATE,
--   "gender" CHAR(1) CHECK (gender IN ('F', 'M', 'O')),

--   "status" VARCHAR(12), -- active, delete, suspended 
--   "email" VARCHAR(255) NOT NULL,
--   "phone" VARCHAR(20) UNIQUE,
--   "is_whatsapp_phone" BOOLEAN DEFAULT FALSE,
  
--   -- Verification
--   "is_phone_verified" BOOLEAN DEFAULT FALSE,
--   "is_email_verified" BOOLEAN DEFAULT FALSE,
--   "is_nin_verified" BOOLEAN DEFAULT FALSE,
  
--   -- Hashing And Security 
--   "hash_type" VARCHAR(12) NOT NULL,
--   "hash_password" VARCHAR(255) NOT NULL,
--   "hash_pin" VARCHAR(255) NOT NULL, 
--   "hash_table_seq" SMALLINT DEFAULT 0, -- 0 means the MD5 hash PIN
--     -- encrypted_nin is encrypted using the user pin (4 number) 
--     -- but using the hash_table_seq to generate a good AES 128
--   "encrypted_nin" VARCHAR(255),

--     -- Soft delete fields
--   "deleted_at" TIMESTAMP DEFAULT NULL,

--   "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

--   CONSTRAINT "unique_email_active" 
--   UNIQUE NULLS NOT DISTINCT ("email", "deleted_at")
-- );