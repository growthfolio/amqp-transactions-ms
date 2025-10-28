package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Transaction struct {
	ID           string  `json:"id" gorm:"primaryKey"`
	Date         string  `json:"date"`
	Document     string  `json:"document"`
	Name         string  `json:"name"`
	Age          int     `json:"age"`
	Amount       float64 `json:"amount"`
	Installments int     `json:"installments"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var (
	db             *gorm.DB
	processedCount int64
	duplicateCount int64
	errorCount     int64
	isHealthy      int32 = 1
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&isHealthy) == 1 {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "UNHEALTHY")
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP consumer_messages_processed_total Total de mensagens processadas\n")
	fmt.Fprintf(w, "# TYPE consumer_messages_processed_total counter\n")
	fmt.Fprintf(w, "consumer_messages_processed_total %d\n", atomic.LoadInt64(&processedCount))

	fmt.Fprintf(w, "# HELP consumer_messages_duplicate_total Total de mensagens duplicadas\n")
	fmt.Fprintf(w, "# TYPE consumer_messages_duplicate_total counter\n")
	fmt.Fprintf(w, "consumer_messages_duplicate_total %d\n", atomic.LoadInt64(&duplicateCount))

	fmt.Fprintf(w, "# HELP consumer_messages_error_total Total de erros\n")
	fmt.Fprintf(w, "# TYPE consumer_messages_error_total counter\n")
	fmt.Fprintf(w, "consumer_messages_error_total %d\n", atomic.LoadInt64(&errorCount))
}

func main() {
	// Configuração via env vars
	dsn := getEnv("POSTGRES_DSN", "host=postgres user=user password=password dbname=transactions port=5432 sslmode=disable")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	queueName := getEnv("QUEUE_NAME", "transactions_queue")
	prefetch := getEnvInt("PREFETCH", 100)
	workers := getEnvInt("WORKERS", 5)
	batchSize := getEnvInt("BATCH_SIZE", 100)
	httpPort := getEnv("HTTP_PORT", "8081")

	log.Printf("Iniciando consumer: workers=%d, prefetch=%d, batch=%d, queue=%s", workers, prefetch, batchSize, queueName)

	// Iniciar servidor HTTP para health e metrics
	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/metrics", metricsHandler)
	go func() {
		log.Printf("HTTP server listening on :%s", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Conectar ao banco de dados
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar no banco:", err)
	}

	// AutoMigrate com índice único no ID para garantir idempotência
	if err := db.AutoMigrate(&Transaction{}); err != nil {
		log.Fatal("AutoMigrate falhou:", err)
	}
	log.Println("✓ AutoMigrate concluído")

	// Conectar ao RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal("Falha ao conectar ao RabbitMQ:", err)
	}
	defer conn.Close()
	log.Println("✓ Conectado ao RabbitMQ")

	// Criar canal principal para declarar fila
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Falha ao abrir um canal:", err)
	}

	// Declarar fila durável (sem DLQ para compatibilidade com producer)
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatal("Falha ao declarar fila:", err)
	}

	ch.Close()
	log.Println("✓ Fila declarada") // Worker pool para processamento concorrente
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Cada worker tem seu próprio canal
			ch, err := conn.Channel()
			if err != nil {
				log.Printf("worker %d: erro ao abrir canal: %v", workerID, err)
				atomic.StoreInt32(&isHealthy, 0)
				return
			}
			defer ch.Close()

			// Configurar QoS para limitar mensagens in-flight
			if err := ch.Qos(prefetch, 0, false); err != nil {
				log.Printf("worker %d: erro ao configurar QoS: %v", workerID, err)
				atomic.StoreInt32(&isHealthy, 0)
				return
			}

			// Consumir mensagens com autoAck=false
			msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
			if err != nil {
				log.Printf("worker %d: erro ao registrar consumidor: %v", workerID, err)
				atomic.StoreInt32(&isHealthy, 0)
				return
			}

			log.Printf("worker %d: aguardando mensagens...", workerID)

			batch := make([]Transaction, 0, batchSize)
			deliveries := make([]amqp.Delivery, 0, batchSize)
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case d, ok := <-msgs:
					if !ok {
						// Canal fechado
						return
					}

					var t Transaction
					if err := json.Unmarshal(d.Body, &t); err != nil {
						log.Printf("worker %d: erro ao desserializar: %v", workerID, err)
						// Mensagem malformada -> descartar (sem DLQ por enquanto)
						d.Ack(false) // Ack para remover da fila
						atomic.AddInt64(&errorCount, 1)
						continue
					}

					batch = append(batch, t)
					deliveries = append(deliveries, d)

					// Processar quando atingir tamanho do batch
					if len(batch) >= batchSize {
						processBatch(workerID, batch, deliveries)
						batch = batch[:0]
						deliveries = deliveries[:0]
					}

				case <-ticker.C:
					// Processar batch parcial a cada 2 segundos
					if len(batch) > 0 {
						processBatch(workerID, batch, deliveries)
						batch = batch[:0]
						deliveries = deliveries[:0]
					}
				}
			}
		}(i)
	}

	// Goroutine para exibir estatísticas periodicamente
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			log.Printf("📊 Estatísticas: processadas=%d, duplicadas=%d, erros=%d",
				atomic.LoadInt64(&processedCount),
				atomic.LoadInt64(&duplicateCount),
				atomic.LoadInt64(&errorCount))
		}
	}()

	fmt.Println("✓ Consumer iniciado e aguardando mensagens...")
	wg.Wait()
}

func processBatch(workerID int, batch []Transaction, deliveries []amqp.Delivery) {
	if len(batch) == 0 {
		return
	}

	// Inserção em batch com idempotência (OnConflict DoNothing)
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).CreateInBatches(batch, len(batch))

	if result.Error != nil {
		log.Printf("worker %d: erro ao salvar batch (size=%d): %v", workerID, len(batch), result.Error)

		// Em caso de erro, tentar inserir individualmente para identificar problemas
		for i, t := range batch {
			if err := saveSingle(&t); err != nil {
				log.Printf("worker %d: erro ao salvar id=%s: %v", workerID, t.ID, err)
				// Requeue em caso de erro transitório
				deliveries[i].Nack(false, true)
				atomic.AddInt64(&errorCount, 1)
			} else {
				deliveries[i].Ack(false)
				atomic.AddInt64(&processedCount, 1)
			}
		}
		return
	}

	// Calcular duplicatas (rowsAffected < batch size)
	inserted := result.RowsAffected
	duplicates := int64(len(batch)) - inserted

	atomic.AddInt64(&processedCount, inserted)
	atomic.AddInt64(&duplicateCount, duplicates)

	// Ack todas as mensagens do batch
	for _, d := range deliveries {
		d.Ack(false)
	}

	if len(batch) > 10 {
		log.Printf("worker %d: batch processado (size=%d, inseridas=%d, duplicadas=%d)",
			workerID, len(batch), inserted, duplicates)
	}
}

func saveSingle(t *Transaction) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(t).Error
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
