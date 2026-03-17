package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kevalsabhani/go-boilerplate/internal/config"
	"go.uber.org/zap"
)

const databasePingTimeout = 10 * time.Second

type DB struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func New(ctx context.Context, cfg config.DatabaseConfig, logger *zap.Logger) (*DB, error) {
	logger = logger.Named("database")

	poolCfg, err := buildPoolConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("database: build pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}

	db := &DB{
		pool:   pool,
		logger: logger,
	}

	ctx, cancel := context.WithTimeout(ctx, databasePingTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	logger.Info("Database connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("dbname", cfg.DBName),
	)

	return db, nil
}

func (db *DB) Close() {
	db.pool.Close()
	db.logger.Info("Database connection pool closed")
}

// buildPoolConfig builds a pgxpool.Config from the given config.
func buildPoolConfig(cfg config.DatabaseConfig) (*pgxpool.Config, error) {

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MaxConnIdleTime = time.Duration(cfg.ConnMaxIdleTime) * time.Second
	poolCfg.MaxConnLifetime = time.Duration(cfg.ConnMaxLifetime) * time.Second
	poolCfg.ConnConfig.ConnectTimeout = 5 * time.Second

	return poolCfg, nil
}
