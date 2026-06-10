package pgsql

import (
	"context"

	"github.com/chgrape/storage-app/services/media-service/internal/repository"
	"github.com/google/uuid"
)

type pgFileRecordRepo struct {
	q *Queries
}

func (repo *pgFileRecordRepo) Delete(ctx context.Context, id uuid.UUID, user_id uuid.UUID) error {
	return repo.q.DeleteRecord(ctx, DeleteRecordParams{
		ID:     id,
		UserID: user_id,
	})
}

func (repo *pgFileRecordRepo) Get(ctx context.Context, id uuid.UUID, user_id uuid.UUID) (repository.FileRecord, error) {
	rec, err := repo.q.GetRecord(ctx, GetRecordParams{
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
		UserID:     rec.UserID,
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
			UserID:     rec.UserID,
			UploadedAt: rec.UploadedAt,
		})
	}

	return res, nil
}

func (repo *pgFileRecordRepo) Save(ctx context.Context, f repository.FileRecord) error {
	_, err := repo.q.CreateRecord(ctx, CreateRecordParams{
		ID:         f.ID,
		Filename:   f.Filename,
		Path:       f.Path,
		MimeType:   f.MIMEType,
		Size:       f.Size,
		UserID:     f.UserID,
		UploadedAt: f.UploadedAt,
	})
	return err
}

func NewPgFileRecordRepo(q *Queries) repository.FileRecordRepo {
	return &pgFileRecordRepo{q: q}
}
