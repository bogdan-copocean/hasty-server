package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bogdan-copocean/hasty-server/services/api-server/app"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events/listeners"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events/publishers"
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

	// Mongo Repository
	repo := repository.ConnectToMongo()

	// Services
	service := app.NewApiService(repo)

	// Nats
	conn := events.ConnectToNats(clientId)
	publisher := publishers.NewNatsPublisher(conn, "job:created")

	listenerSubject := "job:finished"
	listenerQueueGroup := "job-finished-group"
	listener := listeners.NewJobFinishedListener(conn, listenerSubject, listenerQueueGroup, service)

	listener.Listen(listenerSubject)

	// Handlers
	handler := interfaces.NewHandler(service, publisher)

	r.Post("/", handler.PostHandler)
	r.Put("/", handler.PutHandler)
	r.Get("/", handler.GetHandler)

	http.ListenAndServe(":9090", r)
}
