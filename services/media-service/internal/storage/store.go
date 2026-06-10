package storage

import (
	"context"
	"io"
	"os"

	"github.com/chgrape/storage-app/services/media-service/internal/repository"
)

type Store interface {
	Save(chunk io.Reader, file io.WriteCloser) error
	Delete(ctx context.Context, path string) error
	Download(ctx context.Context, path string) (*os.File, error)
	CreateFile(rec *repository.FileRecord) (io.WriteCloser, error)
}
