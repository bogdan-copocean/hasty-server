package events

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/stan.go"
)

type NatsListenerInterface interface {
	SetSubscriptionOptions()
	Listen()
}

type natsListener struct {
	Client         stan.Conn
	QueueGroupName string
	Subject        string
	AckWait        time.Duration
	options        []stan.SubscriptionOption
}

func NewNatsListener(client stan.Conn, queueGroupName, subject string, ackwait time.Duration) NatsListenerInterface {
	return &natsListener{
		Client:         client,
		QueueGroupName: queueGroupName,
		Subject:        subject,
		AckWait:        ackwait,
	}

}

func (nl *natsListener) SetSubscriptionOptions() {
	nl.options = append(nl.options, stan.SetManualAckMode(), stan.DeliverAllAvailable(), stan.AckWait(nl.AckWait), stan.DurableName(nl.QueueGroupName))
}

func (nl *natsListener) Listen() {
	_, err := nl.Client.QueueSubscribe(string(nl.Subject), nl.QueueGroupName, func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))

	}, nl.options...)

	if err != nil {
		log.Fatal(err.Error())
	}

}
