package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once sync.Once

	httpRequestTotal *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec

	authTokensIssuedTotal *prometheus.CounterVec
	authTokensRevokedTotal *prometheus.CounterVec
)

func MustRegister(){
	once.Do(func() {
		httpRequestTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		)

		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Histogram of HTTP request durations in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status"},
		)

		authTokensIssuedTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_tokens_issued_total",
				Help: "Total number of issued authentication tokens",
			},
			[]string{"user_id"},
		)

		authTokensRevokedTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_tokens_revoked_total",
				Help: "Total number of revoked authentication tokens",
			},
			[]string{"user_id"},
		)

		prometheus.MustRegister(
			httpRequestTotal,
			httpRequestDuration,
			authTokensIssuedTotal,
			authTokensRevokedTotal,
		)
	})
}

func ObserveHTTPRequests(method, route, status string, seconds float64) {
	httpRequestTotal.WithLabelValues(method, route, status).Inc()
	httpRequestDuration.WithLabelValues(method, route, status).Observe(seconds)
}

func IncAuthTokensIssued(userID string) {
	authTokensIssuedTotal.WithLabelValues(userID).Inc()
}

func IncAuthTokensRevoked(userID string) {
	authTokensRevokedTotal.WithLabelValues(userID).Inc()
}