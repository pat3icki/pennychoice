-- ==============================================================
-- timescaledb.votes DML
-- File: tsd_votes.sql
-- ==============================================================

-- name: CastVote :one
INSERT INTO timescaledb.votes (
    time, campaign_id, contestant_id, user_id, transaction_id, amount, vote_count, currency
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: ListVotesByCampaign :many
SELECT * FROM timescaledb.votes WHERE campaign_id = $1;

-- name: ListVotesByContestant :many
SELECT * FROM timescaledb.votes WHERE contestant_id = $1;

-- name: ListVotesByUser :many
SELECT * FROM timescaledb.votes WHERE user_id = $1;

-- name: GetContestantVoteSum :one
SELECT COALESCE(SUM(vote_count), 0)::FLOAT AS total_votes FROM timescaledb.votes WHERE contestant_id = $1;