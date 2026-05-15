package main

import (
	"log"
	"net/http"

	"github.com/chgrape/storage-app/services/api_gateway/clients"
	"github.com/chgrape/storage-app/services/api_gateway/handler"
)

func main() {
	mediaClient, err := clients.NewMediaClient("localhost:9090")
	if err != nil {
		log.Fatalf("error establishing connection to service: %s", err)
		return
	}
	defer mediaClient.Close()

	h := handler.NewMediaHandler(mediaClient)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /download/{id}", h.Download)
	mux.HandleFunc("GET /list", h.List)
	http.ListenAndServe("localhost:8081", mux)
}
