package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bogdan-copocean/hasty-server/services/api-server/events"
	"github.com/bogdan-copocean/hasty-server/services/api-server/interfaces"
	"github.com/bogdan-copocean/hasty-server/services/api-server/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	clientId, err := os.Hostname()
	if err != nil {
		log.Fatalf("could not get the host name: %v\n", err)
	}

	// Nats
	natsConn := events.ConnectToNats(clientId)
	publisher := events.NewNatsPublisher(natsConn, "job:created")

	repo := repository.ConnectToMongo()
	handler := interfaces.NewHandler(repo, publisher)

	// Handlers
	r.Post("/", handler.PostHandler)
	r.Put("/", handler.PutHandler)
	r.Get("/", handler.GetHandler)

	http.ListenAndServe(":9090", r)
}
