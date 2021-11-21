package listeners

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/api-server/events"
	"github.com/bogdan-copocean/hasty-server/services/api-server/repository"
	"github.com/nats-io/stan.go"
)

type NatsListenerInterface interface {
	Listen(pubSubject string, ctx context.Context)
}

type natsListener struct {
	client         stan.Conn
	subject        string
	queueGroupName string
	repository     repository.MongoRepository
}

func NewJobFinishedListener(client stan.Conn, subject, queueGroupName string, repository repository.MongoRepository) NatsListenerInterface {
	return &natsListener{
		client:         client,
		queueGroupName: queueGroupName,
		subject:        subject,
		repository:     repository,
	}
}

func (nl *natsListener) Listen(pubSubject string, ctx context.Context) {

	msgHandler := func(msg *stan.Msg) {

		jobEvent := events.JobEvent{}

		err := json.Unmarshal(msg.Data, &jobEvent)
		if err != nil {
			log.Fatalf("could not unmarshal msg: %v\n", err.Error())
		}

		go func() {

			if err := nl.repository.UpdateJob(&jobEvent.Job, ctx); err != nil {
				log.Printf("could not insert to repo: %v\n", err.Error())
				return
			}

			msg.Ack()
		}()
	}

	aw, _ := time.ParseDuration("50s")

	_, err := nl.client.QueueSubscribe(nl.subject, nl.queueGroupName, msgHandler,
		stan.SetManualAckMode(),
		stan.AckWait(aw),
		stan.DeliverAllAvailable(),
		stan.DurableName("job-created-durable-name"),
	)

	if err != nil {
		log.Fatalf("queue subscribe error: %v\n", err)
	}

}
