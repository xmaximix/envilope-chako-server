package db

import (
	"database/sql"
	"fmt"
	"github.com/xmaximix/envilope-chako-server/internal/config"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgres(cfg config.DBConfig) (*sqlx.DB, error) {
	if url := os.Getenv("DB_URL"); url != "" {
		return sqlx.Connect("pgx", url)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	)
	return sqlx.Connect("pgx", dsn)
}

func MigrateUp(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://scripts/migrations", "postgres", driver,
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
