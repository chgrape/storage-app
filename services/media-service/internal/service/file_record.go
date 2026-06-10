package service

import (
	"context"
	"io"
	"time"

	"github.com/chgrape/storage-app/services/media-service/internal/repository"
	"github.com/chgrape/storage-app/services/media-service/internal/storage"
	"github.com/chgrape/storage-app/shared"
	"github.com/google/uuid"
)

type FileRecordSvc struct {
	repo  repository.FileRecordRepo
	store storage.Store
}

func NewFileRecordSvc(repo repository.FileRecordRepo, store storage.Store) FileRecordSvc {
	return FileRecordSvc{
		repo:  repo,
		store: store,
	}
}

func (s *FileRecordSvc) ListRecords(ctx context.Context, user_id uuid.UUID) ([]repository.FileRecord, error) {
	records, err := s.repo.List(ctx, user_id)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *FileRecordSvc) Erase(ctx context.Context, id uuid.UUID, user_id uuid.UUID) error {
	rec, err := s.repo.Get(ctx, id, user_id)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, id, user_id)
	if err != nil {
		return err
	}

	err = s.store.Delete(ctx, rec.Path)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileRecordSvc) Download(ctx context.Context, id uuid.UUID, user_id uuid.UUID) (repository.FileRecord, io.ReadCloser, error) {
	record, err := s.repo.Get(ctx, id, user_id)
	if err != nil {
		return repository.FileRecord{}, nil, err
	}

	file, err := s.store.Download(ctx, record.Path)
	if err != nil {
		return repository.FileRecord{}, nil, err
	}

	return record, file, nil
}

func (s *FileRecordSvc) Init(metadata shared.Metadata, user_id uuid.UUID) (io.WriteCloser, *repository.FileRecord, error) {
	rec := &repository.FileRecord{
		Filename:   metadata.Filename,
		Size:       metadata.Size,
		MIMEType:   metadata.MimeType,
		ID:         uuid.New(),
		UserID:     user_id,
		UploadedAt: time.Now(),
	}

	file, err := s.store.CreateFile(rec)
	if err != nil {
		return nil, nil, err
	}
	return file, rec, nil
}

func (s *FileRecordSvc) Write(chunk io.Reader, file io.WriteCloser) error {
	err := s.store.Save(chunk, file)
	if err != nil {
		return err
	}
	return nil
}

func (s *FileRecordSvc) Commit(ctx context.Context, record repository.FileRecord) (*uuid.UUID, error) {
	err := s.repo.Save(ctx, record)
	if err != nil {
		return nil, err
	}
	return &record.ID, nil
}
