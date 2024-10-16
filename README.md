
# Transaction Producer Microservice

This is a Go-based microservice responsible for processing transaction data from CSV files, converting the data into JSON format, and sending the data to a message broker (RabbitMQ) for further processing by other microservices.

## Features

- Reads transaction data from a CSV file located in the `inputs` directory.
- Converts data fields to the appropriate types (e.g., `time.Time` for dates, `float64` for amounts, and `int` for installments).
- Handles errors during data parsing (e.g., skipping rows with invalid data).
- Sends the processed transaction data to RabbitMQ in JSON format.
- **Implements SOLID principles** with separation of concerns between file processing, message queuing, and business logic.
- Uses **dependency injection** for the message broker, allowing easy replacement of RabbitMQ with other message brokers in the future.

## Project Structure

```bash
transaction-producer-ms/
├── cmd/
│   └── main.go                   # Entry point of the microservice
├── internal/
│   ├── dto/                      # Defines the Transaction struct
│   ├── handler/                  # (Deprecated) CSV processing logic
│   ├── queue/                    # Message queue handling (RabbitMQ implementation)
│   │   ├── queue.go              # Queue interface for abstraction
│   │   └── rabbitmq.go           # RabbitMQ implementation of the Queue interface
│   └── service/                  # Business logic for processing CSV files
│       └── csv_service.go        # CSV processing and data handling
└── input/                        
    └── input-data.csv            # Example transaction CSV file
```

## Prerequisites

- Go 1.16+ installed
- RabbitMQ instance running

## Setup Instructions

1. Clone the repository:

```bash
git clone https://github.com/felipemacedo1/transaction-producer-ms.git
cd transaction-producer-ms
```

2. Install dependencies:

```bash
go mod tidy
```

3. Start RabbitMQ locally or set the `RABBITMQ_URL` environment variable to point to a remote RabbitMQ instance.

4. Run the microservice:

```bash
go run cmd/main.go
```

5. Place a CSV file in the `input/` directory for processing.

## Usage

The microservice automatically reads from the CSV file placed in the `input/` directory and processes the data. It then sends the data to RabbitMQ for consumption by other microservices. You can customize the RabbitMQ connection settings and CSV input file path via environment variables.

## Configuration

Environment variables:

- `RABBITMQ_URL`: URL to connect to RabbitMQ (default: `amqp://guest:guest@localhost:5672/`).
- `CSV_FILE_PATH`: Path to the input CSV file (default: `input/input-data.csv`).

## Improvements and Changes

- The **`handler` package** is now deprecated and replaced with a more modular approach.
- The **`queue` package** introduces an abstraction layer for handling message queues, with RabbitMQ as the default implementation.
- The **`service` package** handles the business logic for CSV processing, following the **Single Responsibility Principle (SRP)**.

<!-- ## Testing

To run tests, use the following command:

```bash
go test ./...
``` -->

## Contributing

Feel free to submit issues, fork the repository, and send pull requests. For major changes, please open an issue first to discuss what you would like to change.

<!-- ## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. -->
