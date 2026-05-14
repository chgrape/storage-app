package storage

import (
	"context"
	"io"
	"os"

	"github.com/chgrape/storage-app/internal/repository"
)

type Store interface {
	Save(ctx context.Context, file io.Reader, rec *repository.FileRecord) (*os.File, error)
	Delete(ctx context.Context, path string) error
	Download(ctx context.Context, path string) (*os.File, error)
}
