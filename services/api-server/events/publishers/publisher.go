package publishers

import (
	"encoding/json"
	"fmt"

	"github.com/bogdan-copocean/hasty-server/services/api-server/events"
	"github.com/nats-io/stan.go"
)

type NatsPublisherInterface interface {
	PublishData(event *events.JobEvent, doneCh chan<- struct{}, errChan chan<- error)
}

type natsPublisher struct {
	Client  stan.Conn
	Subject string
}

func NewNatsPublisher(client stan.Conn, subject string) NatsPublisherInterface {
	return &natsPublisher{
		Client:  client,
		Subject: subject,
	}
}

func (nl *natsPublisher) PublishData(event *events.JobEvent, doneCh chan<- struct{}, errChan chan<- error) {

	data, err := json.Marshal(event)
	if err != nil {
		errChan <- err
		close(errChan)
	}
	fmt.Println(string(data))
	if err := nl.Client.Publish(nl.Subject, data); err != nil {
		errChan <- err
		close(errChan)
	}

	doneCh <- struct{}{}
	close(doneCh)
}
