package service

import (
	"context"
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
	file_name := uuid.String() + extFromMIME(metadata.MimeType)

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
