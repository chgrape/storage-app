package handler

import (
	"context"
	"io"
	"time"

	"github.com/chgrape/storage-app/services/media-service/internal/service"
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
	return nil, nil
}

func (s *server) Download(downloadRequest *pb.DownloadRequest, stream grpc.ServerStreamingServer[pb.DownloadResponse]) error {
	id, err := uuid.Parse(downloadRequest.Id)
	if err != nil {
		return err
	}

	record, file, err := s.svc.Download(stream.Context(), id)
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
	records, err := s.svc.ListRecords(ctx)
	if err != nil {
		return nil, err
	}

	res := &pb.ListResponse{}

	for _, r := range records {
		res.Files = append(res.Files, &pb.UploadResponse{
			Id:         r.ID.String(),
			Filename:   r.Filename,
			MimeType:   r.MIMEType,
			Size:       r.Size,
			UploadedAt: r.UploadedAt.Format(time.RFC3339),
		})
	}

	return res, nil
}
func (s *server) Delete(context.Context, *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return nil, nil
}
