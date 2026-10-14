package langfuse

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/4books-sparta/utils/chatbots/langfuse/model"
)

const testUUID = "2f1c9a3e-4b5d-4e6f-8a7b-9c0d1e2f3a4b"

func newTestProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.InMemoryExporter) {
	t.Helper()
	exp := tracetest.NewInMemoryExporter()
	tp, err := newTracerProvider(context.Background(), providerConfig{
		serviceName:   "test-service",
		environment:   "test",
		spanProcessor: sdktrace.NewSimpleSpanProcessor(exp),
	})
	if err != nil {
		t.Fatalf("newTracerProvider: %v", err)
	}
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return tp, exp
}

func TestProviderStarts(t *testing.T) {
	tp, exp := newTestProvider(t)

	_, span := tp.Tracer(tracerName).Start(context.Background(), "s")
	span.End()

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	res := spans[0].Resource
	found := map[string]string{}
	for _, kv := range res.Attributes() {
		found[string(kv.Key)] = kv.Value.Emit()
	}
	if found["service.name"] != "test-service" || found["deployment.environment"] != "test" {
		t.Fatalf("unexpected resource attributes: %v", found)
	}
}

func TestProviderNoKeyIsNoop(t *testing.T) {
	t.Setenv("LANGFUSE_PUBLIC_KEY", "")
	tp, err := newTracerProvider(context.Background(), providerConfig{})
	if err != nil {
		t.Fatalf("newTracerProvider: %v", err)
	}
	defer tp.Shutdown(context.Background())

	_, span := tp.Tracer(tracerName).Start(context.Background(), "s")
	span.End()
	if err := tp.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush: %v", err)
	}
}

func TestTraceIDFromUUID(t *testing.T) {
	id, ok := model.TraceIDFromUUID(testUUID)
	if !ok {
		t.Fatal("expected a valid trace ID")
	}
	if got, want := id.String(), "2f1c9a3e4b5d4e6f8a7b9c0d1e2f3a4b"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	upper, ok := model.TraceIDFromUUID("2F1C9A3E-4B5D-4E6F-8A7B-9C0D1E2F3A4B")
	if !ok || upper != id {
		t.Fatalf("uppercase UUID: got %s, %v", upper, ok)
	}

	for _, bad := range []string{"", "not-a-uuid", "msg_123", "00000000-0000-0000-0000-000000000000"} {
		if _, ok := model.TraceIDFromUUID(bad); ok {
			t.Errorf("expected %q to be rejected", bad)
		}
	}
}

func TestIDGeneratorFallsBackToRandom(t *testing.T) {
	g := idGenerator{}

	// No trace ID in context.
	a, sa := g.NewIDs(context.Background())
	b, sb := g.NewIDs(context.Background())
	if !a.IsValid() || !b.IsValid() || a == b {
		t.Fatalf("expected two distinct random trace IDs, got %s and %s", a, b)
	}
	if !sa.IsValid() || !sb.IsValid() || sa == sb {
		t.Fatalf("expected two distinct random span IDs, got %s and %s", sa, sb)
	}

	// A non-UUID ID yields no trace ID to put in the context: still random.
	tid, ok := model.TraceIDFromUUID("not-a-uuid")
	ctx := context.Background()
	if ok {
		ctx = contextWithTraceID(ctx, tid)
	}
	c, _ := g.NewIDs(ctx)
	if !c.IsValid() || c == a || c == b {
		t.Fatalf("expected a new random trace ID, got %s", c)
	}

	// An invalid (zero) trace ID in the context is ignored.
	d, _ := g.NewIDs(contextWithTraceID(context.Background(), trace.TraceID{}))
	if !d.IsValid() {
		t.Fatal("expected a valid random trace ID for a zero desired ID")
	}
}

func TestSpanUsesDesiredTraceID(t *testing.T) {
	tp, exp := newTestProvider(t)

	want, _ := model.TraceIDFromUUID(testUUID)
	ctx := contextWithTraceID(context.Background(), want)
	ctx, root := tp.Tracer(tracerName).Start(ctx, "root")
	_, child := tp.Tracer(tracerName).Start(ctx, "child")
	child.End()
	root.End()

	spans := exp.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}
	for _, s := range spans {
		if s.SpanContext.TraceID() != want {
			t.Errorf("span %s: trace ID %s, want %s", s.Name, s.SpanContext.TraceID(), want)
		}
	}
}

func TestRegistry(t *testing.T) {
	tp, exp := newTestProvider(t)
	r := newRegistry(tp.Tracer(tracerName))

	r.putTrace(&model.Trace{ID: testUUID, Name: "qa"})
	now := time.Now()

	root := r.startSpan(testUUID, "root", "", "StreamChat", now)
	if !root.isRoot {
		t.Fatal("first span without parent should be the root")
	}
	child := r.startSpan(testUUID, "child", "root", "PrepareConversation", now)
	orphan := r.startSpan(testUUID, "orphan", "", "streamQ", now)
	if child.isRoot || orphan.isRoot {
		t.Fatal("only the first span should be the root")
	}

	wantTID, _ := model.TraceIDFromUUID(testUUID)
	rootSC := root.span.SpanContext()
	if rootSC.TraceID() != wantTID {
		t.Fatalf("root trace ID %s, want %s", rootSC.TraceID(), wantTID)
	}

	for _, id := range []string{"orphan", "child", "root"} {
		ls, ts, ok := r.endSpan(id)
		if !ok {
			t.Fatalf("endSpan(%s): not found", id)
		}
		if ts == nil || ts.rootID != "root" {
			t.Fatalf("endSpan(%s): unexpected trace state %+v", id, ts)
		}
		ls.span.End()
		if _, _, ok := r.endSpan(id); ok {
			t.Fatalf("endSpan(%s): entry still present", id)
		}
	}
	if len(r.spans) != 0 || len(r.traces) != 0 {
		t.Fatalf("registry not empty: %d spans, %d traces", len(r.spans), len(r.traces))
	}

	byName := map[string]tracetest.SpanStub{}
	for _, s := range exp.GetSpans() {
		byName[s.Name] = s
	}
	if len(byName) != 3 {
		t.Fatalf("expected 3 exported spans, got %d", len(byName))
	}
	if byName["StreamChat"].Parent.IsValid() {
		t.Error("root should have no parent")
	}
	if got := byName["PrepareConversation"].Parent.SpanID(); got != rootSC.SpanID() {
		t.Errorf("child parent %s, want root %s", got, rootSC.SpanID())
	}
	if got := byName["streamQ"].Parent.SpanID(); got != rootSC.SpanID() {
		t.Errorf("orphan parent %s, want root %s", got, rootSC.SpanID())
	}
	for name, s := range byName {
		if s.SpanContext.TraceID() != wantTID {
			t.Errorf("%s: trace ID %s, want %s", name, s.SpanContext.TraceID(), wantTID)
		}
	}
}

func TestRegistryEndAll(t *testing.T) {
	tp, exp := newTestProvider(t)
	r := newRegistry(tp.Tracer(tracerName))

	r.startSpan(testUUID, "root", "", "root", time.Now())
	r.startSpan(testUUID, "child", "root", "child", time.Now())

	if n := r.endAll(); n != 2 {
		t.Fatalf("endAll ended %d spans, want 2", n)
	}
	spans := exp.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("expected 2 exported spans, got %d", len(spans))
	}
	for _, s := range spans {
		if s.Status.Code != codes.Error || s.Status.Description != unterminatedMessage {
			t.Errorf("%s: status %+v", s.Name, s.Status)
		}
	}
	if len(r.spans) != 0 || len(r.traces) != 0 {
		t.Fatal("registry not empty after endAll")
	}
}

func TestMetadataAttrs(t *testing.T) {
	got := metadataAttrs(attrTraceMetadata, map[string]string{"b": "2", "a": "1"})
	if len(got) != 2 || string(got[0].Key) != "langfuse.trace.metadata.a" || got[1].Value.AsString() != "2" {
		t.Fatalf("unexpected attrs: %v", got)
	}

	got = metadataAttrs(attrObservationMetadata, model.M{"n": 3, "s": "x"})
	if len(got) != 2 || got[0].Value.AsString() != "3" || got[1].Value.AsString() != "x" {
		t.Fatalf("unexpected attrs: %v", got)
	}

	got = metadataAttrs(attrObservationMetadata, []int{1, 2})
	if len(got) != 1 || string(got[0].Key) != attrObservationMetadata || got[0].Value.AsString() != "[1,2]" {
		t.Fatalf("unexpected attrs: %v", got)
	}

	if _, ok := jsonAttr(attrObservationInput, ""); ok {
		t.Error("empty string should be omitted")
	}
	if _, ok := jsonAttr(attrObservationInput, nil); ok {
		t.Error("nil should be omitted")
	}
	if kv, ok := jsonAttr(attrObservationInput, map[string]string{"q": "hi"}); !ok || kv.Value.AsString() != `{"q":"hi"}` {
		t.Errorf("unexpected json attr: %v", kv)
	}
}
