package main

import (
	"log"
	"transaction-producer-ms/internal/handler"
)

func main() {
	// Inicia o processamento do arquivo CSV
	err := handler.ProcessCSVFile("input-data.csv")
	if err != nil {
		log.Fatal("Erro ao processar o arquivo CSV:", err)
	}
}
