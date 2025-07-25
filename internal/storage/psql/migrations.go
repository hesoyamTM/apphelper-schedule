package psql

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(ctx context.Context, cfg PsqlConfig) error {
	const op = "migrations"
	log := logger.GetLoggerFromCtx(ctx)

	log.Info(ctx, "running migrations")

	m, err := migrate.New(
		"file://migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info(ctx, "no migrations to run")

			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info(ctx, "migrations completed")
	return nil
}
