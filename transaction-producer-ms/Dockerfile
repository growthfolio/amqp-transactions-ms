# Utilizando a imagem oficial do Go
FROM golang:1.22

# Definir diretório de trabalho dentro do container
WORKDIR /app

# Copiar os arquivos necessários para o diretório de trabalho
COPY go.mod go.sum ./
RUN go mod download

# Copiar o código da aplicação
COPY . .

# Compilar a aplicação
RUN go build -o producer

# Definir a entrada do container
CMD ["./producer"]
