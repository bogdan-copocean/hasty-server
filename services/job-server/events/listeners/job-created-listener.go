package listeners

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/job-server/events"
	"github.com/bogdan-copocean/hasty-server/services/job-server/repository"
	"github.com/nats-io/stan.go"
)

type NatsListenerInterface interface {
	ListenAndPublish(pubSubject string)
}

type natsListener struct {
	client         stan.Conn
	subject        string
	queueGroupName string
	repository     repository.MongoRepository
}

func NewJobCreatedListener(client stan.Conn, subject, queueGroupName string, repository repository.MongoRepository) NatsListenerInterface {
	return &natsListener{
		client:         client,
		queueGroupName: queueGroupName,
		subject:        subject,
		repository:     repository,
	}
}

func (nl *natsListener) ListenAndPublish(pubSubject string) {
	minSleepTime := 15
	maxSleepTime := 45

	msgHandler := func(msg *stan.Msg) {

		jobEvent := events.JobEvent{}

		err := json.Unmarshal(msg.Data, &jobEvent)
		if err != nil {
			log.Fatal(err.Error())
		}

		sleepTimeUsed := rand.Intn(maxSleepTime-minSleepTime) + minSleepTime

		go func() {

			fmt.Println("sleeping...", sleepTimeUsed)
			jobEvent.SleepTimeUsed = sleepTimeUsed

			if err := nl.repository.SetJob(&jobEvent); err != nil {
				log.Printf("could not insert to repo: %v\n", err.Error())
				return
			}

			time.Sleep(time.Duration(sleepTimeUsed) * time.Second)
			fmt.Println("done sleeping", sleepTimeUsed)

			jobEventBytes, err := json.Marshal(&jobEvent)
			if err != nil {
				log.Fatalf("could not marshal job event: %v\n", err)
			}

			// Publish job finished
			if err := nl.client.Publish(pubSubject, jobEventBytes); err != nil {
				log.Fatalf("could not publish event: %v\n", err)
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
