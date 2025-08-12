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
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/cache"
	internalhttp "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/transport/http"
	"github.com/joho/godotenv"

	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Erro ao iniciar serviço de log: ", err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPass := os.Getenv("REDIS_PASS")
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	redisClient := cache.NewRedisClient(redisAddr, redisPass, redisDB)

	router := internalhttp.NewRouter(sugar, redisClient)

	sugar.Info("server started in port :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		sugar.Fatalw("Error starting server on port :8080", "error", err)
	}
}
