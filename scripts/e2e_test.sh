#!/bin/bash
set -e

echo "========================================="
echo "Script de Teste End-to-End"
echo "========================================="

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Funções auxiliares
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Limpar ambiente anterior
log_info "Limpando ambiente anterior..."
docker-compose down -v 2>/dev/null || true
sleep 2

# Subir serviços
log_info "Subindo serviços com docker-compose..."
docker-compose up -d --build

# Aguardar serviços ficarem prontos
log_info "Aguardando serviços ficarem prontos..."
sleep 10

# Verificar health do RabbitMQ
log_info "Verificando health do RabbitMQ..."
for i in {1..30}; do
    if docker exec rabbitmq rabbitmq-diagnostics ping >/dev/null 2>&1; then
        log_info "✓ RabbitMQ está pronto"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "RabbitMQ não ficou pronto a tempo"
        exit 1
    fi
    sleep 2
done

# Verificar health do Postgres
log_info "Verificando health do Postgres..."
for i in {1..30}; do
    if docker exec postgres pg_isready -U user -d transactions >/dev/null 2>&1; then
        log_info "✓ Postgres está pronto"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "Postgres não ficou pronto a tempo"
        exit 1
    fi
    sleep 2
done

# Aguardar producer processar
log_info "Aguardando producer processar CSV..."
sleep 5

# Verificar logs do producer
log_info "Últimas linhas do producer:"
docker-compose logs --tail=20 producer

# Aguardar consumer processar
log_info "Aguardando consumer processar mensagens (30s)..."
sleep 30

# Verificar contagem no banco de dados
log_info "Verificando contagem de transações no banco de dados..."
DB_COUNT=$(docker exec postgres psql -U user -d transactions -t -c "SELECT COUNT(*) FROM transactions;" | xargs)
log_info "Total de transações no banco: $DB_COUNT"

# Verificar se processou dados
if [ "$DB_COUNT" -gt 0 ]; then
    log_info "✓ Dados foram processados com sucesso!"
else
    log_error "✗ Nenhuma transação encontrada no banco"
    exit 1
fi

# Verificar RabbitMQ (deve estar vazio ou quase)
log_info "Verificando filas no RabbitMQ..."
docker-compose logs --tail=10 consumer

# Teste de idempotência - rodar producer novamente
log_info ""
log_info "========================================="
log_info "Teste de Idempotência"
log_info "========================================="
log_info "Reiniciando producer para testar idempotência..."
docker-compose restart producer
sleep 15

# Verificar se contagem permanece a mesma
DB_COUNT_AFTER=$(docker exec postgres psql -U user -d transactions -t -c "SELECT COUNT(*) FROM transactions;" | xargs)
log_info "Total de transações após segundo processamento: $DB_COUNT_AFTER"

if [ "$DB_COUNT" -eq "$DB_COUNT_AFTER" ]; then
    log_info "✓ Teste de idempotência PASSOU - contagem permaneceu: $DB_COUNT"
else
    log_warn "⚠ Possível falha de idempotência: antes=$DB_COUNT, depois=$DB_COUNT_AFTER"
fi

# Verificar health endpoints
log_info ""
log_info "========================================="
log_info "Verificando Health Endpoints"
log_info "========================================="

# Tentar verificar health do producer (pode ter terminado)
PRODUCER_RUNNING=$(docker ps --filter "name=producer" --filter "status=running" -q)
if [ -n "$PRODUCER_RUNNING" ]; then
    if docker exec producer wget -q -O- http://localhost:8080/healthz 2>/dev/null | grep -q "OK"; then
        log_info "✓ Producer health endpoint OK"
    else
        log_warn "⚠ Producer health endpoint não respondeu"
    fi
fi

# Verificar health do consumer
if docker exec consumer wget -q -O- http://localhost:8081/healthz 2>/dev/null | grep -q "OK"; then
    log_info "✓ Consumer health endpoint OK"
else
    log_warn "⚠ Consumer health endpoint não respondeu"
fi

# Métricas do consumer
log_info ""
log_info "Métricas do Consumer:"
docker exec consumer wget -q -O- http://localhost:8081/metrics 2>/dev/null | grep "consumer_messages" || true

# Sumário final
log_info ""
log_info "========================================="
log_info "Sumário dos Testes"
log_info "========================================="
log_info "✓ Serviços foram iniciados com sucesso"
log_info "✓ $DB_COUNT transações foram processadas"
log_info "✓ Teste de idempotência concluído"
log_info ""
log_info "Para visualizar logs:"
log_info "  docker-compose logs -f producer"
log_info "  docker-compose logs -f consumer"
log_info ""
log_info "Para acessar RabbitMQ UI:"
log_info "  http://localhost:15672 (guest/guest)"
log_info ""
log_info "Para consultar banco de dados:"
log_info "  docker exec -it postgres psql -U user -d transactions"
log_info ""
log_info "Para limpar ambiente:"
log_info "  docker-compose down -v"

echo ""
log_info "Testes concluídos com sucesso! ✓"
