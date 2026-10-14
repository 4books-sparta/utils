package langfuse

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/4books-sparta/utils/chatbots/langfuse/model"
)

func newTestClient(t *testing.T) (*Langfuse, *tracetest.InMemoryExporter) {
	t.Helper()
	exp := tracetest.NewInMemoryExporter()
	l := New(context.Background()).
		WithServiceName("test-service").
		WithEnvironment("test").
		WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(keepOnShutdown{exp}))
	t.Cleanup(func() { _ = l.Shutdown(context.Background()) })
	return l, exp
}

// keepOnShutdown keeps the exported spans readable after the provider shuts
// down (InMemoryExporter.Shutdown resets them).
type keepOnShutdown struct{ *tracetest.InMemoryExporter }

func (keepOnShutdown) Shutdown(context.Context) error { return nil }

func attrMap(s tracetest.SpanStub) map[attribute.Key]attribute.Value {
	m := make(map[attribute.Key]attribute.Value, len(s.Attributes))
	for _, kv := range s.Attributes {
		m[kv.Key] = kv.Value
	}
	return m
}

func spansByName(exp *tracetest.InMemoryExporter) map[string]tracetest.SpanStub {
	out := map[string]tracetest.SpanStub{}
	for _, s := range exp.GetSpans() {
		out[s.Name] = s
	}
	return out
}

func ptr(s string) *string { return &s }

// TestChatFlow reproduces the observations emitted by the chatbot endpoints.
func TestChatFlow(t *testing.T) {
	l, exp := newTestClient(t)

	tr, err := l.Trace(&model.Trace{
		ID:        testUUID,
		Name:      "qa",
		UserID:    "42",
		SessionID: "sess-1",
		Tags:      []string{"books"},
		Metadata:  map[string]string{"bot": "aibot-1"},
		Input:     map[string]string{"q": "hello"},
	})
	if err != nil {
		t.Fatalf("Trace: %v", err)
	}

	pid := "0b6c2a1e-1111-4222-8333-444455556666"
	root := &model.Span{ID: pid, Name: "StreamChat"}
	root.TraceID = tr.ID
	if _, err := l.Span(root, nil); err != nil {
		t.Fatalf("Span root: %v", err)
	}
	// The caller adds tags to the registered trace after Trace(): they must
	// reach every span ended afterwards.
	tr.Tags = append(tr.Tags, "stream")

	prep := &model.Span{TraceID: tr.ID, Name: "PrepareConversation", ParentObservationID: pid}
	if _, err := l.Span(prep, nil); err != nil {
		t.Fatalf("Span prep: %v", err)
	}
	if prep.ID == "" || prep.StartTime == nil {
		t.Fatal("Span should assign ID and start time")
	}

	gen := &model.Generation{
		TraceID:         tr.ID,
		Name:            "streamQ",
		Model:           "claude-test",
		ModelParameters: model.M{"temperature": 0.2},
		Input:           []model.M{{"role": "user", "content": "hello"}},
	}
	if _, err := l.Generation(gen, ptr(pid)); err != nil {
		t.Fatalf("Generation: %v", err)
	}

	ev := &model.Event{TraceID: tr.ID, Name: "error", Level: model.ObservationLevelError, StatusMessage: "boom"}
	if _, err := l.Event(ev, ptr(pid)); err != nil {
		t.Fatalf("Event: %v", err)
	}

	l.SetTraceOutput(tr.ID, "the answer")

	gen.Output = "the answer"
	gen.Usage = model.Usage{Input: 10, Output: 5}
	cst := time.Now().UTC()
	gen.CompletionStartTime = &cst
	if _, err := l.GenerationEnd(gen); err != nil {
		t.Fatalf("GenerationEnd: %v", err)
	}
	if _, err := l.SpanEnd(prep); err != nil {
		t.Fatalf("SpanEnd prep: %v", err)
	}
	if _, err := l.SpanEnd(root); err != nil {
		t.Fatalf("SpanEnd root: %v", err)
	}
	l.Flush(context.Background())

	spans := exp.GetSpans()
	if len(spans) != 4 {
		t.Fatalf("expected 4 spans, got %d", len(spans))
	}
	byName := spansByName(exp)
	for _, n := range []string{"StreamChat", "PrepareConversation", "streamQ", "error"} {
		if _, ok := byName[n]; !ok {
			t.Fatalf("missing span %q", n)
		}
	}

	wantTID := strings.ReplaceAll(testUUID, "-", "")
	rootStub := byName["StreamChat"]
	rootSID := rootStub.SpanContext.SpanID()
	for name, s := range byName {
		if got := s.SpanContext.TraceID().String(); got != wantTID {
			t.Errorf("%s: trace ID %s, want %s", name, got, wantTID)
		}
		a := attrMap(s)
		if a[attrUserID].AsString() != "42" || a[attrSessionID].AsString() != "sess-1" {
			t.Errorf("%s: missing user/session: %v", name, a)
		}
		if a[attrTraceName].AsString() != "qa" {
			t.Errorf("%s: missing trace name", name)
		}
		if a[attrEnvironment].AsString() != "test" {
			t.Errorf("%s: environment %q", name, a[attrEnvironment].AsString())
		}
		if a["langfuse.trace.metadata.bot"].AsString() != "aibot-1" {
			t.Errorf("%s: missing trace metadata", name)
		}
		tags := a[attrTraceTags]
		if tags.Type() != attribute.STRINGSLICE {
			t.Errorf("%s: tags type %v", name, tags.Type())
		} else if got := tags.AsStringSlice(); len(got) != 2 || got[0] != "books" || got[1] != "stream" {
			t.Errorf("%s: tags %v", name, got)
		}
		if name == "StreamChat" {
			continue
		}
		if _, ok := a[attrTraceInput]; ok {
			t.Errorf("%s: trace input must be on the root only", name)
		}
		if _, ok := a[attrTraceOutput]; ok {
			t.Errorf("%s: trace output must be on the root only", name)
		}
		if got := s.Parent.SpanID(); got != rootSID {
			t.Errorf("%s: parent %s, want root %s", name, got, rootSID)
		}
	}

	if rootStub.Parent.IsValid() {
		t.Error("root should have no parent")
	}
	ra := attrMap(rootStub)
	if ra[attrTraceInput].AsString() != `{"q":"hello"}` {
		t.Errorf("trace input %q", ra[attrTraceInput].AsString())
	}
	if ra[attrTraceOutput].AsString() != "the answer" {
		t.Errorf("trace output %q", ra[attrTraceOutput].AsString())
	}
	if ra[attrObservationType].AsString() != observationTypeSpan {
		t.Errorf("root type %q", ra[attrObservationType].AsString())
	}

	ga := attrMap(byName["streamQ"])
	if ga[attrObservationType].AsString() != observationTypeGeneration || ga[attrObservationModelName].AsString() != "claude-test" {
		t.Errorf("generation attrs: %v", ga)
	}
	if ga[attrObservationOutput].AsString() != "the answer" {
		t.Errorf("generation output %q", ga[attrObservationOutput].AsString())
	}
	if ga[attrObservationModelParameters].AsString() != `{"temperature":0.2}` {
		t.Errorf("model parameters %q", ga[attrObservationModelParameters].AsString())
	}
	if ga[attrObservationUsageDetails].AsString() != `{"input":10,"output":5,"total":15}` {
		t.Errorf("usage details %q", ga[attrObservationUsageDetails].AsString())
	}
	if ga[attrObservationCompletionStartTime].AsString() != cst.Format(time.RFC3339Nano) {
		t.Errorf("completion start time %q", ga[attrObservationCompletionStartTime].AsString())
	}

	evStub := byName["error"]
	ea := attrMap(evStub)
	if ea[attrObservationType].AsString() != observationTypeEvent || ea[attrObservationLevel].AsString() != "ERROR" {
		t.Errorf("event attrs: %v", ea)
	}
	if evStub.Status.Code != codes.Error || evStub.Status.Description != "boom" {
		t.Errorf("event status %+v", evStub.Status)
	}
	if !evStub.StartTime.Equal(evStub.EndTime) {
		t.Errorf("event start %v != end %v", evStub.StartTime, evStub.EndTime)
	}
}

func TestSpanWithoutParentAttachesToRoot(t *testing.T) {
	l, exp := newTestClient(t)

	tr, _ := l.Trace(&model.Trace{ID: testUUID, Name: "qa"})
	root := &model.Span{TraceID: tr.ID, Name: "root"}
	l.Span(root, nil)
	other := &model.Span{TraceID: tr.ID, Name: "other"}
	l.Span(other, nil)
	if _, err := l.SpanEnd(other); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SpanEnd(root); err != nil {
		t.Fatal(err)
	}

	byName := spansByName(exp)
	if got, want := byName["other"].Parent.SpanID(), byName["root"].SpanContext.SpanID(); got != want {
		t.Fatalf("parent %s, want root %s", got, want)
	}
}

func TestEndUnknownIDFails(t *testing.T) {
	l, _ := newTestClient(t)

	if _, err := l.SpanEnd(&model.Span{ID: "unknown"}); err == nil {
		t.Error("SpanEnd on unknown ID should fail")
	}
	if _, err := l.SpanEnd(&model.Span{}); err == nil {
		t.Error("SpanEnd without ID should fail")
	}
	if _, err := l.GenerationEnd(&model.Generation{ID: "unknown"}); err == nil {
		t.Error("GenerationEnd on unknown ID should fail")
	}
}

func TestObservationWithoutTraceCreatesTrace(t *testing.T) {
	l, exp := newTestClient(t)

	s, err := l.Span(&model.Span{Name: "standalone"}, nil)
	if err != nil {
		t.Fatalf("Span: %v", err)
	}
	if s.TraceID == "" {
		t.Fatal("expected a trace ID")
	}
	if _, err := l.SpanEnd(s); err != nil {
		t.Fatal(err)
	}
	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if got, want := spans[0].SpanContext.TraceID().String(), strings.ReplaceAll(s.TraceID, "-", ""); got != want {
		t.Errorf("trace ID %s, want %s", got, want)
	}
	if attrMap(spans[0])[attrTraceName].AsString() != "standalone" {
		t.Error("trace should be named after the observation")
	}
}

func TestShutdownEndsOpenSpans(t *testing.T) {
	l, exp := newTestClient(t)

	tr, _ := l.Trace(&model.Trace{ID: testUUID, Name: "qa"})
	l.Span(&model.Span{TraceID: tr.ID, Name: "open"}, nil)

	l.Flush(context.Background())
	if n := len(exp.GetSpans()); n != 0 {
		t.Fatalf("Flush must not end open spans, got %d exported", n)
	}

	if err := l.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if err := l.Shutdown(context.Background()); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status.Code != codes.Error || spans[0].Status.Description != unterminatedMessage {
		t.Errorf("status %+v", spans[0].Status)
	}
}

func TestScoreWithoutKeyIsNoop(t *testing.T) {
	t.Setenv("LANGFUSE_PUBLIC_KEY", "")
	l, _ := newTestClient(t)

	s, err := l.Score(&model.Score{ID: "s_" + testUUID, TraceID: testUUID, Name: "like", Value: 1})
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	if s.TraceID != strings.ReplaceAll(testUUID, "-", "") {
		t.Errorf("trace ID %s not converted", s.TraceID)
	}
	if s.ID != "s_"+testUUID {
		t.Errorf("score ID changed to %s", s.ID)
	}

	s, err = l.Score(&model.Score{TraceID: "msg_123", Name: "like"})
	if err != nil || s.TraceID != "msg_123" {
		t.Errorf("non-UUID trace ID: %v, %s", err, s.TraceID)
	}
	if _, err := l.Score(&model.Score{Name: "like"}); err == nil {
		t.Error("Score without trace ID should fail")
	}
}

func TestUsageDetails(t *testing.T) {
	cases := []struct {
		u    model.Usage
		want string
	}{
		{model.Usage{}, ""},
		{model.Usage{Input: 3, Output: 4, Total: 8}, `{"input":3,"output":4,"total":8}`},
		{model.Usage{PromptTokens: 2, CompletionTokens: 1}, `{"input":2,"output":1,"total":3}`},
		{model.Usage{Input: 5}, `{"input":5,"output":0,"total":5}`},
	}
	for _, c := range cases {
		if got := usageDetails(c.u); got != c.want {
			t.Errorf("usageDetails(%+v) = %q, want %q", c.u, got, c.want)
		}
	}
}

// TestFailedGenerationReportsZeroUsage: a generation ended with level ERROR
// and no usage reports zero tokens explicitly, so Langfuse does not estimate
// tokens and cost from the input/output text.
func TestFailedGenerationReportsZeroUsage(t *testing.T) {
	l, exp := newTestClient(t)
	tr, err := l.Trace(&model.Trace{Name: "qa"})
	if err != nil {
		t.Fatalf("Trace: %v", err)
	}
	g, err := l.Generation(&model.Generation{TraceID: tr.ID, Name: "embedding", Model: "m"}, nil)
	if err != nil {
		t.Fatalf("Generation: %v", err)
	}
	g.Level = model.ObservationLevelError
	g.StatusMessage = "401 Unauthorized"
	g.Output = "401 Unauthorized"
	if _, err := l.GenerationEnd(g); err != nil {
		t.Fatalf("GenerationEnd: %v", err)
	}
	l.Flush(context.Background())

	a := attrMap(spansByName(exp)["embedding"])
	if got := a[attrObservationUsageDetails].AsString(); got != `{"input":0,"output":0,"total":0}` {
		t.Errorf("usage_details = %q", got)
	}
	if got := a[attrObservationLevel].AsString(); got != "ERROR" {
		t.Errorf("level = %q", got)
	}
	if got := a[attrObservationStatusMessage].AsString(); got != "401 Unauthorized" {
		t.Errorf("status_message = %q", got)
	}
}
