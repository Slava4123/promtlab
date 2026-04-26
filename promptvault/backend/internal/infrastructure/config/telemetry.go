package config

// TelemetryConfig — настройки OpenTelemetry SDK для distributed tracing
// (Phase 16 Этап 3). Spans экспортируются в Tempo через OTLP gRPC.
type TelemetryConfig struct {
	// Enabled — feature flag. По default false: SDK не инициализируется,
	// нулевой overhead. В prod выставляется через env TELEMETRY_ENABLED=true.
	Enabled bool `koanf:"enabled"`

	// OTLPEndpoint — host:port gRPC receiver Tempo. Default "tempo:4317"
	// для Docker network production setup.
	OTLPEndpoint string `koanf:"otlp_endpoint"`

	// TracesSampleRate — доля запросов трассируемых (0.0..1.0).
	// Default 0.1 в prod (10%) — балансирует storage cost и debugging coverage.
	// В dev обычно 1.0 (100%) для полной видимости при отладке.
	TracesSampleRate float64 `koanf:"traces_sample_rate"`
}
