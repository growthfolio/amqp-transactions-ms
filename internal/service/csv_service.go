package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/felipemacedo1/transaction-producer-ms/internal/dto"
	"github.com/felipemacedo1/transaction-producer-ms/internal/queue"
)

type CSVService struct {
	queue queue.Queue
}

func NewCSVService(q queue.Queue) *CSVService {
	return &CSVService{queue: q}
}

func (s *CSVService) ProcessCSVFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("csv could not be read: %v", err)
			continue
		}

		transaction, err := s.parseTransaction(record)
		if err != nil {
			log.Printf("transaction could not be parsed: %v", err)
			continue
		}

		jsonData, err := json.Marshal(transaction)
		if err != nil {
			log.Printf("JSON could not be created: %v", err)
			continue
		}

		if err := s.queue.Publish(jsonData); err != nil {
			log.Printf("transaction could not be sent: %v", err)
		}
	}

	return nil
}

func (s *CSVService) parseTransaction(record []string) (*dto.Transaction, error) {
	age, err := strconv.Atoi(record[4])
	if err != nil {
		return nil, fmt.Errorf("age could not be converted: %v", err)
	}

	amount, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return nil, fmt.Errorf("amount could not be converted: %v", err)
	}

	installments, err := strconv.Atoi(record[6])
	if err != nil {
		return nil, fmt.Errorf("installments could not be converted: %v", err)
	}

	date, err := time.Parse(time.RFC3339, record[1])
	if err != nil {
		return nil, fmt.Errorf("date could not be converted: %v", err)
	}

	return &dto.Transaction{
		TransactionID: record[0],
		Date:          date,
		ClientID:      record[2],
		Name:          record[3],
		Age:           age,
		Amount:        amount,
		Installments:  installments,
	}, nil
}
