package main

import (
	"log"

	"github.com/felipemacedo1/transaction-producer-ms/internal/handler"
)

func main() {
	const filePath = "/app/input/input-data.csv"
	err := handler.ProcessCSVFile(filePath)
	if err != nil {
		log.Fatal("error processing csv file: ", err)
	}
}
