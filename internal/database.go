package internal

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	OutboxEventStatusPending   = "pending"
	OutboxEventStatusPublished = "published"
	OutboxEventStatusFailed    = "failed"
)

var ErrEventNotFound = fmt.Errorf("eventStore not found")

type IDatabase interface {
	WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error
	GetEvents(ctx context.Context, tx pgx.Tx) ([]*Event, error)
	UpdateEventsStatusPublished(ctx context.Context, tx pgx.Tx, ids []string) error
	UpdateEventsStatusFailed(ctx context.Context, tx pgx.Tx, ids []string) error
}

type Database struct {
	database *pgxpool.Pool
	config   *Config
}

func NewDatabase(database *pgxpool.Pool, config *Config) *Database {
	return &Database{
		database: database,
		config:   config,
	}
}

func (d *Database) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := d.database.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin error: %w", err)
	}

	err = fn(tx)

	if err != nil {
		rbErr := tx.Rollback(ctx)
		if rbErr != nil {
			return fmt.Errorf("rollback error: %v, transaction error: %w", rbErr, err)
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit error: %w", err)
	}

	return nil
}

func (d *Database) GetEvents(ctx context.Context, tx pgx.Tx) ([]*Event, error) {
	query := fmt.Sprintf(`
	select 
    	id, 
    	version, 
    	aggregate_type, 
    	event_type, 
    	content, 
    	status,
		published_at,
		created_at
	from %s where published_at is null limit $1 for update skip locked`, d.config.OutboxTable)

	var events []*Event

	eventRows, err := tx.Query(ctx, query, d.config.BatchSize)
	if err != nil {
		if pgxscan.NotFound(eventRows.Err()) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	err = pgxscan.ScanAll(&events, eventRows)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (d *Database) UpdateEventsStatusPublished(ctx context.Context, tx pgx.Tx, ids []string) error {
	query := fmt.Sprintf(`update %s set status = $1, published_at = now() where id = any($2)`, d.config.OutboxTable)

	_, err := tx.Exec(ctx, query, OutboxEventStatusPublished, ids)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) UpdateEventsStatusFailed(ctx context.Context, tx pgx.Tx, ids []string) error {
	query := fmt.Sprintf(`update %s set status = $1 where id = any($2)`, d.config.OutboxTable)

	_, err := tx.Exec(ctx, query, OutboxEventStatusFailed, ids)
	if err != nil {
		return err
	}

	return nil
}
