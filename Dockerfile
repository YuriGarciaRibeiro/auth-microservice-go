# ------------------------------------
# Build Stage
# ------------------------------------
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app

# Cache de deps
COPY go.mod go.sum ./
RUN go mod download

# Código
COPY . .

# Args para popular /buildz
ARG VERSION=dev
ARG COMMIT=none
ARG BUILDTIME=unknown

# Compila binário estático e “slim”
# Observação: ajuste o import abaixo para o caminho do seu package version
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "\
      -s -w \
      -X 'github.com/YuriGarciaRibeiro/auth-microservice-go/internal/version.Version=${VERSION}' \
      -X 'github.com/YuriGarciaRibeiro/auth-microservice-go/internal/version.Commit=${COMMIT}' \
      -X 'github.com/YuriGarciaRibeiro/auth-microservice-go/internal/version.BuildTime=${BUILDTIME}'" \
    -o /app/auth-service ./cmd/auth-service

# ------------------------------------
# Runtime Stage
# ------------------------------------
FROM alpine:3.20

# Certificados e timezone
RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

WORKDIR /app
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/auth-service .
COPY --from=builder /app/docs ./docs

# Segurança: usuário não-root
RUN addgroup -S app && adduser -S app -G app
USER app

ENV TZ=UTC \
    OTEL_SERVICE_NAME=auth-service \
    PORT=8080

EXPOSE 8080

# Healthcheck (usa /healthz que acabamos de criar)
HEALTHCHECK --interval=30s --timeout=2s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:${PORT}/healthz || exit 1

CMD ["./auth-service"]
