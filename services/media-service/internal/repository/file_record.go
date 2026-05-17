package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type FileRecord struct {
	ID         uuid.UUID `json:"id"`
	Filename   string    `json:"filename"`
	MIMEType   string    `json:"mime_type"`
	Size       int64     `json:"size"`
	Path       string    `json:"path"`
	UserID     uuid.UUID `json:"user_id"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type FileRecordRepo interface {
	Save(ctx context.Context, f FileRecord) (FileRecord, error)
	Get(ctx context.Context, id uuid.UUID, user_id uuid.UUID) (FileRecord, error)
	List(ctx context.Context, user_id uuid.UUID) ([]FileRecord, error)
	Delete(ctx context.Context, id uuid.UUID, user_id uuid.UUID) error
}
