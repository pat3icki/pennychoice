CREATE SCHEMA IF NOT EXISTS timescaledb;


-- ======================================================================
-- TIMESCALEDB SCHEMA (Hypertable for Votes)
-- ======================================================================

-- Votes table – each row is a single vote (or vote transaction)
CREATE TABLE timescaledb.votes (
    time            TIMESTAMPTZ NOT NULL,
    campaign_id     UUID NOT NULL,
    contestant_id   UUID NOT NULL,
    user_id         UUID NOT NULL,
    transaction_id  BIGINT,   -- references finance.campaign_transactions.id
    amount          BIGINT NOT NULL,          -- money paid (in smallest unit)
    vote_count      DOUBLE PRECISION NOT NULL, -- amount / campaign.vote_per_one_amount
    currency        VARCHAR(3) NOT NULL DEFAULT 'NGN',

    id              BIGINT GENERATED ALWAYS AS IDENTITY,
    PRIMARY KEY (time, id),

    CONSTRAINT fk_votes_campaign
        FOREIGN KEY (campaign_id)
        REFERENCES campaigns (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_votes_contestant
        FOREIGN KEY (contestant_id)
        REFERENCES contestants (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_votes_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_votes_transaction
        FOREIGN KEY (transaction_id)
        REFERENCES finance.campaign_transactions (id)
        ON DELETE SET NULL
);

COMMENT ON TABLE timescaledb.votes IS 'Time‑series of every vote cast, hypertable partitioned by time.';
COMMENT ON COLUMN timescaledb.votes.amount IS 'Amount paid for this vote (in smallest currency unit).';
COMMENT ON COLUMN timescaledb.votes.vote_count IS 'Number of votes credited to the contestant (computed as amount / campaign.vote_per_one_amount).';

CREATE INDEX idx_votes_campaign ON timescaledb.votes (campaign_id);
CREATE INDEX idx_votes_contestant ON timescaledb.votes (contestant_id);
CREATE INDEX idx_votes_user ON timescaledb.votes (user_id);
CREATE INDEX idx_votes_transaction ON timescaledb.votes (transaction_id);

-- Convert to hypertable (chunked by 7 days)
SELECT create_hypertable('timescaledb.votes', 'time', chunk_time_interval => INTERVAL '7 days');

-- Optional: add a retention policy (e.g., keep 2 years)
SELECT add_retention_policy('timescaledb.votes', INTERVAL '6 months');