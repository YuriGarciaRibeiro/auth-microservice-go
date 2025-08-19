package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(code int) {
	if w.wroteHeader {
		// não chama de novo; só atualiza o campo para log
		w.status = code
		return
	}
	w.status = code
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	// se ninguém chamou WriteHeader ainda, force 200 uma única vez
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(sw, r)

		requestID, _ := GetRequestID(r.Context())
		zap.L().Info("http_request",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", sw.status),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote_ip", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}
