package middleware

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// Recover middleware captures panics, logs the stack trace, and returns 500.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				reqID, _ := GetRequestID(r.Context())
				zap.L().Error("panic recovered",
					zap.Any("error", rec),
					zap.ByteString("stack", debug.Stack()),
					zap.String("request_id", reqID),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
