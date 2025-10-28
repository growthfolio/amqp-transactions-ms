package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
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

var (
	publishedCount int64
	failedCount    int64
	isHealthy      int32 = 1
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
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

// findFirstCSV localiza o primeiro arquivo CSV no diretório (grep curto)
func findFirstCSV(dirPath string) (string, error) {
	d, err := os.Open(dirPath)
	if err != nil {
		return "", err
	}
	defer d.Close()

	names, err := d.Readdirnames(0)
	if err != nil {
		return "", err
	}

	for _, n := range names {
		if len(n) > 4 && (n[len(n)-4:] == ".csv" || n[len(n)-4:] == ".CSV") {
			return dirPath + "/" + n, nil
		}
	}
	return "", fmt.Errorf("nenhum CSV encontrado em %s", dirPath)
}

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
	fmt.Fprintf(w, "# HELP producer_messages_published_total Total de mensagens publicadas\n")
	fmt.Fprintf(w, "# TYPE producer_messages_published_total counter\n")
	fmt.Fprintf(w, "producer_messages_published_total %d\n", atomic.LoadInt64(&publishedCount))

	fmt.Fprintf(w, "# HELP producer_messages_failed_total Total de mensagens que falharam\n")
	fmt.Fprintf(w, "# TYPE producer_messages_failed_total counter\n")
	fmt.Fprintf(w, "producer_messages_failed_total %d\n", atomic.LoadInt64(&failedCount))
}

func main() {
	// Configuração via env vars
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	inputPath := getEnv("INPUT_PATH", "/app/data")
	queueName := getEnv("QUEUE_NAME", "transactions_queue")
	workers := getEnvInt("WORKERS", 4)
	httpPort := getEnv("HTTP_PORT", "8080")

	log.Printf("Iniciando producer: workers=%d, queue=%s, input=%s", workers, queueName, inputPath)

	// Iniciar servidor HTTP para health e metrics
	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/metrics", metricsHandler)
	go func() {
		log.Printf("HTTP server listening on :%s", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Localizar primeiro CSV (grep curto - não lê conteúdo)
	csvFile, err := findFirstCSV(inputPath)
	failOnError(err, "localizar CSV")
	log.Printf("Processando arquivo: %s", csvFile)

	// Conectar ao RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	failOnError(err, "Falha ao conectar ao RabbitMQ")
	defer conn.Close()

	// Worker pool - cada worker com seu próprio channel
	jobs := make(chan []string, 500)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			ch, err := conn.Channel()
			if err != nil {
				log.Printf("worker %d: erro ao abrir canal: %v", workerID, err)
				atomic.StoreInt32(&isHealthy, 0)
				return
			}
			defer ch.Close()

			// Declarar fila durável
			_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
			if err != nil {
				log.Printf("worker %d: erro ao declarar fila: %v", workerID, err)
				atomic.StoreInt32(&isHealthy, 0)
				return
			}

			// Habilitar publisher confirms para garantir entrega
			if err := ch.Confirm(false); err != nil {
				log.Printf("worker %d: erro ao habilitar confirms: %v", workerID, err)
				atomic.StoreInt32(&isHealthy, 0)
				return
			}
			confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 100))

			for rec := range jobs {
				if len(rec) < 7 {
					log.Printf("worker %d: linha inválida (campos insuficientes)", workerID)
					atomic.AddInt64(&failedCount, 1)
					continue
				}

				t := Transaction{
					ID:           rec[0],
					Date:         rec[1],
					Document:     rec[2],
					Name:         rec[3],
					Age:          atoi(rec[4]),
					Amount:       atof(rec[5]),
					Installments: atoi(rec[6]),
				}

				body, _ := json.Marshal(t)
				err = ch.Publish("", queueName, false, false, amqp.Publishing{
					DeliveryMode: amqp.Persistent, // Mensagem persistente
					ContentType:  "application/json",
					MessageId:    t.ID,
					Body:         body,
					Timestamp:    time.Now(),
				})

				if err != nil {
					log.Printf("worker %d: erro ao publicar id=%s: %v", workerID, t.ID, err)
					atomic.AddInt64(&failedCount, 1)
					continue
				}

				// Aguardar confirmação
				select {
				case confirm := <-confirms:
					if confirm.Ack {
						atomic.AddInt64(&publishedCount, 1)
					} else {
						log.Printf("worker %d: nack recebido para id=%s", workerID, t.ID)
						atomic.AddInt64(&failedCount, 1)
					}
				case <-time.After(5 * time.Second):
					log.Printf("worker %d: timeout aguardando confirm para id=%s", workerID, t.ID)
					atomic.AddInt64(&failedCount, 1)
				}
			}
		}(i)
	}

	// Abrir CSV em streaming (linha por linha)
	startTime := time.Now()
	f, err := os.Open(csvFile)
	failOnError(err, "Erro ao abrir o CSV")
	defer f.Close()

	reader := csv.NewReader(bufio.NewReader(f))
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // Permitir número variável de campos

	lineCount := 0
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("erro ao ler linha %d: %v", lineCount, err)
			continue
		}

		lineCount++
		jobs <- rec

		// Log de progresso a cada 1000 linhas
		if lineCount%1000 == 0 {
			log.Printf("Processadas %d linhas...", lineCount)
		}
	}

	close(jobs)
	wg.Wait()

	elapsed := time.Since(startTime)
	log.Printf("✓ Publicação concluída em %v", elapsed)
	log.Printf("  Total de linhas: %d", lineCount)
	log.Printf("  Publicadas: %d", atomic.LoadInt64(&publishedCount))
	log.Printf("  Falhas: %d", atomic.LoadInt64(&failedCount))

	// Manter servidor HTTP rodando
	select {}
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
