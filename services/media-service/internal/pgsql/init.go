package pgsql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host string
	User string
	Pass string
	Port string
	DB   string
}

func Connect(cfg Config) (*pgxpool.Pool, error) {
	ctx := context.Background()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DB)

	pool, err := pgxpool.New(ctx, connStr)

	if err != nil {
		panic("Error initializing postgres connection")
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
