// Package repository owns all data access. This file builds the GORM/MySQL
// connection and configures the connection pool, which is the app's primary
// backpressure valve under load (see plan §3.3c).
package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mrkaak/restaurant-api/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDB opens a GORM connection to MySQL and configures the pool from cfg.
func NewDB(cfg config.DB, isProd bool) (*gorm.DB, error) {
	logLevel := gormlogger.Info
	if isProd {
		logLevel = gormlogger.Warn
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true, // we manage transactions explicitly in services
		PrepareStmt:            true, // cache prepared statements for hot queries
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db, nil
}

// Ping verifies the DB is reachable within the context deadline.
func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Close releases the underlying connection pool.
func Close(db *gorm.DB, log *slog.Logger) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Error("closing db", slog.Any("err", err))
	}
}

// WaitForDB retries Ping until the DB is up or the context is cancelled.
// Useful at startup when MySQL and the API boot together (e.g. docker-compose).
func WaitForDB(ctx context.Context, db *gorm.DB, log *slog.Logger) error {
	backoff := 500 * time.Millisecond
	for attempt := 1; ; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err := Ping(pingCtx, db)
		cancel()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("db not ready: %w", ctx.Err())
		case <-time.After(backoff):
			log.Warn("waiting for db", slog.Int("attempt", attempt), slog.Any("err", err))
			if backoff < 5*time.Second {
				backoff *= 2
			}
		}
	}
}
