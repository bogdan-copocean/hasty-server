package events

import (
	"log"

	"github.com/nats-io/stan.go"
)

func ConnectToNats(clientId string) stan.Conn {

	url := "nats://nats-streaming:4222"
	// url := "nats://localhost:4222"

	sc, err := stan.Connect("test-cluster", clientId, stan.NatsURL(url),
		stan.Pings(1, 3),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, url)
	}

	log.Println("Connected to Nats")

	return sc
}
