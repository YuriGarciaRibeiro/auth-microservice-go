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

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": version.Service,
	})
}

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
