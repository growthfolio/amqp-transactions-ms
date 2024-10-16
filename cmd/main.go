package main

import (
	"log"

	"github.com/felipemacedo1/transaction-producer-ms/internal/queue"
	"github.com/felipemacedo1/transaction-producer-ms/internal/service"
)

func main() {
	const filePath = "/app/input/input-data.csv"
	rabbitMQ, err := queue.NewRabbitMQ("amqp://guest:guest@rabbitmq:5672/", "transaction_queue")
	if err != nil {
		log.Fatal("error connecting to RabbitMQ: ", err)
	}
	defer rabbitMQ.Close()

	csvService := service.NewCSVService(rabbitMQ)
	err = csvService.ProcessCSVFile(filePath)
	if err != nil {
		log.Fatal("error processing csv file: ", err)
	}
}
