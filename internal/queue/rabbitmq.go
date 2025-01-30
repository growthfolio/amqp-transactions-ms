package queue

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	channel   *amqp.Channel
	queueName string
}

func (r RabbitMQ) Publish(message []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r RabbitMQ) Close() {

}

func NewRabbitMQ(url, queueName string) (*RabbitMQ, error) {
	var conn *amqp.Connection
	var err error
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ, retrying in %d seconds...\n", i+1)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("could not connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("could not create channel: %v", err)
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("could not declare queue: %v", err)
	}

	return &RabbitMQ{channel: ch, queueName: queueName}, nil
}
