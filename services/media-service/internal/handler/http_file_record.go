package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/chgrape/storage-app/services/media-service/internal/repository"
	"github.com/chgrape/storage-app/services/media-service/internal/service"
	"github.com/google/uuid"
)

type fileRecordHandler struct {
	svc service.FileRecordSvc
}

func NewFileRecordHandler(svc service.FileRecordSvc) fileRecordHandler {
	return fileRecordHandler{
		svc: svc,
	}

}

func (h *fileRecordHandler) ListRecords(w http.ResponseWriter, r *http.Request) {
	records, err := h.svc.ListRecords(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func (h *fileRecordHandler) Erase(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.svc.Erase(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "Record successfully deleted")
}

func (h *fileRecordHandler) Download(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rec, file, err := h.svc.Download(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	w.Header().Add("Content-Type", rec.MIMEType)
	w.Header().Add("Content-Length", strconv.FormatInt(rec.Size, 10))
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *fileRecordHandler) Upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	mime := mime.TypeByExtension(ext)

	metadata := repository.Metadata{
		Filename: header.Filename,
		Size:     header.Size,
		MimeType: mime,
	}

	id, err := h.svc.Save(r.Context(), file, metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Id of file: %s", id.String())
}
