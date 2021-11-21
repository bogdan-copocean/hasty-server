package publishers

import (
	"log"

	"github.com/nats-io/stan.go"
)

type NatsPublisherInterface interface {
	PublishData()
}

type natsPublisher struct {
	Client  stan.Conn
	Subject string
	data    []byte
}

func NewNatsPublisher(client stan.Conn, subject string) NatsPublisherInterface {
	return &natsPublisher{
		Client:  client,
		Subject: subject,
	}
}

func (nl *natsPublisher) PublishData() {
	if err := nl.Client.Publish(string(nl.Subject), nl.data); err != nil {
		log.Println(err)
	}
	log.Println("event published")
}
