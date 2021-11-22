package publishers

import (
	"encoding/json"
	"log"

	"github.com/bogdan-copocean/hasty-server/services/job-server/events"
	"github.com/nats-io/stan.go"
)

const (
	JobFinishedSubject  = "job:finished"
	JobCancelledSubject = "job:cancelled"
)

type JobEventPublisher interface {
	PublishData(jobEvent *events.JobEvent) error
}

type jobEventPublisher struct {
	Client  stan.Conn
	Subject string
}

func NewJobEventPublisher(client stan.Conn, subject string) JobEventPublisher {
	return &jobEventPublisher{
		Client:  client,
		Subject: subject,
	}
}

func (nl *jobEventPublisher) PublishData(jobEvent *events.JobEvent) error {

	data, err := json.Marshal(jobEvent)
	if err != nil {
		log.Fatalf("could not marshal event with jobId: %v, reason: %v", jobEvent.Job.Id, err.Error())
	}

	if err := nl.Client.Publish(nl.Subject, data); err != nil {
		return err
	}

	return nil
}
