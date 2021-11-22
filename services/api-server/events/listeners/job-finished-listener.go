package listeners

import (
	"encoding/json"
	"log"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/api-server/app"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events"
	"github.com/nats-io/stan.go"
)

type JobEventListenerInterface interface {
	Listen()
}

type jobEventListener struct {
	client         stan.Conn
	subject        string
	queueGroupName string
	apiService     app.ApiService
}

func NewJobEventListener(client stan.Conn, subject, queueGroupName string, apiService app.ApiService) JobEventListenerInterface {
	return &jobEventListener{
		client:         client,
		queueGroupName: queueGroupName,
		subject:        subject,
		apiService:     apiService,
	}
}

func (nl *jobEventListener) Listen() {

	aw, _ := time.ParseDuration("50s")

	_, err := nl.client.QueueSubscribe(nl.subject, nl.queueGroupName, func(msg *stan.Msg) {
		go msgHandler(msg, nl.apiService)
	},
		stan.SetManualAckMode(),
		stan.AckWait(aw),
		stan.DeliverAllAvailable(),
		stan.DurableName("job-created-durable-name"),
	)

	if err != nil {
		log.Fatalf("job finished listener subscribe error: %v\n", err)
	}
}

func msgHandler(msg *stan.Msg, apiService app.ApiService) {
	jobEvent := events.JobEvent{}

	err := json.Unmarshal(msg.Data, &jobEvent)
	if err != nil {
		log.Fatalf("could not unmarshal msg: %v\n", err.Error())
	}

	if err := apiService.UpdateJob(jobEvent.Job); err != nil {
		log.Fatalf("could not update to repo: %v\n", err.Error())
	}

	msg.Ack()
}
