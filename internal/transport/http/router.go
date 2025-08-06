package http

import (
	"net/http"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db"
	handler "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/transport/http/Handler"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.SugaredLogger) http.Handler{

	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Infow("New request",
				"Method", r.Method,
				"Path", r.URL.Path,
			)
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/Healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	validate := validator.New()

	// Repository (in-memory)
	userRepo := db.NewInMemoryUserRepository()

	// Use Case
	signUpUseCase := usecase.NewSignupUseCase(userRepo)

	// Handler
	authHandler := &handler.AuthHandler{
		Signgup: signUpUseCase,
		Validate : validate,
	}

	r.Post("/signup", authHandler.SignUpHandler)

	return r
}