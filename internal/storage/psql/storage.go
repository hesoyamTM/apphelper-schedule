package psql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PsqlConfig struct {
	Host     string `env:"PSQL_HOST"`
	Port     int    `env:"PSQL_PORT"`
	User     string `env:"PSQL_USER"`
	Password string `env:"PSQL_PASSWORD"`
	DB       string `env:"PSQL_DB"`
}

type Storage struct {
	db *pgxpool.Pool
}

func New(cfg PsqlConfig) *Storage {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	return &Storage{
		db: pool,
	}
}
