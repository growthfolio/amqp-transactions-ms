package main

import (
	"github.com/growthfolio/transaction-producer-ms/internal/logging"
	"github.com/growthfolio/transaction-producer-ms/internal/queue"
	"github.com/growthfolio/transaction-producer-ms/internal/service"
)

func main() {
	const filePath = "../"
	log := logging.Logger // Usando o logger global

	log.Info("Starting application...")

	rabbitMQ, err := queue.NewRabbitMQ("amqp://guest:guest@rabbitmq:5672/", "transaction_queue")
	if err != nil {
		log.WithError(err).Fatal("Error connecting to RabbitMQ")
	}
	defer rabbitMQ.Close()

	log.Info("Connected to RabbitMQ successfully")

	csvService := service.NewCSVService(rabbitMQ)
	err = csvService.ProcessCSVFile(filePath)
	if err != nil {
		log.WithError(err).Fatal("Error processing CSV file")
	}

	log.Info("CSV processing completed successfully")
}
