package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type FileRecord struct {
	ID         uuid.UUID
	Filename   string
	MIMEType   string
	Size       int64
	Path       string
	CreatedAt  time.Time
	UploadedAt time.Time
}

type FileRecordRepo interface {
	Save(ctx context.Context, f FileRecord) error
	Get(ctx context.Context, id uuid.UUID) (FileRecord, error)
	List(ctx context.Context) ([]FileRecord, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
