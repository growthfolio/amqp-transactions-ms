package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Transaction struct {
	ID           string  `json:"id"`
	Date         string  `json:"date"`
	Document     string  `json:"document"`
	Name         string  `json:"name"`
	Age          int     `json:"age"`
	Amount       float64 `json:"amount"`
	Installments int     `json:"installments"`
}

var db *gorm.DB

func main() {
	var err error
	dsn := "host=postgres user=user password=password dbname=transactions port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar no banco:", err)
	}

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatal("Falha ao conectar ao RabbitMQ", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Falha ao abrir um canal", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume("transactions_queue", "", true, false, false, false, nil)
	if err != nil {
		log.Fatal("Falha ao registrar o consumidor", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var transaction Transaction
			json.Unmarshal(d.Body, &transaction)

			fmt.Println("Processando transação:", transaction)
			// Aqui entra a lógica para salvar no banco de dados.
		}
	}()

	fmt.Println("Aguardando mensagens...")
	<-forever
}
