package main

import (
	"log"
	"net/http"
	"os"

	"github.com/chgrape/storage-app/internal/handler"
	"github.com/chgrape/storage-app/internal/pgsql"
	"github.com/chgrape/storage-app/internal/service"
	"github.com/chgrape/storage-app/internal/storage"
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

	repo := storage.NewPgFileRecordRepo(q)
	svc := service.NewFileRecordSvc(repo)
	h := handler.NewFileRecordHandler(svc)

	router := http.NewServeMux()

	router.HandleFunc("POST /upload", h.Upload)
	router.HandleFunc("GET /download/{id}", h.Download)
	router.HandleFunc("GET /list", h.ListRecords)
	router.HandleFunc("DELETE /delete/{id}", h.Erase)

	http.ListenAndServe("localhost:8081", router)
}
