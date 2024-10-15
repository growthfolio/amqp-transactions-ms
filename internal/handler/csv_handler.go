package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/felipemacedo1/transaction-producer-ms/internal/dto"
	"github.com/streadway/amqp"
)

// ProcessCSVFile processes a CSV file and sends data to RabbitMQ.
func ProcessCSVFile(filePath string) error {
	conn, ch, err := connectRabbitMQ()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	queue, err := declareQueue(ch)
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	return processRecords(file, ch, queue.Name)
}

// connectRabbitMQ establishes a connection and channel with RabbitMQ.
func connectRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		return nil, nil, fmt.Errorf("rabbitmq connection could not be established: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("channel could not be created: %v", err)
	}

	return conn, ch, nil
}

// declareQueue declares the queue to be used.
func declareQueue(ch *amqp.Channel) (amqp.Queue, error) {
	queue, err := ch.QueueDeclare(
		"transaction_queue", // queue name
		true,                // durable
		false,               // auto-delete
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("queue could not be declared: %v", err)
	}
	return queue, nil
}

// processRecords processes each record in the CSV and sends to RabbitMQ.
func processRecords(file io.Reader, ch *amqp.Channel, queueName string) error {
	reader := csv.NewReader(file)
	reader.Comma = ';'

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("csv could not be read: %v", err)
			continue
		}

		transaction, err := parseTransaction(record)
		if err != nil {
			log.Printf("transaction could not be parsed: %v", err)
			continue
		}

		if err := publishTransaction(ch, queueName, transaction); err != nil {
			log.Printf("transaction could not be sent: %v", err)
		}
	}
	return nil
}

// parseTransaction parses the CSV record into a Transaction object.
func parseTransaction(record []string) (*dto.Transaction, error) {
	age, err := strconv.Atoi(record[4])
	if err != nil {
		return nil, fmt.Errorf("age could not be converted: %v", err)
	}

	amount, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return nil, fmt.Errorf("amount could not be converted: %v", err)
	}

	installments, err := strconv.Atoi(record[6])
	if err != nil {
		return nil, fmt.Errorf("installments could not be converted: %v", err)
	}

	date, err := time.Parse(time.RFC3339, record[1])
	if err != nil {
		return nil, fmt.Errorf("date could not be converted: %v", err)
	}

	return &dto.Transaction{
		TransactionID: record[0],
		Date:          date,
		ClientID:      record[2],
		Name:          record[3],
		Age:           age,
		Amount:        amount,
		Installments:  installments,
	}, nil
}

// publishTransaction sends a transaction message to RabbitMQ.
func publishTransaction(ch *amqp.Channel, queueName string, transaction *dto.Transaction) error {
	jsonData, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("JSON could not be created: %v", err)
	}

	return ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)
}
