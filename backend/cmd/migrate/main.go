// Command migrate applies/rolls back database migrations using the embedded
// SQL files. Cross-platform: no separate golang-migrate CLI install needed.
//
//	go run ./cmd/migrate up          # apply all pending
//	go run ./cmd/migrate down 1      # roll back one step
//	go run ./cmd/migrate version     # print current version
//	go run ./cmd/migrate force 1     # set version without running (recovery)
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"database/sql"
	gomysql "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/mrkaak/restaurant-api/migrations"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		fmt.Println("usage: migrate <up|down|version|force> [n]")
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		fatal("config", err)
	}

	m, closeFn, err := newMigrator(cfg.DB)
	if err != nil {
		fatal("init migrator", err)
	}
	defer closeFn()

	cmd := os.Args[1]
	switch cmd {
	case "up":
		run(m.Up)
	case "down":
		if len(os.Args) >= 3 {
			n := mustAtoi(os.Args[2])
			run(func() error { return m.Steps(-n) })
		} else {
			run(m.Down)
		}
	case "version":
		v, dirty, verr := m.Version()
		if errors.Is(verr, migrate.ErrNilVersion) {
			fmt.Println("no migrations applied yet")
			return
		}
		if verr != nil {
			fatal("version", verr)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
	case "force":
		if len(os.Args) < 3 {
			fatal("force", errors.New("force requires a version number"))
		}
		if err := m.Force(mustAtoi(os.Args[2])); err != nil {
			fatal("force", err)
		}
		fmt.Println("forced version", os.Args[2])
	default:
		fmt.Printf("unknown command %q\n", cmd)
		os.Exit(2)
	}
}

func newMigrator(dbc config.DB) (*migrate.Migrate, func(), error) {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("load embedded migrations: %w", err)
	}

	// multiStatements lets a single migration file run several statements
	// (e.g. CREATE TABLE + INSERT seeds).
	dsn := dbc.DSN()
	cfg, err := gomysql.ParseDSN(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("parse dsn: %w", err)
	}
	cfg.MultiStatements = true

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("mysql driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return nil, nil, fmt.Errorf("new migrate: %w", err)
	}
	return m, func() { _, _ = m.Close() }, nil
}

func run(fn func() error) {
	if err := fn(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no change: database already up to date")
			return
		}
		fatal("migrate", err)
	}
	fmt.Println("migration ok")
}

func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		fatal("parse number", err)
	}
	return n
}

func fatal(stage string, err error) {
	slog.Error("migrate failed", slog.String("stage", stage), slog.Any("err", err))
	os.Exit(1)
}
