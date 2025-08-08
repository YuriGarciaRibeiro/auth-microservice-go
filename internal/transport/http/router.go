package http

import (
	"net/http"
	"os"
	"time"

	"strconv"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/service/token"
	handler "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/transport/http/handler"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/YuriGarciaRibeiro/auth-microservice-go/docs"
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

	gormDb := db.ConnectPostgres()
	// Token Service
	jwtExpirationHoursStr := os.Getenv("JWT_EXPIRATION_HOURS")
	jwtExpirationHours, err := strconv.Atoi(jwtExpirationHoursStr)
	if err != nil {
		logger.Fatalf("Invalid JWT_EXPIRATION_HOURS: %v", err)
	}

	jwtRefreshExpirationHoursStr := os.Getenv("JWT_REFRESH_EXPIRATION_HOURS")
	jwtRefreshExpirationHours, err := strconv.Atoi(jwtRefreshExpirationHoursStr)
	if err != nil {
		logger.Fatalf("Invalid JWT_REFRESH_EXPIRATION_HOURS: %v", err)
	}

	tokenService := token.NewTokenService(
		os.Getenv("JWT_ISSUER"),
		os.Getenv("JWT_AUDIENCE"),
		time.Duration(jwtExpirationHours)*time.Hour,
		time.Duration(jwtRefreshExpirationHours)*time.Hour,
	)

	// Repository
	userRepo := db.NewGormUserRepository(gormDb)

	// Use Case
	signUpUseCase := usecase.NewSignupUseCase(userRepo)
	loginUseCase := usecase.NewLoginUseCase(userRepo)

	// Handler
	authHandler := &handler.AuthHandler{
		Signup:      signUpUseCase,
		Login:       loginUseCase,
		Validate:    validate,
		TokenService: tokenService,
	}

	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.SignUpHandler)
		r.Post("/login", authHandler.LoginHandler)
		r.Get("/me", authHandler.MeHandler)
	})

	r.Get("/docs/*", httpSwagger.WrapHandler)

	return r
}