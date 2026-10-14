// Package langfuse sends chatbot traces to Langfuse through its OpenTelemetry
// (OTLP/HTTP) endpoint and scores through the public ingestion API.
package langfuse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/4books-sparta/utils/chatbots/langfuse/internal/pkg/api"
	"github.com/4books-sparta/utils/chatbots/langfuse/model"
)

const (
	defaultFlushInterval = 500 * time.Millisecond
	flushTimeout         = 15 * time.Second
	scoreTimeout         = 5 * time.Second
)

// Langfuse is a tracing client for Langfuse. The OTel provider is built
// lazily on first use, so the With* options must be applied before tracing.
type Langfuse struct {
	ctx           context.Context
	flushInterval time.Duration
	environment   string
	serviceName   string
	spanProcessor sdktrace.SpanProcessor

	client *api.Client // nil when LANGFUSE_PUBLIC_KEY is not set
	scores sync.WaitGroup

	initOnce     sync.Once
	provider     *sdktrace.TracerProvider
	reg          *registry
	shutdownOnce sync.Once
	shutdownErr  error
}

// New returns a client configured from LANGFUSE_HOST, LANGFUSE_PUBLIC_KEY and
// LANGFUSE_SECRET_KEY. ctx is only used to build the OTLP exporter.
func New(ctx context.Context) *Langfuse {
	l := &Langfuse{
		ctx:           ctx,
		flushInterval: defaultFlushInterval,
		serviceName:   defaultServiceName,
	}
	if os.Getenv("LANGFUSE_PUBLIC_KEY") != "" {
		l.client = api.New()
	}
	return l
}

// WithEnvironment sets the Langfuse environment of traces and scores.
func (l *Langfuse) WithEnvironment(env string) *Langfuse {
	l.environment = env
	return l
}

// WithFlushInterval sets the batch timeout of the span exporter.
func (l *Langfuse) WithFlushInterval(d time.Duration) *Langfuse {
	l.flushInterval = d
	return l
}

// WithServiceName sets the OTel service.name resource attribute.
func (l *Langfuse) WithServiceName(name string) *Langfuse {
	l.serviceName = name
	return l
}

// WithSpanProcessor replaces the OTLP exporter with p (intended for tests).
func (l *Langfuse) WithSpanProcessor(p sdktrace.SpanProcessor) *Langfuse {
	l.spanProcessor = p
	return l
}

func (l *Langfuse) init() {
	l.initOnce.Do(func() {
		ctx := l.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		tp, err := newTracerProvider(ctx, providerConfig{
			serviceName:   l.serviceName,
			environment:   l.environment,
			flushInterval: l.flushInterval,
			spanProcessor: l.spanProcessor,
		})
		if err != nil {
			log.Printf("langfuse otel: cannot build exporter, tracing is disabled: %v", err)
			tp = sdktrace.NewTracerProvider(sdktrace.WithIDGenerator(idGenerator{}))
		}
		l.provider = tp
		l.reg = newRegistry(tp.Tracer(tracerName))
	})
}

// Trace registers a trace. Nothing is exported until its first observation
// starts; trace attributes are read from t when each observation ends, so
// later changes to t (tags, metadata) are picked up.
func (l *Langfuse) Trace(t *model.Trace) (*model.Trace, error) {
	l.init()
	if t.Environment == "" {
		t.Environment = l.environment
	}
	t.ID = buildID(&t.ID)
	l.reg.putTrace(t)
	return t, nil
}

// SetTraceOutput sets the output of a trace, exported on its root span.
func (l *Langfuse) SetTraceOutput(traceID string, out any) {
	l.init()
	l.reg.setTraceOutput(traceID, out)
}

// Span starts a span. parentID, when not nil, overrides s.ParentObservationID.
func (l *Langfuse) Span(s *model.Span, parentID *string) (*model.Span, error) {
	l.init()
	s.TraceID = l.ensureTrace(s.TraceID, s.Name)
	s.ID = buildID(&s.ID)
	if parentID != nil {
		s.ParentObservationID = *parentID
	}
	s.StartTime = nowIfNil(s.StartTime)
	l.reg.startSpan(s.TraceID, s.ID, s.ParentObservationID, s.Name, *s.StartTime)
	return s, nil
}

// SpanEnd ends the span with ID s.ID, recording the final state of s.
func (l *Langfuse) SpanEnd(s *model.Span) (*model.Span, error) {
	if s.ID == "" {
		return nil, errors.New("span ID is required")
	}
	l.init()
	ls, ts, ok := l.reg.endSpan(s.ID)
	if !ok {
		return nil, fmt.Errorf("span %s not found", s.ID)
	}
	s.EndTime = nowIfNil(s.EndTime)

	attrs := observationAttrs(observationTypeSpan, s.Input, s.Output, s.Metadata, s.Level, s.StatusMessage)
	l.finish(ls, ts, s.Name, attrs, s.Level, s.StatusMessage, *s.EndTime)
	return s, nil
}

// Generation starts a generation. parentID, when not nil, overrides
// g.ParentObservationID.
func (l *Langfuse) Generation(g *model.Generation, parentID *string) (*model.Generation, error) {
	l.init()
	g.TraceID = l.ensureTrace(g.TraceID, g.Name)
	g.ID = buildID(&g.ID)
	if parentID != nil {
		g.ParentObservationID = *parentID
	}
	g.StartTime = nowIfNil(g.StartTime)
	l.reg.startSpan(g.TraceID, g.ID, g.ParentObservationID, g.Name, *g.StartTime)
	return g, nil
}

// GenerationEnd ends the generation with ID g.ID, recording the final state of g.
func (l *Langfuse) GenerationEnd(g *model.Generation) (*model.Generation, error) {
	if g.ID == "" {
		return nil, errors.New("generation ID is required")
	}
	l.init()
	ls, ts, ok := l.reg.endSpan(g.ID)
	if !ok {
		return nil, fmt.Errorf("generation %s not found", g.ID)
	}
	g.EndTime = nowIfNil(g.EndTime)

	attrs := observationAttrs(observationTypeGeneration, g.Input, g.Output, g.Metadata, g.Level, g.StatusMessage)
	attrs = append(attrs, generationAttrs(g)...)
	l.finish(ls, ts, g.Name, attrs, g.Level, g.StatusMessage, *g.EndTime)
	return g, nil
}

// Event records a point-in-time observation: its span starts and ends at
// e.StartTime (or now). parentID, when not nil, overrides e.ParentObservationID.
func (l *Langfuse) Event(e *model.Event, parentID *string) (*model.Event, error) {
	l.init()
	e.TraceID = l.ensureTrace(e.TraceID, e.Name)
	e.ID = buildID(&e.ID)
	if parentID != nil {
		e.ParentObservationID = *parentID
	}
	e.StartTime = nowIfNil(e.StartTime)

	l.reg.startSpan(e.TraceID, e.ID, e.ParentObservationID, e.Name, *e.StartTime)
	ls, ts, ok := l.reg.endSpan(e.ID)
	if !ok {
		return nil, fmt.Errorf("event %s not found", e.ID)
	}
	attrs := observationAttrs(observationTypeEvent, e.Input, e.Output, e.Metadata, e.Level, e.StatusMessage)
	l.finish(ls, ts, e.Name, attrs, e.Level, e.StatusMessage, *e.StartTime)
	return e, nil
}

// Score sends a score through the ingestion API, asynchronously. A UUID trace
// ID is converted to the OTel trace ID used for the trace. Without
// LANGFUSE_PUBLIC_KEY it does nothing.
func (l *Langfuse) Score(s *model.Score) (*model.Score, error) {
	if s.TraceID == "" {
		return nil, errors.New("trace ID is required")
	}
	if tid, ok := model.TraceIDFromUUID(s.TraceID); ok {
		s.TraceID = tid.String()
	}
	s.ID = buildID(&s.ID)
	if l.environment != "" {
		s.Environment = l.environment
	}
	if l.client == nil {
		return s, nil
	}

	body := *s
	l.scores.Add(1)
	go func() {
		defer l.scores.Done()
		ctx, cancel := context.WithTimeout(context.Background(), scoreTimeout)
		defer cancel()
		req := api.Ingestion{Batch: []model.IngestionEvent{{
			ID:        buildID(nil),
			Type:      model.IngestionEventTypeScoreCreate,
			Timestamp: time.Now().UTC(),
			Body:      &body,
		}}}
		res := api.IngestionResponse{}
		if err := l.client.Ingestion(ctx, &req, &res); err != nil && os.Getenv("DEBUG") != "" {
			log.Printf("langfuse score ingestion error: %v", err)
		}
	}()
	return s, nil
}

// Flush exports the spans that have ended. Spans still open are not touched.
func (l *Langfuse) Flush(ctx context.Context) {
	l.init()
	ctx, cancel := context.WithTimeout(ctx, flushTimeout)
	defer cancel()
	_ = l.provider.ForceFlush(ctx)
}

// Shutdown ends the spans still open (marked as unterminated errors), exports
// everything, waits for pending scores and stops the provider. It is
// idempotent; tracing after Shutdown is a no-op.
func (l *Langfuse) Shutdown(ctx context.Context) error {
	l.init()
	l.shutdownOnce.Do(func() {
		l.reg.endAll()
		flushErr := l.provider.ForceFlush(ctx)
		shutdownErr := l.provider.Shutdown(ctx)
		scoresErr := waitGroup(ctx, &l.scores)
		l.shutdownErr = errors.Join(flushErr, shutdownErr, scoresErr)
	})
	return l.shutdownErr
}

// ensureTrace returns traceID, creating a trace named name when it is empty.
func (l *Langfuse) ensureTrace(traceID, name string) string {
	if traceID != "" {
		return traceID
	}
	t, _ := l.Trace(&model.Trace{Name: name})
	return t.ID
}

// finish sets the observation and trace attributes on a span removed from the
// registry and ends it at end.
func (l *Langfuse) finish(ls liveSpan, ts *traceState, name string, attrs []attribute.KeyValue,
	level model.ObservationLevel, statusMessage string, end time.Time) {
	if name != "" {
		ls.span.SetName(name)
	}
	t, out := l.reg.traceView(ts)
	attrs = append(attrs, l.traceAttrs(t, out, ls.isRoot)...)
	ls.span.SetAttributes(attrs...)
	if level == model.ObservationLevelError {
		ls.span.SetStatus(codes.Error, statusMessage)
	}
	ls.span.End(trace.WithTimestamp(end))
}

// traceAttrs returns the trace-level attributes repeated on every span; input
// and output are added to the root span only.
func (l *Langfuse) traceAttrs(t *model.Trace, out any, isRoot bool) []attribute.KeyValue {
	if t == nil {
		return nil
	}
	var attrs []attribute.KeyValue
	addString := func(key, v string) {
		if v != "" {
			attrs = append(attrs, attribute.String(key, v))
		}
	}
	addString(attrTraceName, t.Name)
	addString(attrUserID, t.UserID)
	addString(attrSessionID, t.SessionID)
	addString(attrRelease, t.Release)
	addString(attrVersion, t.Version)
	env := t.Environment
	if env == "" {
		env = l.environment
	}
	addString(attrEnvironment, env)
	if len(t.Tags) > 0 {
		attrs = append(attrs, attribute.StringSlice(attrTraceTags, append([]string(nil), t.Tags...)))
	}
	if t.Public {
		attrs = append(attrs, attribute.Bool(attrTracePublic, true))
	}
	attrs = append(attrs, metadataAttrs(attrTraceMetadata, t.Metadata)...)

	if isRoot {
		if kv, ok := jsonAttr(attrTraceInput, t.Input); ok {
			attrs = append(attrs, kv)
		}
		if out == nil {
			out = t.Output
		}
		if kv, ok := jsonAttr(attrTraceOutput, out); ok {
			attrs = append(attrs, kv)
		}
	}
	return attrs
}

func observationAttrs(typ string, input, output, metadata any, level model.ObservationLevel, statusMessage string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{attribute.String(attrObservationType, typ)}
	if kv, ok := jsonAttr(attrObservationInput, input); ok {
		attrs = append(attrs, kv)
	}
	if kv, ok := jsonAttr(attrObservationOutput, output); ok {
		attrs = append(attrs, kv)
	}
	attrs = append(attrs, metadataAttrs(attrObservationMetadata, metadata)...)
	if level != "" {
		attrs = append(attrs, attribute.String(attrObservationLevel, string(level)))
	}
	if statusMessage != "" {
		attrs = append(attrs, attribute.String(attrObservationStatusMessage, statusMessage))
	}
	return attrs
}

func generationAttrs(g *model.Generation) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	if g.Model != "" {
		attrs = append(attrs, attribute.String(attrObservationModelName, g.Model))
	}
	if kv, ok := jsonAttr(attrObservationModelParameters, g.ModelParameters); ok {
		attrs = append(attrs, kv)
	}
	u := usageDetails(g.Usage)
	if u == "" && g.Level == model.ObservationLevelError {
		// Without explicit usage Langfuse estimates tokens (and cost) from the
		// input/output text; a failed call has no billed usage to estimate.
		u = `{"input":0,"output":0,"total":0}`
	}
	if u != "" {
		attrs = append(attrs, attribute.String(attrObservationUsageDetails, u))
	}
	if c := costDetails(g.Usage); c != "" {
		attrs = append(attrs, attribute.String(attrObservationCostDetails, c))
	}
	if g.CompletionStartTime != nil {
		attrs = append(attrs, attribute.String(attrObservationCompletionStartTime,
			g.CompletionStartTime.UTC().Format(time.RFC3339Nano)))
	}
	if g.PromptName != "" {
		attrs = append(attrs, attribute.String(attrObservationPromptName, g.PromptName))
	}
	if g.PromptVersion != 0 {
		attrs = append(attrs, attribute.Int(attrObservationPromptVersion, g.PromptVersion))
	}
	return attrs
}

// usageDetails returns the usage as {"input":n,"output":n,"total":n}, or ""
// when no token count is set. Input/Output/Total take precedence over the
// OpenAI-style PromptTokens/CompletionTokens/TotalTokens.
func usageDetails(u model.Usage) string {
	in, out, total := u.Input, u.Output, u.Total
	if in <= 0 && out <= 0 && total <= 0 {
		in, out, total = u.PromptTokens, u.CompletionTokens, u.TotalTokens
	}
	if in <= 0 && out <= 0 && total <= 0 {
		return ""
	}
	if total <= 0 {
		total = in + out
	}
	b, _ := json.Marshal(map[string]int{"input": in, "output": out, "total": total})
	return string(b)
}

// costDetails returns the explicit costs as {"input":x,"output":y,"total":z},
// or "" when none is set.
func costDetails(u model.Usage) string {
	if u.InputCost <= 0 && u.OutputCost <= 0 && u.TotalCost <= 0 {
		return ""
	}
	total := u.TotalCost
	if total <= 0 {
		total = u.InputCost + u.OutputCost
	}
	b, _ := json.Marshal(map[string]float64{"input": u.InputCost, "output": u.OutputCost, "total": total})
	return string(b)
}

func waitGroup(ctx context.Context, wg *sync.WaitGroup) error {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func nowIfNil(t *time.Time) *time.Time {
	if t != nil {
		return t
	}
	now := time.Now().UTC()
	return &now
}

func buildID(id *string) string {
	if id == nil || *id == "" {
		return uuid.New().String()
	}
	return *id
}
