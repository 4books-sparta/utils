package langfuse

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/4books-sparta/utils/chatbots/langfuse/model"
)

// TestOTLPExport sends a trace through the real OTLP/HTTP exporter to a fake
// Langfuse endpoint and checks the wire format: path, headers and attributes.
func TestOTLPExport(t *testing.T) {
	var (
		mu    sync.Mutex
		reqs  []*http.Request
		spans []*tracepb.Span
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var msg coltracepb.ExportTraceServiceRequest
		if err := proto.Unmarshal(body, &msg); err != nil {
			t.Errorf("unmarshal: %v", err)
		}
		mu.Lock()
		reqs = append(reqs, r)
		for _, rs := range msg.ResourceSpans {
			for _, ss := range rs.ScopeSpans {
				spans = append(spans, ss.Spans...)
			}
		}
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv("LANGFUSE_HOST", srv.URL)
	t.Setenv("LANGFUSE_PUBLIC_KEY", "pk-test")
	t.Setenv("LANGFUSE_SECRET_KEY", "sk-test")

	l := New(context.Background()).
		WithServiceName("export-test").
		WithEnvironment("test").
		WithFlushInterval(50 * time.Millisecond)

	const traceID = "2f1c9a3e-4b5d-4c6e-8f7a-9b0c1d2e3f4a"
	tr, err := l.Trace(&model.Trace{ID: traceID, Name: "qa", UserID: "42", SessionID: "sess", Tags: []string{"a"}})
	if err != nil {
		t.Fatal(err)
	}
	root := &model.Span{TraceID: tr.ID, Name: "StreamChat"}
	if _, err := l.Span(root, nil); err != nil {
		t.Fatal(err)
	}
	gen := &model.Generation{TraceID: tr.ID, Name: "streamQ", Model: "gpt-test", Usage: model.Usage{Input: 10, Output: 5}}
	if _, err := l.Generation(gen, &root.ID); err != nil {
		t.Fatal(err)
	}
	gen.Output = "answer"
	if _, err := l.GenerationEnd(gen); err != nil {
		t.Fatal(err)
	}
	l.SetTraceOutput(tr.ID, "answer")
	if _, err := l.SpanEnd(root); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := l.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(reqs) == 0 {
		t.Fatal("no export request received")
	}
	r := reqs[0]
	if r.URL.Path != "/api/public/otel/v1/traces" {
		t.Errorf("path = %s", r.URL.Path)
	}
	if got := r.Header.Get("x-langfuse-ingestion-version"); got != "4" {
		t.Errorf("ingestion version header = %q", got)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("pk-test:sk-test"))
	if got := r.Header.Get("Authorization"); got != wantAuth {
		t.Errorf("authorization = %q", got)
	}
	if ct := r.Header.Get("Content-Type"); ct != "application/x-protobuf" {
		t.Errorf("content-type = %q", ct)
	}
	if len(spans) != 2 {
		t.Fatalf("spans = %d, want 2", len(spans))
	}
	byName := map[string]*tracepb.Span{}
	for _, s := range spans {
		byName[s.Name] = s
		if got := traceIDHex(s.TraceId); got != "2f1c9a3e4b5d4c6e8f7a9b0c1d2e3f4a" {
			t.Errorf("%s trace id = %s", s.Name, got)
		}
	}
	rootSpan, genSpan := byName["StreamChat"], byName["streamQ"]
	if rootSpan == nil || genSpan == nil {
		t.Fatalf("missing spans: %v", byName)
	}
	if len(rootSpan.ParentSpanId) != 0 {
		t.Error("root has a parent")
	}
	if string(genSpan.ParentSpanId) != string(rootSpan.SpanId) {
		t.Error("generation parent is not the root")
	}
	for name, s := range byName {
		a := pbAttrs(s)
		if a["langfuse.user.id"] != "42" || a["langfuse.session.id"] != "sess" || a["langfuse.trace.name"] != "qa" {
			t.Errorf("%s: trace attributes missing: %v", name, a)
		}
	}
	ra := pbAttrs(rootSpan)
	if ra["langfuse.trace.output"] != "answer" || ra["langfuse.observation.type"] != "span" {
		t.Errorf("root attrs: %v", ra)
	}
	ga := pbAttrs(genSpan)
	if ga["langfuse.observation.type"] != "generation" || ga["langfuse.observation.model.name"] != "gpt-test" ||
		ga["langfuse.observation.usage_details"] != `{"input":10,"output":5,"total":15}` {
		t.Errorf("generation attrs: %v", ga)
	}
	if _, ok := ga["langfuse.trace.output"]; ok {
		t.Error("trace output must be on the root only")
	}
}

func traceIDHex(b []byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, 0, 2*len(b))
	for _, c := range b {
		out = append(out, hexdigits[c>>4], hexdigits[c&0x0f])
	}
	return string(out)
}

// pbAttrs flattens the span attributes: strings as they are, other kinds
// through their protobuf text form.
func pbAttrs(s *tracepb.Span) map[string]string {
	m := map[string]string{}
	for _, kv := range s.Attributes {
		if sv, ok := kv.Value.GetValue().(*commonpb.AnyValue_StringValue); ok {
			m[kv.Key] = sv.StringValue
			continue
		}
		m[kv.Key] = kv.Value.String()
	}
	return m
}
