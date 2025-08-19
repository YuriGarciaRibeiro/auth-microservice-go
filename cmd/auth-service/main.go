// @title Auth Microservice API
// @version 1.0
// @description Serviço de autenticação centralizado com JWT
// @contact.name Yuri Garcia Ribeiro
// @contact.url https://github.com/YuriGarciaRibeiro/auth-microservice-go
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/config"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/cache"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/logger"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/metrics"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/trace"
	internalhttp "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/transport/http"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load .env file if available
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize logger
	l, err := logger.Init()
	if err != nil {
		panic(err)
	}
	defer l.Sync()

	// Initialize tracing
	shutdown, err := trace.Init()
	if err != nil {
		logger.L().Fatal("otel init error", zap.Error(err))
	}
	defer shutdown(context.Background())

	// Initialize metrics
	metrics.MustRegister()

	sugar := l.Sugar()

	// Initialize Redis with config
	redisClient := cache.NewRedisClient(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)

	// Initialize router
	router := internalhttp.NewRouter(sugar, redisClient)

	// Start server with config
	addr := ":" + cfg.Server.Port
	sugar.Infof("server started on port %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		sugar.Fatalw("Error starting server", "addr", addr, "error", err)
	}
}
