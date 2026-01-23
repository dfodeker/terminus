package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	AppRole  string // Optional: role to SET ROLE to after connecting (for privilege separation)
	MaxConns int32
	MinConns int32
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := cfg.URL
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config : %w", err)
	}
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns

	// Only set role if configured (production privilege separation)
	if cfg.AppRole != "" {
		poolCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, "SET ROLE "+cfg.AppRole)
			return err
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return pool, nil
}
