package handler

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/chgrape/storage-app/internal/repository"
	"github.com/chgrape/storage-app/internal/service"
)

type fileRecordHandler struct {
	svc service.FileRecordSvc
}

func NewFileRecordHandler(svc service.FileRecordSvc) fileRecordHandler {
	return fileRecordHandler{
		svc: svc,
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
