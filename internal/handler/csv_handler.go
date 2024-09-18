package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go-csv-to-json/internal/model"
	"log"
	"os"

	"github.com/streadway/amqp"
)

func ProcessCSVFile(filePath string) error {
	// Conecta ao RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("não foi possível conectar ao RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Cria um canal
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("não foi possível abrir o canal: %v", err)
	}
	defer ch.Close()

	// Declara a fila
	queue, err := ch.QueueDeclare(
		"transaction_queue", // nome da fila
		true,                // durável
		false,               // auto-delete
		false,               // exclusive
		false,               // no-wait
		nil,                 // argumentos adicionais
	)
	if err != nil {
		return fmt.Errorf("não foi possível declarar a fila: %v", err)
	}

	// Abre o arquivo CSV
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("não foi possível abrir o arquivo: %v", err)
	}
	defer file.Close()

	// Lê o arquivo CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("erro ao ler o arquivo CSV: %v", err)
	}

	// Percorre cada linha do arquivo CSV e envia como JSON para o RabbitMQ
	for _, record := range records {
		transaction := model.Transaction{
			ID:           record[0],
			Date:         record[1],
			Document:     record[2],
			Name:         record[3],
			Age:          record[4],
			Amount:       record[5],
			Installments: record[6],
		}

		// Converte a transação para JSON
		jsonData, err := json.Marshal(transaction)
		if err != nil {
			log.Printf("Erro ao converter para JSON: %v", err)
			continue
		}

		// Publica a mensagem no RabbitMQ
		err = ch.Publish(
			"",         // exchange
			queue.Name, // routing key
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        jsonData,
			},
		)
		if err != nil {
			log.Printf("Erro ao publicar mensagem no RabbitMQ: %v", err)
		}
	}

	return nil
}
