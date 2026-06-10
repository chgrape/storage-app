package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/chgrape/storage-app/services/media-service/internal/repository"
	"github.com/chgrape/storage-app/shared"
)

type DiskStore struct {
	uploadDir string
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

func (d DiskStore) CreateFile(rec *repository.FileRecord) (io.WriteCloser, error) {
	dir := filepath.Join(d.uploadDir, rec.UploadedAt.Format("2006/01"))
	extension := shared.ExtFromMIME(rec.MIMEType)
	if extension == "" {
		return nil, errors.New("invalid media type")
	}

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	file_name := rec.ID.String() + extension
	rec.Path = filepath.Join(dir, file_name)

	new_file, err := os.OpenFile(rec.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return new_file, nil
}

func (d DiskStore) Save(chunk io.Reader, file io.WriteCloser) error {
	_, err := io.Copy(file, chunk)
	if err != nil {
		return err
	}
	return nil
}

func NewDiskStore(uploadDir string) Store {
	return DiskStore{
		uploadDir: uploadDir,
	}
}
