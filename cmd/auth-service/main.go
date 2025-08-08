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

package main

import (
	"log"
	"net/http"

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

	router := internalhttp.NewRouter(sugar)

	sugar.Info("server started in port :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		sugar.Fatalw("Error starting server on port :8080", "error", err)
	}
}
