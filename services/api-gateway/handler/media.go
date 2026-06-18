package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/chgrape/storage-app/services/api-gateway/clients"
	"github.com/chgrape/storage-app/services/api-gateway/middleware"
	pb "github.com/chgrape/storage-app/shared/gen"
)

type mediaHandler struct {
	media *clients.MediaClient
}

func NewMediaHandler(media *clients.MediaClient) *mediaHandler {
	return &mediaHandler{
		media: media,
	}
}

func (h *mediaHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	stream, err := h.media.Client.Download(r.Context(), &pb.DownloadRequest{Id: id, UserId: userID})
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	for {
		chunk, err := stream.Recv()
		if err != nil {
			return
		}
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", chunk.Mime)
		}

		if err == io.EOF {
			break
		}
		w.Write(chunk.Data)
	}
}

func (h *mediaHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	res, err := h.media.Client.List(r.Context(), &pb.ListRequest{UserId: userID})
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't fetch records: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Files)
}

func (h *mediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	part, err := reader.NextPart()
	if err != nil {
		http.Error(w, "couldn't get next part", http.StatusInternalServerError)
		return
	}

	ext := filepath.Ext(part.FileName())
	mime := mime.TypeByExtension(ext)

	stream, err := h.media.Client.Upload(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("error: couldn't upload data: %v", err), http.StatusInternalServerError)
		return
	}

	size, err := strconv.ParseInt(r.Header.Get("X-File-Size"), 10, 64)
	if err != nil {
		http.Error(w, "X-File-Size wasn't passed", http.StatusBadRequest)
		return
	}

	var res *pb.UploadResponse
	metadata := pb.FileMetadata{
		Filename: part.FileName(),
		Mime:     mime,
		Size:     size,
		UserId:   userID,
	}
	stream.Send(&pb.UploadRequest{
		Payload: &pb.UploadRequest_Metadata{
			Metadata: &metadata,
		},
	})

	buf := make([]byte, 1024*1024) // 1 MB buffer
	for {
		n, err := part.Read(buf)
		if err == io.EOF {
			if n > 0 {
				req := &pb.UploadRequest_Data{
					Data: buf[:n],
				}

				err = stream.Send(&pb.UploadRequest{Payload: req})
				if err != nil {
					http.Error(w, fmt.Sprintf("error: couldn't stream request: %v", err), http.StatusInternalServerError)
					return
				}
			}
			res, err = stream.CloseAndRecv()
			if err != nil {
				http.Error(w, fmt.Sprintf("error: couldn't close stream: %v", err), http.StatusInternalServerError)
				return
			}
			if res.Error != "" {
				http.Error(w, fmt.Sprintf("error: error during saving procedure: %v", err), http.StatusInternalServerError)
				return
			}
			break
		}

		req := &pb.UploadRequest_Data{
			Data: buf[:n],
		}

		err = stream.Send(&pb.UploadRequest{Payload: req})
		if err != nil {
			http.Error(w, fmt.Sprintf("error: couldn't stream request: %v", err), http.StatusInternalServerError)
			return
		}

	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Id)
}

func (h *mediaHandler) Erase(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("id")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	res, err := h.media.Client.Delete(r.Context(), &pb.DeleteRequest{
		Id:     uuid,
		UserId: userID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error: error deleting file: %v", err), http.StatusInternalServerError)
		return
	}

	if !res.Success {
		w.Write([]byte("Record couldn't be deleted"))
		return
	}

	w.Write([]byte("Record successfully deleted"))
}
