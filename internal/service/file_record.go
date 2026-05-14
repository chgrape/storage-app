package service

import (
	"context"
	"io"
	"time"

	"github.com/chgrape/storage-app/internal/repository"
	"github.com/chgrape/storage-app/internal/storage"
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

func (s *FileRecordSvc) ListRecords(ctx context.Context) ([]repository.FileRecord, error) {
	records, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *FileRecordSvc) Erase(ctx context.Context, id uuid.UUID) error {
	rec, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = s.store.Delete(ctx, rec.Path)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileRecordSvc) Download(ctx context.Context, id uuid.UUID) (repository.FileRecord, io.ReadCloser, error) {
	record, err := s.repo.Get(ctx, id)
	if err != nil {
		return repository.FileRecord{}, nil, err
	}

	file, err := s.store.Download(ctx, record.Path)
	if err != nil {
		return repository.FileRecord{}, nil, err
	}

	return record, file, nil
}

func (s *FileRecordSvc) Save(ctx context.Context, file io.Reader, metadata repository.Metadata) (*uuid.UUID, error) {
	uuid := uuid.New()
	upl := time.Now()

	record := &repository.FileRecord{
		Filename:   metadata.Filename,
		Size:       metadata.Size,
		MIMEType:   metadata.MimeType,
		ID:         uuid,
		UploadedAt: upl,
	}

	new_file, err := s.store.Save(ctx, file, record)
	if err != nil {
		return nil, err
	}
	defer new_file.Close()

	rec, err := s.repo.Save(ctx, *record)
	if err != nil {
		return nil, err
	}

	return &rec.ID, nil
}
