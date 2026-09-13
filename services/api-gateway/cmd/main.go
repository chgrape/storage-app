package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/chgrape/storage-app/services/api-gateway/clients"
	"github.com/chgrape/storage-app/services/api-gateway/handler"
	"github.com/chgrape/storage-app/services/api-gateway/middleware"
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
	defer func() {
		err = mediaClient.Close()
		if err != nil {
			log.Fatalf("error:couldn't close connection: %v", err)
		}
	}()

	keycloakKeys, err := keyfunc.NewDefault([]string{
		"http://" + os.Getenv("KEYCLOAK_ADDR") + "/realms/media/protocol/openid-connect/certs",
	})
	if err != nil {
		log.Fatalf("failed to fetch JWKS: %v", err)
	}

	h := handler.NewMediaHandler(mediaClient)

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:         "0.0.0.0:8081",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	mux.HandleFunc("POST /login", handler.Login)

	mux.Handle("GET /download/{id}", middleware.Auth(keycloakKeys.Keyfunc, http.HandlerFunc(h.Download)))
	mux.Handle("GET /list", middleware.Auth(keycloakKeys.Keyfunc, http.HandlerFunc(h.List)))
	mux.Handle("POST /upload", middleware.Auth(keycloakKeys.Keyfunc, http.HandlerFunc(h.Upload)))
	mux.Handle("DELETE /delete/{id}", middleware.Auth(keycloakKeys.Keyfunc, http.HandlerFunc(h.Erase)))

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("error: couldn't start api gateway: %v", err)
	}
}
