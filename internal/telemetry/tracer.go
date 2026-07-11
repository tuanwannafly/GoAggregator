package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// InitTracer initializes OpenTelemetry tracer provider with stdout exporter
func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("create stdout exporter: %w", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("0.1.0"),
			attribute.String("environment", getEnv("ENV", "development")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("OpenTelemetry tracer initialized", "service", serviceName)
	return tp, nil
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}
	slog.Info("Shutting down OpenTelemetry tracer provider")
	return tp.Shutdown(ctx)
}

// Tracer returns a tracer for the given name
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan starts a new span with the given name and options
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer("goaggregator")
	return tracer.Start(ctx, name, opts...)
}

// AddSpanAttributes adds attributes to the current span
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.SetAttributes(attrs...)
	}
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span != nil && err != nil {
		span.RecordError(err, trace.WithAttributes(attrs...))
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// AttributeString creates a string attribute
func AttributeString(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// AttributeInt creates an int attribute
func AttributeInt(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// AttributeInt64 creates an int64 attribute
func AttributeInt64(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// AttributeBool creates a bool attribute
func AttributeBool(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}

// AttributeFloat64 creates a float64 attribute
func AttributeFloat64(key string, value float64) attribute.KeyValue {
	return attribute.Float64(key, value)
}

// WithAttributes creates a SpanStartOption with attributes
func WithAttributes(attrs ...attribute.KeyValue) trace.SpanStartOption {
	return trace.WithAttributes(attrs...)
}