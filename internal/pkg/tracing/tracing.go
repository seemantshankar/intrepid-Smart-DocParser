package tracing

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

// Config holds the configuration for tracing
type Config struct {
	ServiceName  string
	Environment  string
	JaegerURL    string
	SamplingRate float64
}

// TracerProvider wraps the OpenTelemetry TracerProvider
type TracerProvider struct {
	*sdktrace.TracerProvider
}

// InitTracer initializes and returns a new TracerProvider
func InitTracer(cfg Config, logger *zap.Logger) (*TracerProvider, error) {
	var tp *sdktrace.TracerProvider
	// If Jaeger URL is provided, configure exporter; otherwise initialize without exporter
	if cfg.JaegerURL != "" {
		exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.JaegerURL)))
		if err != nil {
			return nil, err
		}
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRate)),
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(cfg.ServiceName),
				semconv.DeploymentEnvironmentKey.String(cfg.Environment),
			)),
		)
	} else {
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRate)),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(cfg.ServiceName),
				semconv.DeploymentEnvironmentKey.String(cfg.Environment),
			)),
		)
	}

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracerProvider{TracerProvider: tp}, nil
}

// Middleware adds tracing to the request context
func Middleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := otel.GetTextMapPropagator().Extract(
			c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header),
		)

		tr := otel.Tracer("http")
		ctx, span := tr.Start(ctx, c.FullPath())
		defer span.End()

		// Add trace ID to response headers
		sc := span.SpanContext()
		if sc.HasTraceID() {
			c.Header("X-Trace-ID", sc.TraceID().String())
		}

		// Update request context
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
