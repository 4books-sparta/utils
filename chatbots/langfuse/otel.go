package langfuse

import (
	"context"
	"encoding/binary"
	"log"
	"math/rand/v2"
	"os"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/4books-sparta/utils/chatbots/langfuse/internal/pkg/api"
)

const (
	defaultHost           = "https://cloud.langfuse.com"
	otlpTracesPath        = "/api/public/otel/v1/traces"
	defaultServiceName    = "chatbot-ms"
	ingestionVersionKey   = "x-langfuse-ingestion-version"
	ingestionVersionValue = "4"
	tracerName            = "github.com/4books-sparta/utils/chatbots/langfuse"
)

var (
	errorHandlerOnce sync.Once
	noKeyWarningOnce sync.Once
)

// providerConfig holds what is needed to build the tracer provider.
type providerConfig struct {
	serviceName   string
	environment   string
	flushInterval time.Duration
	// spanProcessor, when set, replaces the OTLP exporter (used by tests).
	spanProcessor sdktrace.SpanProcessor
}

// newTracerProvider builds the OTel tracer provider that exports to Langfuse.
// Without LANGFUSE_PUBLIC_KEY (and no custom span processor) the provider has
// no processor at all: spans are created but never exported.
func newTracerProvider(ctx context.Context, cfg providerConfig) (*sdktrace.TracerProvider, error) {
	registerErrorHandler()

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithIDGenerator(idGenerator{}),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(buildResource(cfg)),
	}

	switch {
	case cfg.spanProcessor != nil:
		opts = append(opts, sdktrace.WithSpanProcessor(cfg.spanProcessor))
	case os.Getenv("LANGFUSE_PUBLIC_KEY") == "":
		noKeyWarningOnce.Do(func() {
			log.Printf("langfuse otel: LANGFUSE_PUBLIC_KEY is empty, tracing is disabled")
		})
	default:
		exporter, err := newExporter(ctx)
		if err != nil {
			return nil, err
		}
		var bspOpts []sdktrace.BatchSpanProcessorOption
		if cfg.flushInterval > 0 {
			bspOpts = append(bspOpts, sdktrace.WithBatchTimeout(cfg.flushInterval))
		}
		opts = append(opts, sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter, bspOpts...)))
	}

	return sdktrace.NewTracerProvider(opts...), nil
}

func newExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	host := os.Getenv("LANGFUSE_HOST")
	if host == "" {
		host = defaultHost
	}
	endpoint := strings.TrimRight(host, "/") + otlpTracesPath

	return otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization":     api.BasicAuth(os.Getenv("LANGFUSE_PUBLIC_KEY"), os.Getenv("LANGFUSE_SECRET_KEY")),
			ingestionVersionKey: ingestionVersionValue,
		}),
	)
}

func buildResource(cfg providerConfig) *resource.Resource {
	name := cfg.serviceName
	if name == "" {
		name = defaultServiceName
	}
	attrs := []attribute.KeyValue{attribute.String("service.name", name)}
	if cfg.environment != "" {
		attrs = append(attrs, attribute.String("deployment.environment", cfg.environment))
	}

	own := resource.NewSchemaless(attrs...)
	merged, err := resource.Merge(resource.Default(), own)
	if err != nil {
		return own
	}
	return merged
}

// registerErrorHandler installs a process-wide OTel error handler that logs
// exporter failures.
func registerErrorHandler() {
	errorHandlerOnce.Do(func() {
		otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
			log.Printf("langfuse otel: %v", err)
		}))
	})
}

type traceIDKey struct{}

// contextWithTraceID returns a context that makes idGenerator use id as the
// trace ID of the next root span started from it.
func contextWithTraceID(ctx context.Context, id trace.TraceID) context.Context {
	return context.WithValue(ctx, traceIDKey{}, id)
}

// idGenerator uses the trace ID carried by the context, when valid, and
// random IDs otherwise. Span IDs are always random.
type idGenerator struct{}

var _ sdktrace.IDGenerator = idGenerator{}

func (g idGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	if id, ok := ctx.Value(traceIDKey{}).(trace.TraceID); ok && id.IsValid() {
		return id, randomSpanID()
	}
	return randomTraceID(), randomSpanID()
}

func (g idGenerator) NewSpanID(context.Context, trace.TraceID) trace.SpanID {
	return randomSpanID()
}

func randomTraceID() trace.TraceID {
	var id trace.TraceID
	for !id.IsValid() {
		binary.BigEndian.PutUint64(id[:8], rand.Uint64())
		binary.BigEndian.PutUint64(id[8:], rand.Uint64())
	}
	return id
}

func randomSpanID() trace.SpanID {
	var id trace.SpanID
	for !id.IsValid() {
		binary.BigEndian.PutUint64(id[:], rand.Uint64())
	}
	return id
}
