package handler

import (
	"bytes"
	"context"
	"errors"
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

func (s *server) Upload(stream grpc.ClientStreamingServer[pb.UploadRequest, pb.UploadResponse]) error {

	init_req, err := stream.Recv()
	if err != nil {
		return err
	}

	metadata := init_req.GetMetadata()
	if metadata == nil {
		return errors.New("no metadata provided")
	}

	user_id, err := uuid.Parse(metadata.UserId)
	if err != nil {
		return err
	}
	file, rec, err := s.svc.Init(shared.Metadata{
		Filename: metadata.Filename,
		Size:     metadata.Size,
		MimeType: metadata.Mime,
	}, user_id)
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			stream.SendAndClose(&pb.UploadResponse{
				Id:    "",
				Saved: false,
				Error: err.Error(),
			})
			return err
		}

		data := req.GetData()
		if data == nil {
			return errors.New("no data provided")
		}

		chunk := bytes.NewReader(data)

		err = s.svc.Write(chunk, file)
		if err != nil {
			stream.SendAndClose(&pb.UploadResponse{
				Id:    "",
				Saved: false,
				Error: err.Error(),
			})
			return err
		}
	}

	id, err := s.svc.Commit(stream.Context(), *rec)
	if err != nil {
		stream.SendAndClose(&pb.UploadResponse{
			Id:    "",
			Saved: false,
			Error: err.Error(),
		})
		return err
	}

	stream.SendAndClose(&pb.UploadResponse{
		Id:    id.String(),
		Saved: true,
		Error: "",
	})

	return nil
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
