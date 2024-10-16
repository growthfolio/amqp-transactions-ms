package queue

import (
	"fmt"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	channel   *amqp.Channel
	queueName string
}

func NewRabbitMQ(url, queueName string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
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

func (r *RabbitMQ) Publish(message []byte) error {
	return r.channel.Publish(
		"",
		r.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
}

func (r *RabbitMQ) Close() {
	r.channel.Close()
}
