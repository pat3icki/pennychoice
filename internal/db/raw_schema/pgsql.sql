CREATE SCHEMA IF NOT EXISTS public;
CREATE SCHEMA IF NOT EXISTS notification;
CREATE SCHEMA IF NOT EXISTS finance;
CREATE SCHEMA IF NOT EXISTS log;
CREATE SCHEMA IF NOT EXISTS admin;


CREATE TYPE identity_verification_status AS ENUM (
    'none',
    'bvn_verified',
    'nin_verified',
    'both_verified'
);

COMMENT ON TYPE identity_verification_status IS 
'Verification status of a user: none (0), bvn_verified (1), nin_verified (2), both_verified (3).';

-- ======================================================================
-- PUBLIC SCHEMA
-- ======================================================================

-- Users
CREATE TABLE IF NOT EXISTS users (
    id                  UUID PRIMARY KEY,
    first_name          TEXT NOT NULL, 
    middle_name         TEXT,
    last_name           TEXT NOT NULL,
    date_of_birth       DATE,
    gender              VARCHAR(1) CHECK (gender IN ('F', 'M')) NOT NULL,
    status              VARCHAR(12) NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'deactivated', 'suspended')),
    email               TEXT NOT NULL UNIQUE,
    phone               VARCHAR(20) UNIQUE,
    country             VARCHAR(55) NOT NULL,
    is_whatsapp_phone   BOOLEAN NOT NULL DEFAULT FALSE,

    -- Verification flags
    is_phone_verified   BOOLEAN DEFAULT FALSE,
    is_email_verified   BOOLEAN DEFAULT FALSE,
    signup_stages       INT NOT NULL DEFAULT 0,

    -- Security
    hash_password       BYTEA NOT NULL,   -- bcrypt
    hash_pin            TEXT,   -- bcrypt

    -- Identity verification
    is_identity_verified identity_verification_status NOT NULL DEFAULT 'none',

    identity_document_type VARCHAR(20) DEFAULT NULL 
        CHECK (identity_document_type IN ('BVN', 'NIN', 'BOTH')),

    encrypted_identity_data TEXT, -- encrypted JSON,

    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deactivated_at      TIMESTAMPTZ DEFAULT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE users IS 'Core user accounts with profile, security, and verification data.';
COMMENT ON COLUMN users.hash_password IS 'Bcrypt hash of the user’s login password.';
COMMENT ON COLUMN users.hash_pin IS 'Bcrypt hash of the 4‑digit transaction PIN.';
COMMENT ON COLUMN users.encrypted_identity_data IS 'JSON Format that shows BVN and NIN encrypted with XChaCha20‑Poly1305 or AES using (a key derived from the user’s PIN.) OR Master Key';
COMMENT ON COLUMN users.signup_stages IS 'Bitwise representation of completed signup stages.';

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users (phone);
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);

-- User Contestant Profiles 
CREATE TABLE IF NOT EXISTS user_contestant_profiles (
    id                      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id                 UUID NOT NULL,
    profile_name            VARCHAR(100) NOT NULL,
    display_name            TEXT NOT NULL,
    bio                     TEXT,
    photo_root              TEXT,
    cover_photo             TEXT,
    social_links            JSONB,
    website_url             TEXT,
    location                TEXT,
    organisation_name       TEXT,
    category                VARCHAR(50),
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    profile_completed_at    TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uk_user_contestant_profiles_user_name 
        UNIQUE (user_id, profile_name),
    
    CONSTRAINT fk_user_contestant_profiles_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

COMMENT ON TABLE user_contestant_profiles IS 'Multiple contestant profiles per user, allowing different personas or categories.';
COMMENT ON COLUMN user_contestant_profiles.social_links IS 'JSON structure: {"twitter": "url", "instagram": "url", ...}';

CREATE INDEX IF NOT EXISTS idx_user_contestant_profiles_user ON user_contestant_profiles (user_id);
CREATE INDEX IF NOT EXISTS idx_user_contestant_profiles_active ON user_contestant_profiles (is_active);
CREATE INDEX IF NOT EXISTS idx_user_contestant_profiles_category ON user_contestant_profiles (category);
CREATE INDEX IF NOT EXISTS idx_user_contestant_profiles_display_name ON user_contestant_profiles (display_name);

-- Organisations
CREATE TABLE IF NOT EXISTS organisations (
    id                    UUID PRIMARY KEY,
    name                  VARCHAR(255) NOT NULL UNIQUE,
    description           TEXT,
    creator_user_id       UUID NOT NULL,
    balance               BIGINT NOT NULL DEFAULT 0,
    max_co_organisers     SMALLINT NOT NULL DEFAULT 10,
    max_active_campaign   SMALLINT NOT NULL DEFAULT 12,
    total_members         SMALLINT NOT NULL DEFAULT 0,
    total_campaign        SMALLINT NOT NULL DEFAULT 0,
    active_campaign       SMALLINT NOT NULL DEFAULT 0
                          CHECK (active_campaign <= total_campaign),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    total_contributions   BIGINT NOT NULL DEFAULT 0,
    permissions           JSONB,

    CONSTRAINT fk_org_creator
        FOREIGN KEY (creator_user_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
);

COMMENT ON TABLE organisations IS 'Organisations that host campaigns and manage members.';

CREATE INDEX IF NOT EXISTS idx_org_creator ON organisations (creator_user_id);
CREATE INDEX IF NOT EXISTS idx_org_name ON organisations (name);

-- Organisation Members
CREATE TABLE IF NOT EXISTS organisations_members (
    org_id      UUID NOT NULL,
    member_id   UUID NOT NULL,
    position    INT NOT NULL,
    invited_by  UUID,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (org_id, member_id),
    UNIQUE (org_id, position),

    CONSTRAINT fk_org_member_org
        FOREIGN KEY (org_id)
        REFERENCES organisations (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_org_member_user
        FOREIGN KEY (member_id)
        REFERENCES users (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_org_member_inviter
        FOREIGN KEY (invited_by)
        REFERENCES users (id)
        ON DELETE SET NULL
);

COMMENT ON TABLE organisations_members IS 'Membership of users in organisations; position determines roles (e.g., 1=admin).';

CREATE INDEX IF NOT EXISTS idx_org_members_member ON organisations_members (member_id);
CREATE INDEX IF NOT EXISTS idx_org_members_invited_by ON organisations_members (invited_by);

-- Campaigns
CREATE TABLE IF NOT EXISTS campaigns (
    id                        UUID PRIMARY KEY,
    organisation_id           UUID NOT NULL,
    creator_user_id           UUID NOT NULL,
    name                      TEXT NOT NULL ,
    description               TEXT,
    status                    VARCHAR(12) NOT NULL DEFAULT 'draft'
                              CHECK (status IN ('draft', 'active', 'paused', 'ended', 'cancelled')),
    bank_name_suffix          VARCHAR(12) NOT NULL,
    tags                      TEXT,
    amount_raised             BIGINT NOT NULL DEFAULT 0,
    total_votes               DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    total_contestants         SMALLINT NOT NULL DEFAULT 0,
    vote_currency             VARCHAR(3) NOT NULL DEFAULT 'NGN',
    vote_per_one_amount       INT NOT NULL,
    vote_metrics_visibility   VARCHAR(20) NOT NULL DEFAULT 'private'
                              CHECK (vote_metrics_visibility IN ('private', 'contestants', 'public')),
    registration_fee          BIGINT NOT NULL DEFAULT 0,
    contestant_requirements   TEXT,
    allow_refunds             BOOLEAN NOT NULL DEFAULT FALSE,
    verdict_partial_vote      BOOLEAN NOT NULL DEFAULT FALSE,
    platform_fee_percentage   SMALLINT NOT NULL CHECK (platform_fee_percentage BETWEEN 0 AND 100),
    activation_method         VARCHAR(12) NOT NULL DEFAULT 'schedule'   
                              CHECK (activation_method IN ('schedule', 'manual')),  
    votes_starts_at           TIMESTAMPTZ NOT NULL,
    votes_ends_at             TIMESTAMPTZ NOT NULL,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(organisation_id, name),

    CONSTRAINT fk_campaign_org
        FOREIGN KEY (organisation_id)
        REFERENCES organisations (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_campaign_creator
        FOREIGN KEY (creator_user_id)
        REFERENCES users (id)
        ON DELETE RESTRICT,
    CHECK (votes_starts_at < votes_ends_at)
);

COMMENT ON TABLE campaigns IS 'Charity election campaigns, each linked to an organisation.';

CREATE INDEX IF NOT EXISTS idx_campaign_org ON campaigns (organisation_id);
CREATE INDEX IF NOT EXISTS idx_campaign_creator ON campaigns (creator_user_id);
CREATE INDEX IF NOT EXISTS idx_campaign_status ON campaigns (status);
CREATE INDEX IF NOT EXISTS idx_campaign_dates ON campaigns (votes_starts_at, votes_ends_at);

-- Contestants
CREATE TABLE IF NOT EXISTS contestants (
    id                         UUID PRIMARY KEY,
    campaign_id                UUID NOT NULL,
    user_id                    UUID NOT NULL,
    display_name               VARCHAR(255) NOT NULL,
    photo_root                 TEXT,
    social_links               JSONB,
    organisation_name          TEXT,
    checkout                   VARCHAR(55),
    is_reserved_account_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    vote_count                 DOUBLE PRECISION NOT NULL DEFAULT 0,
    hv_per_transact            DOUBLE PRECISION NOT NULL DEFAULT 0,

    UNIQUE (campaign_id, user_id),

    CONSTRAINT fk_contestant_campaign
        FOREIGN KEY (campaign_id)
        REFERENCES campaigns (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_contestant_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
);

COMMENT ON TABLE contestants IS 'Participants in a campaign; users can vote for them.';

CREATE INDEX IF NOT EXISTS idx_contestant_campaign ON contestants (campaign_id);
CREATE INDEX IF NOT EXISTS idx_contestant_user ON contestants (user_id);

-- Organisation Decisions
CREATE TABLE IF NOT EXISTS organisation_decisions (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organisation_id    UUID NOT NULL,
    campaign_id        UUID,
    action_type        VARCHAR(50) NOT NULL
                       CHECK (action_type IN (
                           'withdraw_funds', 'invite_member', 'remove_member',
                           'create_campaign', 'update_campaign', 'delete_campaign'
                       )),
    message            TEXT NOT NULL,
    status             VARCHAR(12) NOT NULL DEFAULT 'pending'
                       CHECK (status IN ('pending', 'in-execution', 'concluded')),
    admin_initiator_id UUID NOT NULL,
    deadline           TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_decisions_org
        FOREIGN KEY (organisation_id)
        REFERENCES organisations (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_decisions_campaign
        FOREIGN KEY (campaign_id)
        REFERENCES campaigns (id)
        ON DELETE SET NULL,
    CONSTRAINT fk_decisions_initiator
        FOREIGN KEY (admin_initiator_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
);

COMMENT ON TABLE organisation_decisions IS 'Governance actions (votes/decisions) taken by organisation admins.';

CREATE INDEX IF NOT EXISTS idx_decisions_org ON organisation_decisions (organisation_id);
CREATE INDEX IF NOT EXISTS idx_decisions_campaign ON organisation_decisions (campaign_id);
CREATE INDEX IF NOT EXISTS idx_decisions_initiator ON organisation_decisions (admin_initiator_id);

-- Organisation Decisions Verdict
CREATE TABLE IF NOT EXISTS organisation_decisions_verdict (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    decision_id        BIGINT NOT NULL,
    organisation_id    UUID NOT NULL,
    user_id            UUID NOT NULL,
    verdict            VARCHAR(16) NOT NULL CHECK (verdict IN ('support', 'against')),
    time_stamp         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_verdict_decision FOREIGN KEY (decision_id) REFERENCES organisation_decisions (id) ON DELETE CASCADE,
    CONSTRAINT fk_verdict_org FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE,
    CONSTRAINT fk_verdict_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_decisions_verdict_decision ON organisation_decisions_verdict (decision_id);
CREATE INDEX IF NOT EXISTS idx_decisions_verdict_org ON organisation_decisions_verdict (organisation_id);
CREATE INDEX IF NOT EXISTS idx_decisions_verdict_user ON organisation_decisions_verdict (user_id);

-- Organisation Join Requests
CREATE TABLE IF NOT EXISTS request_organisation_member (
    request_id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, 
    recipient_user_id  UUID NOT NULL,
    from_user_id       UUID NOT NULL,
    org_id             UUID NOT NULL, 
    initiator_position         VARCHAR(32) NOT NULL
                            CHECK (initiator_position IN ('organisation_member_user','user')),
    verdict            VARCHAR(16) NOT NULL DEFAULT 'pending'
                       CHECK (verdict IN ('pending', 'accept', 'decline')),
    verdict_timestamp  TIMESTAMPTZ,
    invited_timestamp  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_request_recipient
        FOREIGN KEY (recipient_user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_request_from
        FOREIGN KEY (from_user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_request_org
        FOREIGN KEY (org_id)
        REFERENCES organisations (id)
        ON DELETE CASCADE
);

COMMENT ON TABLE request_organisation_member IS 'Pending invitations or requests to join an organisation.';

CREATE INDEX IF NOT EXISTS idx_request_org ON request_organisation_member (org_id);
CREATE INDEX IF NOT EXISTS idx_request_recipient ON request_organisation_member (recipient_user_id);
CREATE INDEX IF NOT EXISTS idx_request_from_user ON request_organisation_member (from_user_id);
CREATE INDEX IF NOT EXISTS idx_request_verdict ON request_organisation_member (verdict);

-- Campaign Join Request Contestants
CREATE TABLE IF NOT EXISTS request_campaign_contestants (
    request_id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    request_contestant_user_id UUID NOT NULL, 
    campaign_id                UUID NOT NULL, 
    org_id                     UUID NOT NULL,
    verdict                    VARCHAR(16) NOT NULL DEFAULT 'pending'
                               CHECK (verdict IN ('pending', 'accept', 'decline')), 
    verdict_timestamp          TIMESTAMPTZ,
    initiator_position         VARCHAR(32) NOT NULL
                               CHECK (initiator_position IN ('organisation_member_user','user')),
    payment_status             VARCHAR(16) NOT NULL DEFAULT 'pending'
                               CHECK (payment_status IN ('pending', 'success', 'decline')),
    requirement_data           BYTEA,
    payment_reference          VARCHAR(255),
    initiated_by               UUID NOT NULL,
    invited_timestamp          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_req_contestant_user
        FOREIGN KEY (request_contestant_user_id) 
        REFERENCES users (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_req_contestant_campaign
        FOREIGN KEY (campaign_id)
        REFERENCES campaigns (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_req_contestant_org
        FOREIGN KEY (org_id)
        REFERENCES organisations (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_req_contestant_initiated_by
        FOREIGN KEY (initiated_by)
        REFERENCES users (id)
        ON DELETE CASCADE,
        
    CONSTRAINT chk_payment_ref CHECK (payment_status = 'pending' OR payment_reference IS NOT NULL)
);

COMMENT ON TABLE request_campaign_contestants IS 'Pending invitations or requests for a user to join a campaign as a contestant.';

CREATE INDEX IF NOT EXISTS idx_req_contestant_user ON request_campaign_contestants (request_contestant_user_id);
CREATE INDEX IF NOT EXISTS idx_req_contestant_campaign ON request_campaign_contestants (campaign_id);
CREATE INDEX IF NOT EXISTS idx_req_contestant_org ON request_campaign_contestants (org_id);
CREATE INDEX IF NOT EXISTS idx_req_contestant_initiated ON request_campaign_contestants (initiated_by);
CREATE INDEX IF NOT EXISTS idx_req_contestant_verdict ON request_campaign_contestants (verdict);

-- Notifications
CREATE TABLE IF NOT EXISTS notification.users (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id          UUID NOT NULL,
    title            VARCHAR(255) NOT NULL,
    description      TEXT NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'unread'
                     CHECK (status IN ('unread', 'read', 'archived')),
    action_url       TEXT,
    source           VARCHAR(50) NOT NULL
                     CHECK (source IN ('organisation', 'campaign', 'system')),
    source_target_id TEXT,
    read_at          TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_notification_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

COMMENT ON TABLE notification.users IS 'In‑app notifications for users.';

CREATE INDEX IF NOT EXISTS idx_notifications_users ON notification.users (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notification.users (status);
CREATE INDEX IF NOT EXISTS idx_notifications_source_target ON notification.users (source_target_id);

--  The Organisation Notification Table
CREATE TABLE IF NOT EXISTS notification.organisations (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organisation_id  UUID NOT NULL,
    title            VARCHAR(255) NOT NULL,
    description      TEXT NOT NULL,
    action_url       TEXT,     -- Optional: Deep link to the event (e.g., /org/123/requests)
    
    -- Global status: 'active' means it requires attention, 'archived' means the team resolved it.
    status           VARCHAR(20) NOT NULL DEFAULT 'active'
                     CHECK (status IN ('active', 'archived')),
    
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_org_notification_org
        FOREIGN KEY (organisation_id)
        REFERENCES organisations (id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_org_notifications_org ON notification.organisations (organisation_id);
CREATE INDEX IF NOT EXISTS idx_org_notifications_status ON notification.organisations (status);

-- The Normalized Read Receipts Table
CREATE TABLE IF NOT EXISTS notification.organisations_reads (
    notification_id  BIGINT NOT NULL,
    user_id          UUID NOT NULL,
    read_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Composite Primary Key ensures a user can only have one read receipt per notification
    PRIMARY KEY (notification_id, user_id),

    CONSTRAINT fk_org_notif_read_notif
        FOREIGN KEY (notification_id)
        REFERENCES notification.organisations (id)
        ON DELETE CASCADE,
    
    CONSTRAINT fk_org_notif_read_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

-- Index to quickly find all notifications a specific user has read
CREATE INDEX IF NOT EXISTS idx_org_notif_reads_user ON notification.organisations_reads (user_id);

-- ======================================================================
-- FINANCE SCHEMA
-- ======================================================================

-- Campaign Finalised Transactions
CREATE TABLE IF NOT EXISTS finance.campaign_transactions (
    id                              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    campaign_id                     UUID NOT NULL,
    contestant_id                   UUID,
    transaction_reference           TEXT NOT NULL UNIQUE,
    payment_method                  VARCHAR(20) NOT NULL
                                    CHECK (payment_method IN ('virtual_account', 'checkout')),
    nuban_from_account_number       VARCHAR(11) NOT NULL,
    nuban_from_bank_code            VARCHAR(10) NOT NULL,
    nuban_from_account_name         VARCHAR(255) NOT NULL,
    nuban_to_account_number         VARCHAR(11) NOT NULL,
    nuban_to_bank_code              VARCHAR(10) NOT NULL,
    nuban_to_account_name           VARCHAR(255) NOT NULL,
    payment_provider                VARCHAR(22),
    payment_service_provider        VARCHAR(20) NOT NULL DEFAULT 'monnify',
    payment_service_provider_payload JSONB,
    status                          VARCHAR(12) NOT NULL
                                    CHECK (status IN ('success', 'failed', 'refunded')),
    campaign_balance_before         BIGINT NOT NULL,
    campaign_balance_after          BIGINT NOT NULL,
    amount                          BIGINT NOT NULL,
    currency                        VARCHAR(3) NOT NULL DEFAULT 'NGN',
    entry_type                      VARCHAR(6) NOT NULL CHECK (entry_type IN ('CREDIT', 'DEBIT')),
    purpose                         VARCHAR(50) NOT NULL
                                    CHECK (purpose IN ('VOTE', 'REFUND', 'REGISTRATION_FEE', 'REFUND_REG_FEE')),
    paid_at                         TIMESTAMPTZ NOT NULL,
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_trans_campaign
        FOREIGN KEY (campaign_id)
        REFERENCES campaigns (id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_trans_contestant
        FOREIGN KEY (contestant_id)
        REFERENCES contestants (id)
        ON DELETE SET NULL
);

COMMENT ON TABLE finance.campaign_transactions IS 'Finalised campaign‑level financial movements.';

CREATE INDEX IF NOT EXISTS idx_trans_campaign ON finance.campaign_transactions (campaign_id);
CREATE INDEX IF NOT EXISTS idx_trans_contestant ON finance.campaign_transactions (contestant_id);
CREATE INDEX IF NOT EXISTS idx_trans_reference ON finance.campaign_transactions (transaction_reference);
CREATE INDEX IF NOT EXISTS idx_trans_paid_at ON finance.campaign_transactions (paid_at);
CREATE INDEX IF NOT EXISTS idx_trans_status ON finance.campaign_transactions (status);

-- Organisation Pending Transactions
CREATE TABLE IF NOT EXISTS finance.organisation_pending_transactions (
    id                              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organisation_id                 UUID NOT NULL,
    transaction_reference           TEXT NOT NULL UNIQUE,
    entry_type                      VARCHAR(6) NOT NULL CHECK (entry_type IN ('CREDIT', 'DEBIT')),
    status                          VARCHAR(12) NOT NULL DEFAULT 'pending'
                                    CHECK (status IN ('pending', 'expired')),
    currency                        VARCHAR(3) NOT NULL DEFAULT 'NGN',
    amount                          BIGINT NOT NULL,
    nuban_target_account_number     VARCHAR(11) NOT NULL,
    nuban_target_bank_code          VARCHAR(10) NOT NULL,
    nuban_target_account_name       VARCHAR(255) NOT NULL,
    purpose                         VARCHAR(50) NOT NULL,
    request_id                      UUID,
    paid_at                         TIMESTAMPTZ,
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_org_pending_org
        FOREIGN KEY (organisation_id)
        REFERENCES organisations (id)
        ON DELETE RESTRICT
);

COMMENT ON TABLE finance.organisation_pending_transactions IS 'Pending organisation‑level financial operations.';

CREATE INDEX IF NOT EXISTS idx_org_pending_ref ON finance.organisation_pending_transactions (transaction_reference);
CREATE INDEX IF NOT EXISTS idx_org_pending_org ON finance.organisation_pending_transactions (organisation_id);



-- Campaign Pending Transactions
CREATE TABLE  IF NOT EXISTS finance.campaign_invoices (
    id                               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    campaign_id                      UUID NOT NULL,
    contestant_id                    UUID,
    transaction_reference            TEXT NOT NULL UNIQUE,
    payment_method                   VARCHAR(20) NOT NULL
                                     CHECK (payment_method IN ('virtual_account', 'checkout')),

    payment_service_provider         VARCHAR(20) NOT NULL DEFAULT 'monnify',
    payment_service_provider_payload JSONB,
    amount                           BIGINT NOT NULL,
    currency                         VARCHAR(3) NOT NULL DEFAULT 'NGN',

    -- New invoice‑specific fields
    invoice_issue_date               TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    invoice_due_date                 TIMESTAMPTZ,
    invoice_description              TEXT,
    payment_terms                    TEXT,
    invoice_status                   VARCHAR(20) NOT NULL DEFAULT 'draft'
                                     CHECK (invoice_status IN ('draft', 'issued', 'paid', 'expired', 'cancelled')),

    -- Legacy field kept for traceability
    paid_at                          TIMESTAMPTZ,
    created_at                       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_campaign_invoice_campaign
        FOREIGN KEY (campaign_id)
        REFERENCES campaigns (id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_campaign_invoice_contestant
        FOREIGN KEY (contestant_id)
        REFERENCES contestants (id)
        ON DELETE SET NULL
);

COMMENT ON TABLE finance.campaign_invoices IS 'Invoices for campaign‑related financial operations (votes, registration fees, refunds).';
COMMENT ON COLUMN finance.campaign_invoices.transaction_reference IS 'Unique reference number acting as invoice number.';
COMMENT ON COLUMN finance.campaign_invoices.invoice_status IS 'Current lifecycle status of the invoice: draft → issued → paid (or expired/cancelled).';

CREATE INDEX IF NOT EXISTS idx_campaign_invoice_ref ON finance.campaign_invoices (transaction_reference);
CREATE INDEX IF NOT EXISTS idx_campaign_invoice_campaign ON finance.campaign_invoices (campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_invoice_contestant ON finance.campaign_invoices (contestant_id);
CREATE INDEX IF NOT EXISTS idx_campaign_invoice_status ON finance.campaign_invoices (invoice_status);
CREATE INDEX IF NOT EXISTS idx_campaign_invoice_due_date ON finance.campaign_invoices (invoice_due_date);

-- Organisation Finalised Transactions
CREATE TABLE IF NOT EXISTS finance.organisation_invoices (
    id                               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organisation_id                  UUID NOT NULL,
    transaction_reference            TEXT NOT NULL UNIQUE,
    entry_type                       VARCHAR(6) NOT NULL CHECK (entry_type IN ('CREDIT', 'DEBIT')),
    currency                         VARCHAR(3) NOT NULL DEFAULT 'NGN',
    amount                           BIGINT NOT NULL,
    nuban_target_account_number      VARCHAR(11) NOT NULL,
    nuban_target_bank_code           VARCHAR(10) NOT NULL,
    nuban_target_account_name        VARCHAR(255) NOT NULL,
    purpose                          VARCHAR(50) NOT NULL,
    request_id                       UUID,

    -- New invoice‑specific fields
    invoice_issue_date               TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    invoice_due_date                 TIMESTAMPTZ,
    invoice_description              TEXT,
    payment_terms                    TEXT,
    invoice_status                   VARCHAR(20) NOT NULL DEFAULT 'draft'
                                     CHECK (invoice_status IN ('draft', 'issued', 'paid', 'expired', 'cancelled')),

    -- Legacy field kept for traceability
    paid_at                          TIMESTAMPTZ,
    created_at                       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_org_invoice_org
        FOREIGN KEY (organisation_id)
        REFERENCES organisations (id)
        ON DELETE RESTRICT
);

COMMENT ON TABLE finance.organisation_invoices IS 'Invoices for organisation‑level financial operations, replacing pending transactions.';
COMMENT ON COLUMN finance.organisation_invoices.transaction_reference IS 'Unique reference number acting as invoice number.';
COMMENT ON COLUMN finance.organisation_invoices.invoice_status IS 'Current lifecycle status of the invoice: draft → issued → paid (or expired/cancelled).';

CREATE INDEX IF NOT EXISTS idx_organisation_invoice_ref ON finance.organisation_invoices (transaction_reference);
CREATE INDEX IF NOT EXISTS idx_organisation_invoice_org ON finance.organisation_invoices (organisation_id);
CREATE INDEX IF NOT EXISTS idx_organisation_invoice_status ON finance.organisation_invoices (invoice_status);
CREATE INDEX IF NOT EXISTS idx_organisation_invoice_due_date ON finance.organisation_invoices (invoice_due_date);




-- Contestant Reserved Accounts
CREATE TABLE IF NOT EXISTS finance.contestant_reserved_account (
    contestant_id           UUID PRIMARY KEY,
    account_reference       VARCHAR(255),
    account_name            TEXT NOT NULL,
    account_number          VARCHAR(10) NOT NULL,
    currency_code           VARCHAR(3),
    payment_service_payload JSONB,
    created_on              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_reserved_contestant
        FOREIGN KEY (contestant_id)
        REFERENCES contestants (id)
        ON DELETE CASCADE
);

COMMENT ON TABLE finance.contestant_reserved_account IS 'Details of reserved bank accounts created for contestants.';

-- Webhook Incoming Log
CREATE TABLE IF NOT EXISTS finance.webhook (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    event_type   VARCHAR(50) NOT NULL,
    payload      JSONB NOT NULL
);

COMMENT ON TABLE finance.webhook IS 'Raw incoming webhook events from payment providers.';

CREATE INDEX IF NOT EXISTS idx_webhook_event ON finance.webhook (event_type);
CREATE INDEX IF NOT EXISTS idx_webhook_received ON finance.webhook (received_at);

-- ======================================================================
-- LOG SCHEMA
-- ======================================================================

CREATE TABLE IF NOT EXISTS log.unified_audit (
    id              BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL UNIQUE,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actor_type      VARCHAR(20) NOT NULL CHECK (actor_type IN ('SYSTEM', 'USER', 'ADMIN', 'CRON_JOB')),
    actor_id        UUID,       
    action_type     VARCHAR(50) NOT NULL,  
    table_name      VARCHAR(255),                  
    affected_row_id VARCHAR(255),                  
    old_data        JSONB,
    new_data        JSONB,
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    reason          TEXT,
    metadata        JSONB,

    PRIMARY KEY (id, occurred_at)
);

COMMENT ON TABLE log.unified_audit IS 'Unified database & administrative audit logs tracking actors, actions, and altered states.';

CREATE INDEX IF NOT EXISTS idx_unified_audit_actor ON log.unified_audit (actor_type, actor_id);
CREATE INDEX IF NOT EXISTS idx_unified_audit_occurred ON log.unified_audit (occurred_at);

-- ======================================================================
-- ADMIN SCHEMA
-- ======================================================================

-- Admin Users
CREATE TABLE IF NOT EXISTS admin.users (
    id            UUID PRIMARY KEY,
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    username      VARCHAR(30) UNIQUE,
    role          VARCHAR(20) NOT NULL DEFAULT 'support'
                  CHECK (role IN ('superadmin', 'admin', 'moderator', 'support')),
    status        VARCHAR(12) NOT NULL DEFAULT 'active'
                  CHECK (status IN ('active', 'suspended', 'deactivated')),
    hash_password TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE admin.users IS 'Platform administrative users with varying levels of access.';

-- Admin Settings
CREATE TABLE IF NOT EXISTS admin.settings (
    key           VARCHAR(100) PRIMARY KEY,
    value         TEXT NOT NULL,
    description   TEXT,
    updated_by    UUID,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_settings_user
        FOREIGN KEY (updated_by)
        REFERENCES admin.users (id)
        ON DELETE SET NULL
);

COMMENT ON TABLE admin.settings IS 'Global system settings and feature toggles managed by admins.';

CREATE INDEX IF NOT EXISTS idx_admin_settings_user ON admin.settings (updated_by);



CREATE TABLE IF NOT EXISTS admin.constants (
    key VARCHAR(255) NOT NULL PRIMARY KEY,
    value TEXT,
    description TEXT
);


-- ======================================================================
-- IMMUTABILITY TRIGGERS
-- ======================================================================

CREATE OR REPLACE FUNCTION finance.prevent_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Update is not allowed on table %. This table is immutable.', TG_TABLE_NAME;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION finance.prevent_update() IS 'Trigger function that raises an exception on any UPDATE.';

-- Apply to campaign_transactions
CREATE TRIGGER trg_campaign_transactions_immutable
    BEFORE UPDATE ON finance.campaign_transactions
    FOR EACH ROW
    EXECUTE FUNCTION finance.prevent_update();

-- Apply to organisation_invoices (corrected table name)
CREATE TRIGGER trg_organisation_invoices_immutable
    BEFORE UPDATE ON finance.organisation_invoices
    FOR EACH ROW
    EXECUTE FUNCTION finance.prevent_update();