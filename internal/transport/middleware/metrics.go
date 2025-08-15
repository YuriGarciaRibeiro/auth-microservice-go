package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/metrics"
	"github.com/go-chi/chi"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func PrometheusHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		duration := time.Since(start).Seconds()
		route := routePattern(r)
		status := strconv.Itoa(sw.status)
		metrics.ObserveHTTPRequests(r.Method, route, status, duration)
	})
}

func routePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if pat := rctx.RoutePattern(); pat != "" {
			return pat
		}
	}
	return r.URL.Path
}