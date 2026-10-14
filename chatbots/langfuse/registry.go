package langfuse

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/4books-sparta/utils/chatbots/langfuse/model"
)

const unterminatedMessage = "unterminated"

// liveSpan is an OTel span started and not yet ended, keyed by model ID.
type liveSpan struct {
	ctx     context.Context
	span    trace.Span
	traceID string
	isRoot  bool
}

// traceState is the client-side state of a Langfuse trace.
type traceState struct {
	model  *model.Trace
	rootID string // model ID of the root span
	output any

	rootSC trace.SpanContext // span context of the root, valid once started
	live   int               // spans of this trace still open
}

type registry struct {
	tracer trace.Tracer

	mu     sync.Mutex
	spans  map[string]liveSpan    // key: model ID (UUID)
	traces map[string]*traceState // key: model.Trace.ID (UUID)
}

func newRegistry(tracer trace.Tracer) *registry {
	return &registry{
		tracer: tracer,
		spans:  map[string]liveSpan{},
		traces: map[string]*traceState{},
	}
}

// putTrace registers or replaces the model of a trace. No span is emitted:
// the OTel trace starts with its first span.
func (r *registry) putTrace(t *model.Trace) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ts, ok := r.traces[t.ID]; ok {
		ts.model = t
		return
	}
	r.traces[t.ID] = &traceState{model: t}
}

// setTraceOutput stores the trace output; it returns false if the trace is unknown.
func (r *registry) setTraceOutput(traceID string, out any) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	ts, ok := r.traces[traceID]
	if !ok {
		return false
	}
	ts.output = out
	return true
}

// traceView returns the trace model and the explicit output of ts, read
// under the registry lock.
func (r *registry) traceView(ts *traceState) (*model.Trace, any) {
	if ts == nil {
		return nil, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return ts.model, ts.output
}

// startSpan starts the OTel span for model ID id and registers it.
// Parent resolution:
//   - parentID live in the registry: that span;
//   - otherwise, if the trace already has a root: the root;
//   - otherwise the span becomes the root, with the trace ID derived from traceID.
func (r *registry) startSpan(traceID, id, parentID, name string, startTime time.Time) liveSpan {
	r.mu.Lock()
	defer r.mu.Unlock()

	ts, ok := r.traces[traceID]
	if !ok {
		ts = &traceState{model: &model.Trace{ID: traceID}}
		r.traces[traceID] = ts
	}

	var parentCtx context.Context
	isRoot := false
	if p, ok := r.spans[parentID]; ok && parentID != "" {
		parentCtx = p.ctx
	} else if ts.rootSC.IsValid() {
		parentCtx = trace.ContextWithSpanContext(context.Background(), ts.rootSC)
	} else {
		isRoot = true
		parentCtx = context.Background()
		if tid, ok := model.TraceIDFromUUID(traceID); ok {
			parentCtx = contextWithTraceID(parentCtx, tid)
		}
	}

	opts := []trace.SpanStartOption{trace.WithTimestamp(startTime)}
	if isRoot {
		opts = append(opts, trace.WithNewRoot())
	}
	ctx, span := r.tracer.Start(parentCtx, name, opts...)

	if isRoot {
		ts.rootID = id
		ts.rootSC = span.SpanContext()
	}
	ts.live++

	ls := liveSpan{ctx: ctx, span: span, traceID: traceID, isRoot: isRoot}
	r.spans[id] = ls
	return ls
}

// endSpan removes the span with model ID id from the registry and returns it
// with its trace state. The caller sets the final attributes and ends it.
// The trace state is dropped once all its spans are closed.
func (r *registry) endSpan(id string) (liveSpan, *traceState, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ls, ok := r.spans[id]
	if !ok {
		return liveSpan{}, nil, false
	}
	delete(r.spans, id)

	ts := r.traces[ls.traceID]
	if ts != nil {
		ts.live--
		if ts.live <= 0 {
			delete(r.traces, ls.traceID)
		}
	}
	return ls, ts, true
}

// endAll ends every live span with an error status and clears the registry.
// It returns the number of spans ended.
func (r *registry) endAll() int {
	r.mu.Lock()
	spans := r.spans
	r.spans = map[string]liveSpan{}
	r.traces = map[string]*traceState{}
	r.mu.Unlock()

	for _, ls := range spans {
		ls.span.SetAttributes(
			attribute.String(attrObservationLevel, string(model.ObservationLevelError)),
			attribute.String(attrObservationStatusMessage, unterminatedMessage),
		)
		ls.span.SetStatus(codes.Error, unterminatedMessage)
		ls.span.End()
	}
	return len(spans)
}
