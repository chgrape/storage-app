package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/chgrape/storage-app/services/media-service/internal/repository"
)

type DiskStore struct {
	uploadDir string
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

func (d DiskStore) Delete(ctx context.Context, path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func (d DiskStore) Download(ctx context.Context, path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (d DiskStore) Save(ctx context.Context, file io.Reader, rec *repository.FileRecord) (*os.File, error) {
	dir := filepath.Join(d.uploadDir, rec.UploadedAt.Format("2006/01"))
	extension := extFromMIME(rec.MIMEType)
	if extension == "" {
		return nil, errors.New("invalid media type")
	}

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	file_name := rec.ID.String() + extension
	rec.Path = filepath.Join(dir, file_name)

	new_file, err := os.Create(rec.Path)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(new_file, file)
	if err != nil {
		return nil, err
	}
	return new_file, nil
}

func NewDiskStore(uploadDir string) Store {
	return DiskStore{
		uploadDir: uploadDir,
	}
}
