package service

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/chgrape/storage-app/internal/repository"
	"github.com/google/uuid"
)

type FileRecordSvc struct {
	repo repository.FileRecordRepo
}

func NewFileRecordSvc(repo repository.FileRecordRepo) FileRecordSvc {
	return FileRecordSvc{
		repo: repo,
	}
}

var extMap = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
	"video/x-msvideo": ".avi",
	"video/webm":      ".webm",
}

func extFromMIME(mimeType string) string {
	if ext, ok := extMap[mimeType]; ok {
		return ext
	}
	return ""
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

	err = os.Remove(rec.Path)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileRecordSvc) Download(ctx context.Context, id uuid.UUID) (repository.FileRecord, *os.File, error) {
	record, err := s.repo.Get(ctx, id)
	if err != nil {
		return repository.FileRecord{}, nil, err
	}

	file, err := os.Open(record.Path)
	if err != nil {
		return repository.FileRecord{}, nil, err
	}

	return record, file, nil
}

func (s *FileRecordSvc) Save(ctx context.Context, file io.Reader, metadata repository.Metadata) (*uuid.UUID, error) {
	path, err := filepath.Abs("./uploads")
	if err != nil {
		return nil, err
	}

	upl := time.Now()
	dir := filepath.Join(path, upl.Format("2006/01"))

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	uuid := uuid.New()
	mime := extFromMIME(metadata.MimeType)
	if mime == "" {
		return nil, errors.New("invalid media type")
	}
	file_name := uuid.String() + mime

	new_filepath := filepath.Join(dir, file_name)

	new_file, err := os.Create(new_filepath)
	if err != nil {
		return nil, err
	}
	defer new_file.Close()

	_, err = io.Copy(new_file, file)
	if err != nil {
		return nil, err
	}

	rec, err := s.repo.Save(ctx, repository.FileRecord{
		Filename:   metadata.Filename,
		Size:       metadata.Size,
		MIMEType:   metadata.MimeType,
		ID:         uuid,
		UploadedAt: upl,
		Path:       new_filepath,
	})
	if err != nil {
		return nil, err
	}

	return &rec.ID, nil
}
