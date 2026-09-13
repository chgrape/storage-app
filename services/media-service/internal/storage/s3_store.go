package storage

import (
	"context"
	"io"
	"os"

	"github.com/chgrape/storage-app/services/media-service/internal/repository"
)

type S3Store struct {
	ApiUrl            string
	AccessTokenId     string
	AccessTokenSecret string
}

func (s *S3Store) CreateFile(rec *repository.FileRecord) (io.WriteCloser, error) {
	panic("unimplemented")
}

// Delete implements [Store].
func (s *S3Store) Delete(ctx context.Context, path string) error {
	panic("unimplemented")
}

// Download implements [Store].
func (s *S3Store) Download(ctx context.Context, path string) (*os.File, error) {
	panic("unimplemented")
}

// Save implements [Store].
func (s *S3Store) Save(chunk io.Reader, file io.WriteCloser) error {
	panic("unimplemented")
}

func NewS3Store(ApiUrl string, AccessTokenId string, AccessTokenSecret string) Store {
	return &S3Store{
		ApiUrl:            ApiUrl,
		AccessTokenId:     AccessTokenId,
		AccessTokenSecret: AccessTokenSecret,
	}
}
