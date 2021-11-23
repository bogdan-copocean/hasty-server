package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bogdan-copocean/hasty-server/services/job-server/events"
	"github.com/bogdan-copocean/hasty-server/services/job-server/events/listeners"
	"github.com/bogdan-copocean/hasty-server/services/job-server/events/publishers"
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
	conn := events.ConnectToNats(clientId)

	// Job Finished Publisher
	jobFinishedSubject := publishers.JobFinishedSubject
	jobFinishedPublisher := publishers.NewJobEventPublisher(conn, jobFinishedSubject)

	// Job Cancelled Publisher
	jobCancelledSubject := publishers.JobCancelledSubject
	jobCancelledPublisher := publishers.NewJobEventPublisher(conn, jobCancelledSubject)

	// Job Created Listener
	jobCreatedListenerSubject := "job:created"
	jobCreatedQGroup := "job-created-group"
	jobCreatedListener := listeners.NewJobCreatedListener(conn, jobCreatedListenerSubject, jobCreatedQGroup, jobFinishedPublisher, jobCancelledPublisher, repo)

	// Listen and publish events
	jobCreatedListener.ListenAndPublish()

	http.ListenAndServe(":9091", r)
}
