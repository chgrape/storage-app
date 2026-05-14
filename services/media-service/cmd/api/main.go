package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/chgrape/storage-app/services/media-service/internal/handler"
	"github.com/chgrape/storage-app/services/media-service/internal/pgsql"
	"github.com/chgrape/storage-app/services/media-service/internal/service"
	"github.com/chgrape/storage-app/services/media-service/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(1)
	}

	cfg := pgsql.Config{
		Host: os.Getenv("POSTGRES_HOST"),
		User: os.Getenv("POSTGRES_USER"),
		Pass: os.Getenv("POSTGRES_PASS"),
		Port: os.Getenv("POSTGRES_PORT"),
		DB:   os.Getenv("POSTGRES_DB"),
	}

	pool, err := pgsql.Connect(cfg)
	if err != nil {
		log.Fatalf("Connection to database couldn't be established")
	}
	q := pgsql.New(pool)

	uploadDir, err := filepath.Abs("./uploads")
	if err != nil {
		log.Fatalf("Upload directory doesn't exist")
		return
	}

	repo := storage.NewPgFileRecordRepo(q)
	store := storage.NewDiskStore(uploadDir)
	svc := service.NewFileRecordSvc(repo, store)
	h := handler.NewFileRecordHandler(svc)

	router := http.NewServeMux()

	router.HandleFunc("POST /upload", h.Upload)
	router.HandleFunc("GET /download/{id}", h.Download)
	router.HandleFunc("GET /list", h.ListRecords)
	router.HandleFunc("DELETE /delete/{id}", h.Erase)

	http.ListenAndServe("localhost:8081", router)
}
