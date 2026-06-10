package shared

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var extMap = map[string]string{
	"image/jpeg":       ".jpg",
	"image/png":        ".png",
	"image/gif":        ".gif",
	"image/webp":       ".webp",
	"video/mp4":        ".mp4",
	"video/quicktime":  ".mov",
	"video/x-msvideo":  ".avi",
	"video/webm":       ".webm",
	"video/x-matroska": ".mkv",
}

func ExtFromMIME(mimeType string) string {
	if ext, ok := extMap[mimeType]; ok {
		return ext
	}
	return ""
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
