# 🔄 AMQP Transactions Microservices

## 🎯 Objetivo de Aprendizado
Projeto de microserviços desenvolvido para estudar **arquitetura distribuída** com **RabbitMQ** e **AMQP**. Implementa padrão **Producer-Consumer** para processamento assíncrono de transações financeiras, aplicando princípios **SOLID** e **Clean Architecture**.

## 🛠️ Tecnologias Utilizadas
- **Linguagem:** Go
- **Message Broker:** RabbitMQ (AMQP)
- **Arquitetura:** Microserviços
- **Padrões:** Producer-Consumer, Dependency Injection
- **Containerização:** Docker Compose
- **Processamento:** CSV para JSON

## 🚀 Demonstração
```bash
# Estrutura do sistema
Producer MS → RabbitMQ → Consumer MS
     ↓           ↓           ↓
CSV Input → JSON Queue → Processing
```

## 📁 Estrutura do Projeto
```
amqp-transactions-ms/
├── transaction-producer-ms/    # Microserviço produtor
│   ├── cmd/main.go            # Entry point
│   ├── internal/
│   │   ├── dto/               # Transaction struct
│   │   ├── queue/             # RabbitMQ abstraction
│   │   └── service/           # Business logic
│   └── input/                 # CSV files
├── transaction-consumer-ms/    # Microserviço consumidor
└── docker-compose.yml         # Orquestração dos serviços
```

## 💡 Principais Aprendizados

### 🏗️ Arquitetura de Microserviços
- **Separação de responsabilidades:** Producer e Consumer independentes
- **Comunicação assíncrona:** Desacoplamento via message broker
- **Escalabilidade:** Cada serviço pode escalar independentemente
- **Resiliência:** Falhas isoladas não afetam todo o sistema

### 📨 Message Broker (RabbitMQ)
- **AMQP Protocol:** Advanced Message Queuing Protocol
- **Queue abstraction:** Interface para diferentes brokers
- **Dependency injection:** Facilita testes e manutenção
- **Error handling:** Tratamento de falhas na comunicação

### 🔄 Processamento de Dados
- **CSV parsing:** Leitura e validação de dados
- **Type conversion:** Conversão para tipos apropriados
- **JSON serialization:** Formato padrão para comunicação
- **Data validation:** Tratamento de dados inválidos

## 🧠 Conceitos Técnicos Estudados

### 1. **SOLID Principles**
```go
// Single Responsibility - Cada serviço tem uma responsabilidade
type CSVService struct {
    queue Queue
}

// Dependency Inversion - Abstração do message broker
type Queue interface {
    Publish(data []byte) error
}
```

### 2. **Clean Architecture**
```go
// Camadas bem definidas
internal/
├── dto/        # Entities
├── service/    # Use Cases
└── queue/      # Infrastructure
```

### 3. **Producer-Consumer Pattern**
```go
// Producer: Processa CSV e envia para queue
func (s *CSVService) ProcessFile(filepath string) error {
    // Lê CSV, converte e publica
}

// Consumer: Recebe da queue e processa
func (c *Consumer) ProcessMessage(msg []byte) error {
    // Deserializa e processa transação
}
```

## 🚧 Desafios Enfrentados
1. **Sincronização:** Coordenação entre producer e consumer
2. **Error handling:** Tratamento de falhas em sistemas distribuídos
3. **Data consistency:** Garantir integridade dos dados
4. **Performance:** Otimização do throughput de mensagens
5. **Monitoring:** Observabilidade em arquitetura distribuída

## 📚 Recursos Utilizados
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [Go AMQP Library](https://github.com/streadway/amqp)
- [Microservices Patterns](https://microservices.io/patterns/)
- [Clean Architecture in Go](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## 📈 Próximos Passos
- [ ] Implementar Dead Letter Queue
- [ ] Adicionar métricas e monitoring
- [ ] Implementar circuit breaker
- [ ] Adicionar testes de integração
- [ ] Implementar retry policies
- [ ] Adicionar tracing distribuído

## 🔗 Projetos Relacionados
- [Go Antifraud MS](../go-antifraud-ms/) - Microserviço de detecção de fraude
- [Go PriceGuard API](../go-priceguard-api/) - API com Clean Architecture
- [Transaction Consumer MS](../transaction-consumer-ms/) - Consumer específico

---

**Desenvolvido por:** Felipe Macedo  
**Contato:** contato.dev.macedo@gmail.com  
**GitHub:** [FelipeMacedo](https://github.com/felipemacedo1)  
**LinkedIn:** [felipemacedo1](https://linkedin.com/in/felipemacedo1)

> 💡 **Reflexão:** Este projeto foi fundamental para compreender arquiteturas distribuídas e comunicação assíncrona. A implementação de microserviços com RabbitMQ proporcionou experiência prática em sistemas escaláveis e resilientes.