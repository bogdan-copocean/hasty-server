package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bogdan-copocean/hasty-server/services/job-server/events"
	"github.com/bogdan-copocean/hasty-server/services/job-server/events/listeners"
	"github.com/bogdan-copocean/hasty-server/services/job-server/repository"
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

	// Nats
	conn := events.ConnectToNats(clientId + "1")

	listenerSubject := "job:created"
	listenerQueueGroupName := "job-created-group"
	// Create Job Created Listener
	listener := listeners.NewJobCreatedListener(conn, listenerSubject, listenerQueueGroupName, repo)

	// Listen and publish events
	listener.ListenAndPublish()

	http.ListenAndServe(":9091", r)
}
