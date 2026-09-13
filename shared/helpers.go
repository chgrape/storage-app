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

func FormatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
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
