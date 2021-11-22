package listeners

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/job-server/events"
	"github.com/bogdan-copocean/hasty-server/services/job-server/events/publishers"
	"github.com/bogdan-copocean/hasty-server/services/job-server/repository"
	"github.com/nats-io/stan.go"
)

const (
	MinSleepTime        = 15
	MaxSleepTime        = 45
	CancellationJobTime = 46 * time.Second
)

type NatsListenerInterface interface {
	ListenAndPublish()
}

type natsListener struct {
	client             stan.Conn
	subject            string
	queueGroupName     string
	finishedPublisher  publishers.JobEventPublisher
	cancelledPublisher publishers.JobEventPublisher
	repository         repository.MongoRepository
}

func NewJobCreatedListener(client stan.Conn, subject, queueGroupName string, finishedPublisher, cancelledPublisher publishers.JobEventPublisher, repository repository.MongoRepository) NatsListenerInterface {
	return &natsListener{
		client:             client,
		subject:            subject,
		queueGroupName:     queueGroupName,
		finishedPublisher:  finishedPublisher,
		cancelledPublisher: cancelledPublisher,
		repository:         repository,
	}
}

func (nl *natsListener) ListenAndPublish() {

	aw, _ := time.ParseDuration("50s")

	_, err := nl.client.QueueSubscribe(nl.subject, nl.queueGroupName, func(msg *stan.Msg) {
		go msgHandler(msg, nl.client, nl.finishedPublisher, nl.cancelledPublisher, nl.repository)
	},
		stan.SetManualAckMode(),
		stan.AckWait(aw),
		stan.DeliverAllAvailable(),
		stan.DurableName("job-created-durable-name"),
	)

	if err != nil {
		log.Fatalf("queue subscribe error: %v\n", err)
	}

}

func msgHandler(msg *stan.Msg, client stan.Conn, finishedPublisher, cancelledPublisher publishers.JobEventPublisher, repository repository.MongoRepository) {
	jobEvent := events.JobEvent{}
	doneCh := make(chan struct{})

	// ctx to trigger cancellation inside sleeping goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := json.Unmarshal(msg.Data, &jobEvent)
	if err != nil {
		log.Fatal(err.Error())
	}

	sleepTimeUsed := rand.Intn(MaxSleepTime-MinSleepTime) + MinSleepTime
	jobEvent.SleepTimeUsed = sleepTimeUsed

	go func() {

		time.Sleep(time.Duration(sleepTimeUsed) * time.Second)

		select {
		case <-ctx.Done():
			jobEvent.Job.Status = "cancelled"

			if err := repository.SetJob(&jobEvent); err != nil {
				log.Fatalf("could not insert cancelled msg to repo: %v\n", err.Error())
			}

			// Publish job cancelled
			if err := finishedPublisher.PublishData(&jobEvent); err != nil {
				log.Fatalf("could not publish cancelled job event: %v", err.Error())
			}

			msg.Ack()

		default:
			jobEvent.Job.Status = "finished"

			if err := repository.SetJob(&jobEvent); err != nil {
				log.Fatalf("could not insert finished msg to repo: %v\n", err.Error())
			}

			// Publish job finished
			if err := finishedPublisher.PublishData(&jobEvent); err != nil {
				log.Fatalf("could not publish finished job event: %v", err.Error())
			}
			doneCh <- struct{}{}
		}

	}()

	select {
	case <-doneCh:
		msg.Ack()
	case <-time.After(CancellationJobTime):
		return
	}
}
