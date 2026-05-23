package handler

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/chgrape/storage-app/services/media-service/internal/service"
	"github.com/chgrape/storage-app/shared"
	pb "github.com/chgrape/storage-app/shared/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMediaServer
	svc *service.FileRecordSvc
}

func NewGRPCServer(svc *service.FileRecordSvc) server {
	return server{
		svc: svc,
	}
}

func (s *server) Upload(ctx context.Context, uploadRequest *pb.UploadRequest) (*pb.UploadResponse, error) {
	user_id, err := uuid.Parse(uploadRequest.UserId)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(uploadRequest.Data)
	id, err := s.svc.Save(ctx, reader, shared.Metadata{
		Filename: uploadRequest.Filename,
		MimeType: uploadRequest.Mime,
		Size:     uploadRequest.Size,
	}, user_id)
	if err != nil {
		return nil, err
	}

	return &pb.UploadResponse{
		Id: id.String(),
	}, nil
}

func (s *server) Download(downloadRequest *pb.DownloadRequest, stream grpc.ServerStreamingServer[pb.DownloadResponse]) error {
	id, err := uuid.Parse(downloadRequest.Id)
	if err != nil {
		return err
	}

	user_id, err := uuid.Parse(downloadRequest.UserId)
	if err != nil {
		return err
	}

	record, file, err := s.svc.Download(stream.Context(), id, user_id)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 32*1024) //32kb buffer
	for {
		n, err := file.Read(buf)
		if err != nil {
			return err
		}
		if n > 0 {
			stream.Send(&pb.DownloadResponse{
				Data:     buf,
				Mime:     record.MIMEType,
				Filename: record.Filename,
			})
		}
		if err == io.EOF {
			break
		}
	}

	return nil
}
func (s *server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	user_id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	records, err := s.svc.ListRecords(ctx, user_id)
	if err != nil {
		return nil, err
	}

	res := &pb.ListResponse{}

	for _, r := range records {
		res.Files = append(res.Files, &pb.FileRecord{
			Id:         r.ID.String(),
			Filename:   r.Filename,
			MimeType:   r.MIMEType,
			Size:       r.Size,
			UploadedAt: r.UploadedAt.Format(time.RFC3339),
		})
	}

	return res, nil
}
func (s *server) Delete(ctx context.Context, deleteRequest *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	id, err := uuid.Parse(deleteRequest.Id)
	if err != nil {
		return nil, err
	}
	user_id, err := uuid.Parse(deleteRequest.UserId)
	if err != nil {
		return nil, err
	}

	err = s.svc.Erase(ctx, id, user_id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteResponse{Success: true}, nil
}
