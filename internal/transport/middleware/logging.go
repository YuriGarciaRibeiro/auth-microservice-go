package middleware

import (
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// statusWriter wraps http.ResponseWriter to capture HTTP status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Logging middleware emits a structured log per request with duration and status.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}

		next.ServeHTTP(sw, r)

		duration := time.Since(start)
		reqID, _ := GetRequestID(r.Context())

		zap.L().Info("http_request",
			zap.String("request_id", reqID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", sw.status),
			zap.Duration("duration", duration),
			zap.String("remote_ip", clientIP(r)),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}

func clientIP(r *http.Request) string {
	// honor X-Forwarded-For if present (behind proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
