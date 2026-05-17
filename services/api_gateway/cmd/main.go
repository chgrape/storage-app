package main

import (
	"log"
	"net/http"
	"os"

	"github.com/chgrape/storage-app/services/api_gateway/clients"
	"github.com/chgrape/storage-app/services/api_gateway/handler"
	"github.com/chgrape/storage-app/services/api_gateway/middleware"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	mediaAddr := os.Getenv("MEDIA_SERVICE_ADDR")

	mediaClient, err := clients.NewMediaClient(mediaAddr)
	if err != nil {
		log.Fatalf("error establishing connection to service: %s", err)
		return
	}
	defer mediaClient.Close()

	keycloakKeys, err := NewDefault([]string{
		"http://keycloak-service:8080/realms/media/protocol/openid-connect/certs",
	})
	if err != nil {
		log.Fatalf("failed to fetch JWKS: %v", err)
	}

	h := handler.NewMediaHandler(mediaClient)

	mux := http.NewServeMux()

	mux.Handle("GET /download/{id}", middleware.Auth(h.Download))
	mux.HandleFunc("GET /list", h.List)
	mux.HandleFunc("POST /upload", h.Upload)
	mux.HandleFunc("DELETE /delete/{id}", h.Erase)
	http.ListenAndServe("localhost:8081", mux)
}
