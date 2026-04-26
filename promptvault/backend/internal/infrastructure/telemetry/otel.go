// Package telemetry инициализирует OpenTelemetry SDK для distributed
// tracing. Phase 16 Этап 3 — auto-instrumentation Chi+GORM, экспорт
// spans в Tempo через OTLP gRPC.
//
// Setup() возвращает shutdown функцию для graceful drain буфера spans
// перед exit процесса (defer'ится в main.go).
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc/credentials/insecure"
)

const serviceName = "promptvault-api"

// Setup инициализирует TracerProvider, OTLP exporter и global propagator.
// Возвращает shutdown функцию для defer'а в main.
//
// Если cfg.Telemetry.Enabled=false — возвращает no-op shutdown без
// инициализации SDK (нулевой overhead).
//
// Sampler: ParentBased(TraceIDRatioBased(rate)) — uniform sampling
// с уважением к родительскому решению (для distributed tracing
// корректно прокидывает sample/no-sample вниз по chain).
func Setup(ctx context.Context, cfg config.TelemetryConfig, environment, release string) (shutdown func(context.Context) error, err error) {
	if !cfg.Enabled {
		slog.Info("telemetry.disabled", "reason", "TELEMETRY_ENABLED=false")
		return func(context.Context) error { return nil }, nil
	}

	// Resource — identity сервиса в traces (service.name, version, environment).
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(release),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: resource: %w", err)
	}

	// OTLP gRPC exporter → Tempo (или OTel Collector в будущем).
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
		otlptracegrpc.WithTimeout(30*time.Second),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  2 * time.Minute,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithMaxQueueSize(2048),
		),
		sdktrace.WithResource(res),
		// ParentBased respects propagated decision; root spans sampled by ratio.
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(cfg.TracesSampleRate),
		)),
	)
	otel.SetTracerProvider(tp)

	// W3C TraceContext + Baggage — стандарт для inter-service propagation.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("telemetry.initialized",
		"endpoint", cfg.OTLPEndpoint,
		"sample_rate", cfg.TracesSampleRate,
		"environment", environment,
		"release", release,
	)

	return func(shutdownCtx context.Context) error {
		// Combine: TracerProvider.Shutdown drains batches.
		var combined error
		if err := tp.Shutdown(shutdownCtx); err != nil {
			combined = errors.Join(combined, fmt.Errorf("tracer provider: %w", err))
		}
		return combined
	}, nil
}
