package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type Transaction struct {
	ID            string  `json:"id"`
	Date          string  `json:"date"`
	Document      string  `json:"document"`
	Name          string  `json:"name"`
	Age           int     `json:"age"`
	Amount        float64 `json:"amount"`
	Installments  int     `json:"installments"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Falha ao conectar ao RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Falha ao abrir um canal")
	defer ch.Close()

	q, err := ch.QueueDeclare("transactions_queue", false, false, false, false, nil)
	failOnError(err, "Falha ao declarar a fila")

	file, err := os.Open("/app/input/input-data.csv")
	failOnError(err, "Erro ao abrir o arquivo CSV")
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	records, err := reader.ReadAll()
	failOnError(err, "Erro ao ler o CSV")

	for _, record := range records {
		transaction := Transaction{
			ID:           record[0],
			Date:         record[1],
			Document:     record[2],
			Name:         record[3],
			Age:          atoi(record[4]),
			Amount:       atof(record[5]),
			Installments: atoi(record[6]),
		}

		body, _ := json.Marshal(transaction)
		err = ch.Publish("", q.Name, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

		failOnError(err, "Erro ao publicar a mensagem")
		fmt.Println("Mensagem enviada:", transaction)
		time.Sleep(500 * time.Millisecond)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func atof(s string) float64 {
	n, _ := strconv.ParseFloat(s, 64)
	return n
}
