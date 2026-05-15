package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chgrape/storage-app/services/api_gateway/clients"
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

	stream, err := h.media.Client.Download(r.Context(), &pb.DownloadRequest{Id: id})
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.media.Client.List(ctx, &pb.ListRequest{})
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't fetch records: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Files)
}
