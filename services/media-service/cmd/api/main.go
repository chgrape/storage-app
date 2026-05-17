package main

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/chgrape/storage-app/services/media-service/internal/handler"
	"github.com/chgrape/storage-app/services/media-service/internal/pgsql"
	"github.com/chgrape/storage-app/services/media-service/internal/service"
	"github.com/chgrape/storage-app/services/media-service/internal/storage"
	"github.com/chgrape/storage-app/shared"
	pb "github.com/chgrape/storage-app/shared/gen"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	var err error
	var cfg shared.Config
	var pool *pgxpool.Pool
	var uploadDir string

	if os.Getenv("ENVIRONMENT") == "dev" {
		err = godotenv.Load()
		if err != nil {
			panic(1)
		}
		cfg = shared.Config{
			Host: os.Getenv("POSTGRES_HOST"),
			User: os.Getenv("POSTGRES_USER"),
			Pass: os.Getenv("POSTGRES_PASS"),
			Port: os.Getenv("POSTGRES_PORT"),
			DB:   os.Getenv("POSTGRES_DB"),
		}
		pool, err = shared.Connect(cfg)
		if err != nil {
			log.Fatalf("Connection to dev database couldn't be established: %v", err)
		}
		uploadDir, err = filepath.Abs("./uploads")
		if err != nil {
			log.Fatalf("Upload directory doesn't exist")
			return
		}
	} else if os.Getenv("ENVIRONMENT") == "prod" {
		pool, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatalf("Connection to prod database couldn't be established: %v", err)
		}
	} else {
		log.Fatalf("Connection to database couldn't be established: %v", err)
	}

	q := pgsql.New(pool)

	repo := storage.NewPgFileRecordRepo(q)
	store := storage.NewDiskStore(uploadDir)
	svc := service.NewFileRecordSvc(repo, store)

	lis, err := net.Listen("tcp", ":"+os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	grpcServer := grpc.NewServer()
	sv := handler.NewGRPCServer(&svc)
	pb.RegisterMediaServer(grpcServer, &sv)
	grpcServer.Serve(lis)
}
