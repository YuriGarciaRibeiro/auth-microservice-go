package main

import (
	"log"
	"net/http"

	internalhttp "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/transport/http"

	"go.uber.org/zap"
)

func main() {

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
