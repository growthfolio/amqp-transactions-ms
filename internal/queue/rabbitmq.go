package queue

import (
	"os"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	channel   *amqp.Channel
	queueName string
}

func (r *RabbitMQ) Publish(message []byte) error {
    return r.channel.Publish(
        "",             // exchange
        r.queueName,    // routing key
        false,          // mandatory
        false,          // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        message,
        })
}

func (r RabbitMQ) Close() {

}

func NewRabbitMQ(url, queueName string) (*RabbitMQ, error) {
    if url == "" {
        url = os.Getenv("RABBITMQ_URL")
        if url == "" {
            url = "amqp://guest:guest@rabbitmq:5672/"
        }
    }

    var conn *amqp.Connection
    var err error
    maxRetries := 10
    for i := 0; i < maxRetries; i++ {
        conn, err = amqp.Dial(url)
        if err == nil {
            break
        }
        log.Printf("RabbitMQ connection attempt %d failed: %v. Retrying in %d seconds...\n", 
                   i+1, err, i+1)
        time.Sleep(time.Duration(i+1) * time.Second)
    }

    if err != nil {
        return nil, fmt.Errorf("could not connect to RabbitMQ after %d attempts: %v", 
                                maxRetries, err)
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("could not create channel: %v", err)
    }

    _, err = ch.QueueDeclare(
        queueName, // name
        true,      // durable
        false,     // delete when unused
        false,     // exclusive
        false,     // no-wait
        nil,       // arguments
    )
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, fmt.Errorf("could not declare queue: %v", err)
    }

    return &RabbitMQ{
        channel:   ch,
        queueName: queueName,
    }, nil
}
