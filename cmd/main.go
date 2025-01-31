package main

import (
    "os"
	"github.com/growthfolio/transaction-producer-ms/internal/logging"
	"github.com/growthfolio/transaction-producer-ms/internal/queue"
	"github.com/growthfolio/transaction-producer-ms/internal/service"
)

func main() {
    log := logging.Logger
    log.Info("Starting application...")

    rabbitMQURL := os.Getenv("RABBITMQ_URL")
    if rabbitMQURL == "" {
        rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
    }

    rabbitMQ, err := queue.NewRabbitMQ(rabbitMQURL, "transaction_queue")
    if err != nil {
        log.WithError(err).Fatal("Fatal error connecting to RabbitMQ")
    }
    defer rabbitMQ.Close()

    log.Info("Connected to RabbitMQ successfully")

    csvService := service.NewCSVService(rabbitMQ)
    err = csvService.ProcessCSVFile("./input/input-data.csv")
    if err != nil {
        log.WithError(err).Fatal("Error processing CSV file")
    }

    log.Info("CSV processing completed successfully")
}
