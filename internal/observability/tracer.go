package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.8.0"
)

// InitTracer initializes the OpenTelemetry tracer provider
// It returns a shutdown function that should be called before the application exits
func InitTracer(env string) (func(context.Context) error, error) {
	exporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go-boilerplate"),
			semconv.DeploymentEnvironmentKey.String(env),
		)),
		trace.WithSampler(sampler(env)),
	)

	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

// sampler returns a trace sampler based on the environment
func sampler(env string) trace.Sampler {
	switch env {
	case "production":
		return trace.TraceIDRatioBased(0.1)
	default:
		return trace.AlwaysSample()
	}
}
