package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pat3icki/pennychoice/internal/db/cache"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
)

type Client struct {
	_pool        *pgxpool.Pool
	TrueUniqueID bool
	Query        sqlc.Querier
	Cache        cache.Cache
}

type Table string

const (
	TableUsers        Table = "users"
	TableOrganisation Table = "organisations"
	TableCampaigns    Table = "campaigns"
	TableContestant   Table = "contestants"
)

func (c *Client) Exists(id uuid.UUID, table Table) (bool, error) {
	ctx := context.Background()
	var exists bool
	var err error
	if id == uuid.Nil {
		return false, errors.New("parameter `id` cannot be empty or nil")
	}

	switch table {
	case TableUsers:
		exists, err = c.Query.ExistsUser(ctx, id)
	case TableOrganisation:
		exists, err = c.Query.ExistsOrganisation(ctx, id)
	case TableCampaigns:
		exists, err = c.Query.ExistsCampaign(ctx, id)
	case TableContestant:
		exists, err = c.Query.ExistsContestant(ctx, id)
	}
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *Client) WithTx(ctx context.Context, fn func(tx pgx.Tx, q sqlc.Querier) error) error {
	q := &sqlc.Queries{}
	tx, err := c._pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := fn(tx, q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (c *Client) Exec(ctx context.Context, sql string, arguments ...any) error {
	_, err := c._pool.Exec(ctx, sql, arguments...)
	return err
}

func (c *Client) GenerateUUID(ctx context.Context, table Table) (uuid.UUID, error) {
GEN_ID:
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}
	id_exists, _ := c.Exists(id, table)
	if id_exists {
		goto GEN_ID
	}
	return id, nil
}

// Close closes the database connection pool
func (c *Client) Close() {
	if c._pool != nil {
		c._pool.Close()
	}
}
