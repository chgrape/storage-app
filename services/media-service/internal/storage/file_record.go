package storage

import (
	"context"

	"github.com/chgrape/storage-app/services/media-service/internal/pgsql"
	"github.com/chgrape/storage-app/services/media-service/internal/repository"
	"github.com/google/uuid"
)

type pgFileRecordRepo struct {
	q *pgsql.Queries
}

func (repo *pgFileRecordRepo) Delete(ctx context.Context, id uuid.UUID, user_id uuid.UUID) error {
	return repo.q.DeleteRecord(ctx, pgsql.DeleteRecordParams{
		ID:     id,
		UserID: user_id,
	})
}

func (repo *pgFileRecordRepo) Get(ctx context.Context, id uuid.UUID, user_id uuid.UUID) (repository.FileRecord, error) {
	rec, err := repo.q.GetRecord(ctx, pgsql.GetRecordParams{
		ID:     id,
		UserID: user_id,
	})
	if err != nil {
		return repository.FileRecord{}, err
	}

	return repository.FileRecord{
		ID:         rec.ID,
		Filename:   rec.Filename,
		MIMEType:   rec.MimeType,
		Size:       rec.Size,
		Path:       rec.Path,
		UploadedAt: rec.UploadedAt,
	}, nil

}

func (repo *pgFileRecordRepo) List(ctx context.Context, user_id uuid.UUID) ([]repository.FileRecord, error) {
	records, err := repo.q.ListRecords(ctx, user_id)
	if err != nil {
		return nil, err
	}

	var res []repository.FileRecord

	for _, rec := range records {
		res = append(res, repository.FileRecord{
			ID:         rec.ID,
			Filename:   rec.Filename,
			MIMEType:   rec.MimeType,
			Size:       rec.Size,
			Path:       rec.Path,
			UploadedAt: rec.UploadedAt,
		})
	}

	return res, nil
}

func (repo *pgFileRecordRepo) Save(ctx context.Context, f repository.FileRecord) (repository.FileRecord, error) {
	rec, err := repo.q.CreateRecord(ctx, pgsql.CreateRecordParams{
		ID:         f.ID,
		Filename:   f.Filename,
		Path:       f.Path,
		MimeType:   f.MIMEType,
		Size:       f.Size,
		UploadedAt: f.UploadedAt,
	})
	return repository.FileRecord{
		ID:         rec.ID,
		Filename:   rec.Filename,
		Path:       rec.Path,
		MIMEType:   rec.MimeType,
		Size:       rec.Size,
		UploadedAt: rec.UploadedAt,
	}, err
}

func NewPgFileRecordRepo(q *pgsql.Queries) repository.FileRecordRepo {
	return &pgFileRecordRepo{q: q}
}
