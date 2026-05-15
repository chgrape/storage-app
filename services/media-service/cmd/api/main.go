package main

import (
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
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(1)
	}

	cfg := shared.Config{
		Host: os.Getenv("POSTGRES_HOST"),
		User: os.Getenv("POSTGRES_USER"),
		Pass: os.Getenv("POSTGRES_PASS"),
		Port: os.Getenv("POSTGRES_PORT"),
		DB:   os.Getenv("POSTGRES_DB"),
	}

	pool, err := shared.Connect(cfg)
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

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	grpcServer := grpc.NewServer()
	sv := handler.NewGRPCServer(&svc)
	pb.RegisterMediaServer(grpcServer, &sv)
	grpcServer.Serve(lis)
}
