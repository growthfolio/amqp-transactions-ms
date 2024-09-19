package main

import (
	"log"

	"github.com/felipemacedo1/transaction-producer-ms/internal/handler"
)

func main() {
	err := handler.ProcessCSVFile("input/input-data.csv")
	if err != nil {
		log.Fatal("error processing csv file: ", err)
	}
}
