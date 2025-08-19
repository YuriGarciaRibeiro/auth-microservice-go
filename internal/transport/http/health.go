package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/version"
	"github.com/prometheus/client_golang/prometheus"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
	// timeouts
	dbTimeout    time.Duration
	redisTimeout time.Duration
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client, dbTimeout, redisTimeout time.Duration) *HealthHandler {
	// registra métricas só uma vez
	onceRegisterReadinessMetrics()
	return &HealthHandler{
		db:           db,
		redis:        rdb,
		dbTimeout:    dbTimeout,
		redisTimeout: redisTimeout,
	}
}

// HealthzResponse represents the health check response
type HealthzResponse struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"auth-microservice"`
}

// @Summary Health Check
// @Description Basic health check endpoint that returns service status
// @Tags health
// @Produce json
// @Success 200 {object} HealthzResponse
// @Router /healthz [get]
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": version.Service,
	})
}

// ReadyzResponse represents the readiness check response
type ReadyzResponse struct {
	Status string `json:"status" example:"ok"`
	Deps   struct {
		Postgres struct {
			Ready bool `json:"ready" example:"true"`
		} `json:"postgres"`
		Redis struct {
			Ready bool `json:"ready" example:"true"`
		} `json:"redis"`
	} `json:"deps"`
}

// @Summary Readiness Check
// @Description Readiness check that verifies database and Redis connectivity
// @Tags health
// @Produce json
// @Success 200 {object} ReadyzResponse "All dependencies are ready"
// @Failure 503 {object} ReadyzResponse "One or more dependencies are not ready"
// @Router /readyz [get]
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	dbOK := h.pingDB(r.Context())
	redisOK := h.pingRedis(r.Context())

	status := http.StatusOK
	if !dbOK || !redisOK {
		status = http.StatusServiceUnavailable
	}

	// atualiza métricas
	setReadyGauge("postgres", boolToGauge(dbOK))
	setReadyGauge("redis", boolToGauge(redisOK))

	writeJSON(w, status, map[string]any{
		"status": "ok",
		"deps": map[string]any{
			"postgres": map[string]any{"ready": dbOK},
			"redis":    map[string]any{"ready": redisOK},
		},
	})
}

// BuildzResponse represents the build information response
type BuildzResponse struct {
	Service   string `json:"service" example:"auth-microservice"`
	Version   string `json:"version" example:"v1.0.0"`
	Commit    string `json:"commit" example:"abc123def456"`
	BuildTime string `json:"build_time" example:"2025-08-19T12:34:56Z"`
	Env       string `json:"env" example:"dev"`
}

// @Summary Build Information
// @Description Returns build and version information about the service
// @Tags health
// @Produce json
// @Success 200 {object} BuildzResponse
// @Router /buildz [get]
func (h *HealthHandler) Buildz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service":    version.Service,
		"version":    version.Version,
		"commit":     version.Commit,
		"build_time": version.BuildTime,
		"env":        r.Context().Value("app_env"),
	})
}

// ---- internals ----

func (h *HealthHandler) pingDB(parent context.Context) bool {
	if h.db == nil {
		return false
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(parent, orDefault(h.dbTimeout, 2*time.Second))
	defer cancel()
	return sqlDB.PingContext(ctx) == nil
}

func (h *HealthHandler) pingRedis(parent context.Context) bool {
	if h.redis == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(parent, orDefault(h.redisTimeout, 1*time.Second))
	defer cancel()
	return h.redis.Ping(ctx).Err() == nil
}

func orDefault[T comparable](v T, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}

func boolToGauge(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// ------------- Prometheus -------------

var (
	readyGauge     *prometheus.GaugeVec
	readyGaugeOnce = make(chan struct{}, 1)
)

func onceRegisterReadinessMetrics() {
	select {
	case readyGaugeOnce <- struct{}{}:
		readyGauge = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "service_ready",
				Help: "Readiness of dependencies; 1=ready, 0=not ready",
			},
			[]string{"component"},
		)
		prometheus.MustRegister(readyGauge)
	default:
		// já registrado
	}
}

func setReadyGauge(component string, value float64) {
	if readyGauge != nil {
		readyGauge.WithLabelValues(component).Set(value)
	}
}
