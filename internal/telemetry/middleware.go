package telemetry

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// GinMiddleware returns a Gin middleware that traces HTTP requests
func GinMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// Extract context from incoming headers
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start span
		spanName := c.Request.Method + " " + c.FullPath()
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPTarget(c.Request.URL.Path),
				semconv.HTTPScheme(c.Request.URL.Scheme),
				semconv.NetHostName(c.Request.Host),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Inject context into response headers for downstream tracing
		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))

		// Wrap response writer to capture status code
		wrapped := &ginResponseWriter{ResponseWriter: c.Writer, statusCode: http.StatusOK}
		c.Writer = wrapped

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// Record response attributes
		span.SetAttributes(
			semconv.HTTPStatusCode(wrapped.statusCode),
			attribute.Int("http.response.size", wrapped.size),
		)

		// Record error if status >= 400
		if wrapped.statusCode >= 400 {
			span.SetAttributes(attribute.String("error", "http_error"))
		}
	}
}

// HTTPMiddleware returns a standard HTTP middleware that traces HTTP requests
func HTTPMiddleware(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract context from incoming headers
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start span
			spanName := r.Method + " " + r.URL.Path
			ctx, span := tracer.Start(ctx, spanName,
				trace.WithAttributes(
					semconv.HTTPMethod(r.Method),
					semconv.HTTPTarget(r.URL.Path),
					semconv.HTTPScheme(r.URL.Scheme),
					semconv.NetHostName(r.Host),
				),
				trace.WithSpanKind(trace.SpanKindServer),
			)
			defer span.End()

			// Inject context into response headers for downstream tracing
			propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Record response attributes
			span.SetAttributes(
				semconv.HTTPStatusCode(wrapped.statusCode),
				attribute.Int("http.response.size", wrapped.size),
			)

			// Record error if status >= 400
			if wrapped.statusCode >= 400 {
				span.SetAttributes(attribute.String("error", "http_error"))
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// ginResponseWriter wraps gin.ResponseWriter to capture status code and size
type ginResponseWriter struct {
	gin.ResponseWriter
	statusCode int
	size       int
}

func (rw *ginResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ginResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// StartClientSpan starts a span for outgoing HTTP client requests
func StartClientSpan(ctx context.Context, method, url string) (context.Context, trace.Span) {
	tracer := otel.Tracer("goaggregator.client")
	return tracer.Start(ctx, method+" "+url,
		trace.WithAttributes(
			semconv.HTTPMethod(method),
			semconv.HTTPTarget(url),
			semconv.HTTPScheme("http"),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
}

// InjectHeaders injects trace context into HTTP headers for outgoing requests
func InjectHeaders(ctx context.Context, headers http.Header) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(headers))
}
