package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

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

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	mime := mime.TypeByExtension(ext)

	stream, err := h.media.Client.Upload(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("error: couldn't upload data: %v", err), http.StatusInternalServerError)
		return
	}

	var res *pb.UploadResponse
	metadata := pb.FileMetadata{
		Filename: header.Filename,
		Mime:     mime,
		Size:     header.Size,
		UserId:   userID,
	}
	stream.Send(&pb.UploadRequest{
		Payload: &pb.UploadRequest_Metadata{
			Metadata: &metadata,
		},
	})

	buf := make([]byte, 1024*1024) // 1 MB buffer
	for {
		_, err := file.Read(buf)
		if err == io.EOF {
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
			Data: buf,
		}

		err = stream.Send(&pb.UploadRequest{Payload: req})
		if err != nil {
			http.Error(w, fmt.Sprintf("error: couldn't stream request: %v", err), http.StatusInternalServerError)
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
