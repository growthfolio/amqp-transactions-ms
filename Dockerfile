FROM golang:1.18-alpine AS builder

COPY input/ /app/input/

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main.go

# Imagem final diretamente com Alpine
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 8080
CMD ["./app"]
