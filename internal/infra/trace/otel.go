package trace

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Shutdown func(ctx context.Context) error

func Init() (Shutdown, error) {
	svc := getenv("OTEL_SERVICE_NAME", "auth-service")
	jaegerEndpoint := getenv("OTEL_EXPORTER_JAEGER_ENDPOINT", "http://localhost:14268/api/traces")

	// Exportador Jaeger (Thrift over HTTP)
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return nil, err
	}

	// Recursos (identidade do serviço)
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(svc),
		),
	)
	if err != nil {
		return nil, err
	}

	// Tracer Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
