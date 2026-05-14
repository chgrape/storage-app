package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Metadata struct {
	Filename string
	Size     int64
	MimeType string
}

type FileRecord struct {
	ID         uuid.UUID
	Filename   string
	MIMEType   string
	Size       int64
	Path       string
	UploadedAt time.Time
}

type FileRecordRepo interface {
	Save(ctx context.Context, f FileRecord) (FileRecord, error)
	Get(ctx context.Context, id uuid.UUID) (FileRecord, error)
	List(ctx context.Context) ([]FileRecord, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
