package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// wraps database operations with tenant context
type TenantDB struct {
	pool *pgxpool.Pool
}

func NewTenantDB(pool *pgxpool.Pool) *TenantDB {
	return &TenantDB{pool: pool}
}

//with tentnat executes a function with tenatn context set
//this sets the rls variable before executing queries

func (t *TenantDB) WithTenant(ctx context.Context, shopID uuid.UUID, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set tenant context for RLS
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_shop_id = '%s'", shopID.String()))
	if err != nil {
		return fmt.Errorf("set tenant context: %w", err)
	}

	if err := fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// AcquireWithTenant gets a connection with tenant context
// Useful for read operations that don't need transactions
func (t *TenantDB) AcquireWithTenant(ctx context.Context, shopID uuid.UUID) (*pgxpool.Conn, error) {
	conn, err := t.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire conn: %w", err)
	}

	_, err = conn.Exec(ctx, fmt.Sprintf("SET app.current_shop_id = '%s'", shopID.String()))
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("set tenant context: %w", err)
	}

	return conn, nil
}

// QueryWithTenant helper for simple tenant-scoped queries
func (t *TenantDB) QueryWithTenant(ctx context.Context, shopID uuid.UUID, query string, args ...any) (pgx.Rows, error) {
	conn, err := t.AcquireWithTenant(ctx, shopID)
	if err != nil {
		return nil, err
	}
	// Note: caller must release connection after processing rows

	return conn.Query(ctx, query, args...)
}
