package http

import (
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/YuriGarciaRibeiro/auth-microservice-go/docs"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/cache"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db"
	tokenSvc "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/service/token"
	handler "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/transport/http/handler"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/transport/middleware"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

func mustDuration(envKey, def string) time.Duration {
	v := os.Getenv(envKey)
	if v == "" {
		v = def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		panic("invalid duration for " + envKey + ": " + err.Error())
	}
	return d
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func newGoRedisFromEnv() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}

func NewRouter(logger *zap.SugaredLogger, appCache *cache.RedisClient) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recover)
	r.Use(middleware.Logging)
	r.Use(middleware.PrometheusHTTP) 

	r.Use(func(h http.Handler) http.Handler {
		return otelhttp.NewHandler(h, "http.server").(http.HandlerFunc)
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	validate := validator.New()

	gormDb := db.ConnectPostgres()

	// TokenService config (Plan B: ACCESS_/REFRESH_ envs).
	accessSecret := []byte(os.Getenv("ACCESS_SECRET"))
	refreshSecret := []byte(os.Getenv("REFRESH_SECRET"))
	if len(accessSecret) == 0 || len(refreshSecret) == 0 {
		logger.Fatal("ACCESS_SECRET and REFRESH_SECRET must be set")
	}

	cfg := tokenSvc.Config{
		AccessSecret:    accessSecret,
		RefreshSecret:   refreshSecret,
		AccessTTL:       mustDuration("ACCESS_TOKEN_TTL", "15m"),
		RefreshTTL:      mustDuration("REFRESH_TOKEN_TTL", "168h"),
		Issuer:          os.Getenv("JWT_ISSUER"),
		DefaultAudience: splitCSV(os.Getenv("JWT_AUDIENCE")),
	}

	// Raw go-redis client for TokenService.
	rawRedis := newGoRedisFromEnv()
	tokenService := tokenSvc.NewService(cfg, rawRedis)

	// Repositories.
	userRepo := db.NewGormUserRepository(gormDb)
	clientRepo := db.NewGormClientRepository(gormDb)
	permRepo := db.NewGormPermissionRepositoryWithCache(
		gormDb,
		rawRedis,
		mustDuration("PERM_CACHE_TTL", "15m"),
	)

	

	// Use cases.
	signUpUC := usecase.NewSignupUseCase(userRepo)
	loginUC := usecase.NewLoginUseCase(userRepo)
	clientUC := usecase.NewClientCredentialsUseCase(clientRepo)
	permUC := usecase.NewPermAdminUseCase(permRepo)

	// Handlers.
	authHandler := &handler.AuthHandler{
		Signup:       signUpUC,
		Login:        loginUC,
		Validate:     validate,
		TokenService: tokenService,
		Cache:        appCache,
		PermissionRepository: permRepo,
	}

	clientTokenHandler := &handler.ClientTokenHandler{
		Validate:     validate,
		UC:           clientUC,
		TokenService: tokenService,
		PermissionRepository: permRepo,
	}

	adminHandler := &handler.AdminPermHandler{UC: permUC, Validate: validate}

	// Routes.
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.SignUpHandler)
		r.Post("/login", authHandler.LoginHandler)
		r.Post("/logout", authHandler.LogoutHandler)
		r.Post("/refresh", authHandler.RefreshHandler)
		r.Post("/introspect", authHandler.IntrospectHandler)
		r.Post("/token", clientTokenHandler.ServeHTTP)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.Authn(tokenService))
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/scopes", adminHandler.CreateScope)
		r.Get("/scopes", adminHandler.ListScopes)

		r.Post("/roles", adminHandler.CreateRole)
		r.Get("/roles", adminHandler.ListRoles)
		r.Post("/roles/{roleId}/scopes", adminHandler.AddScopesToRole)

		r.Post("/users/{userId}/roles", adminHandler.AddRolesToUser)
		r.Get("/users/{userId}/roles", adminHandler.ListUserRoles)
		r.Get("/users/{userId}/scopes", adminHandler.ListUserEffective)
		r.Post("/users/{userId}/scopes/grant", adminHandler.GrantUserScope)
		r.Post("/users/{userId}/scopes/revoke", adminHandler.RevokeUserScope)

		r.Post("/clients/{clientId}/scopes", adminHandler.AddScopesToClient)
		r.Get("/clients/{clientId}/scopes", adminHandler.ListClientScopes)
	})

	r.Handle("/metrics", promhttp.Handler())

	// Swagger UI.
	r.Get("/docs/*", httpSwagger.WrapHandler)

	return r
}
