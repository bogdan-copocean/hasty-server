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

	// Job Created Publisher
	jobCreatedSubject := "job:created"
	publisher := publishers.NewJobEventPublisher(conn, jobCreatedSubject)

	// Job Finished listener
	jobEventFinishedSubject := "job:finished"
	jobEventFinishedQGroup := "job-finished-group"
	finishedListener := listeners.NewJobEventListener(conn, jobEventFinishedSubject, jobEventFinishedQGroup, service)
	finishedListener.Listen()

	// Job Cancelled listener
	jobEventCancelledSubject := "job:cancelled"
	jobEventCancelledQGroup := "job-cancelled-group"
	cancelledListener := listeners.NewJobEventListener(conn, jobEventCancelledSubject, jobEventCancelledQGroup, service)
	cancelledListener.Listen()

	// Handlers
	handler := interfaces.NewApiHandler(service, publisher)

	r.Post("/", handler.PostHandler)
	r.Get("/{jobId}", handler.GetHandler)

	http.ListenAndServe(":9090", r)
}
