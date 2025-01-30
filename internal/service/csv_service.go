package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	_ "strconv"
	"sync"
	"time"
	_ "time"

	"github.com/growthfolio/transaction-producer-ms/internal/dto"
	"github.com/growthfolio/transaction-producer-ms/internal/logging"
	"github.com/growthfolio/transaction-producer-ms/internal/queue"
)

type CSVService struct {
	queue queue.Queue
}

func NewCSVService(q queue.Queue) *CSVService {
	return &CSVService{
		queue: q,
	}
}

func (s *CSVService) parseTransaction(record []string) (*dto.Transaction, error) {
	if len(record) < 7 {
		return nil, fmt.Errorf("invalid record length: expected 7 fields, got %d", len(record))
	}

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

func (s *CSVService) ProcessCSVFile(filePath string) error {
	logging.Logger.WithField("filePath", filePath).Info("Starting CSV processing")

	file, err := os.Open(filePath)
	if err != nil {
		logging.Logger.WithError(err).Error("Could not open file")
		return fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var wg sync.WaitGroup
	ch := make(chan *dto.Transaction, 100)
	errCh := make(chan error, 10)

	// Worker para enviar mensagens ao RabbitMQ
	go func() {
		defer close(ch)
		for transaction := range ch {
			jsonData, err := json.Marshal(transaction)
			if err != nil {
				logging.Logger.WithError(err).Warn("JSON could not be created")
				continue
			}

			if err := s.queue.Publish(jsonData); err != nil {
				logging.Logger.WithError(err).
					WithField("transactionId", transaction.TransactionID).
					Error("Transaction could not be sent")
			} else {
				logging.Logger.WithField("transactionId", transaction.TransactionID).
					Info("Transaction sent successfully")
			}
		}
	}()

	// Processamento das linhas do CSV
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logging.Logger.WithError(err).Warn("CSV line could not be read")
			errCh <- fmt.Errorf("csv could not be read: %v", err)
			continue
		}

		wg.Add(1)
		go func(record []string) {
			defer wg.Done()

			transaction, err := s.parseTransaction(record) // Corrigido: chamada correta
			if err != nil {
				logging.Logger.WithError(err).Warn("Transaction could not be parsed")
				errCh <- fmt.Errorf("transaction could not be parsed: %v", err)
				return
			}
			ch <- transaction
		}(record)
	}

	wg.Wait()
	close(errCh) // Fechar errCh somente após todas as goroutines terminarem

	// Exibir erros ocorridos durante o processamento
	for err := range errCh {
		logging.Logger.WithError(err).Warn("Processing error")
	}

	logging.Logger.Info("CSV processing completed")
	return nil
}
