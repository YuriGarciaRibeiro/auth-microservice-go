# ------------------------------------
# Build Stage
# ------------------------------------
FROM golang:1.21-alpine AS builder

# Instala dependências mínimas
RUN apk add --no-cache git

WORKDIR /app

# Copia os arquivos de dependência primeiro (cache)
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Compila o binário
RUN go build -o auth-service ./cmd/auth-service

# ------------------------------------
# Runtime Stage
# ------------------------------------
FROM alpine:latest

WORKDIR /app

# Copia o binário compilado
COPY --from=builder /app/auth-service .

# Copia a pasta com os arquivos do Swagger
COPY --from=builder /app/docs ./docs

# Expõe a porta (ajuste se necessário)
EXPOSE 8080

# Comando de execução
CMD ["./auth-service"]
