package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/job-server/events"
	"github.com/bogdan-copocean/hasty-server/services/job-server/events/listeners"
	"github.com/bogdan-copocean/hasty-server/services/job-server/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	ctx, cancel := context.WithTimeout(context.Background(), 47*time.Second)
	defer cancel()

	clientId, err := os.Hostname()
	if err != nil {
		log.Fatalf("could not get the host name: %v\n", err)
	}

	// Mongo Repository
	repo := repository.ConnectToMongo()

	// Nats
	conn := events.ConnectToNats(clientId + "asoda")

	listenerSubject := "job:created"
	listenerQueueGroupName := "job-created-group"
	// Create Job Created Listener
	listener := listeners.NewJobCreatedListener(conn, listenerSubject, listenerQueueGroupName, repo)

	publishSubject := "job:finished"
	// Listen and publish events
	listener.ListenAndPublish(publishSubject, ctx)

	http.ListenAndServe(":9091", r)
}
