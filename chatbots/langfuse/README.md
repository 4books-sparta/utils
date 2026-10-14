# langfuse

Go client used by the chatbots to send traces and scores to [Langfuse](https://langfuse.com) (v4).

- Traces, spans, generations and events are exported as OpenTelemetry spans
  over OTLP/HTTP to `<LANGFUSE_HOST>/api/public/otel/v1/traces`, with the
  `x-langfuse-ingestion-version: 4` header so they are ingested in real time.
  Spans are batched by the OTel `BatchSpanProcessor`.
- Scores are still sent to the public ingestion API (`score-create` events),
  one asynchronous request per score.

## Configuration

| Variable | Description |
| --- | --- |
| `LANGFUSE_HOST` | Langfuse base URL, default `https://cloud.langfuse.com` |
| `LANGFUSE_PUBLIC_KEY` | Project public key. When empty, tracing and scores are disabled (no-op) |
| `LANGFUSE_SECRET_KEY` | Project secret key |

Set `DEBUG` to log score ingestion errors. Exporter errors are always logged.

## Usage

```go
lf := langfuse.New(ctx).
	WithServiceName("chatbot-ms").
	WithEnvironment("production").
	WithFlushInterval(time.Second)
defer lf.Shutdown(context.Background())

tr, _ := lf.Trace(&model.Trace{
	ID:        messageID, // UUID
	Name:      "qa",
	UserID:    "42",
	SessionID: sessionID,
	Input:     map[string]string{"q": question},
})

root := &model.Span{TraceID: tr.ID, Name: "StreamChat"}
lf.Span(root, nil)

gen := &model.Generation{TraceID: tr.ID, Name: "streamQ", Model: "claude-sonnet-4-5", Input: messages}
lf.Generation(gen, &root.ID)

lf.Event(&model.Event{TraceID: tr.ID, Name: "error", Level: model.ObservationLevelError, StatusMessage: err.Error()}, &root.ID)

gen.Output = answer
gen.Usage = model.Usage{Input: inTokens, Output: outTokens}
lf.GenerationEnd(gen)

lf.SetTraceOutput(tr.ID, answer)
lf.SpanEnd(root)

lf.Score(&model.Score{ID: "s_" + messageID, TraceID: messageID, Name: "like", Value: 1})
```

The `With*` options must be applied before the first trace: the OTel
provider is built on first use.

## Notes

- **Trace IDs.** A UUID trace ID is mapped to the OTel trace ID with the same
  16 bytes (lowercase, no dashes), so `2f1c9a3e-4b5d-...` appears in Langfuse
  as `2f1c9a3e4b5d...`. `Score` applies the same conversion, so scores attach
  to the right trace. Non-UUID IDs get a random OTel trace ID.
- **Observation IDs** are local keys mapping model IDs to live spans; Langfuse
  shows the OTel span IDs. An observation without a known parent is attached
  to the root span of its trace.
- **Trace attributes** (name, user, session, tags, metadata, release, version,
  environment, public) are read from the `*model.Trace` passed to `Trace` when
  each observation ends and copied onto every span. Trace input and output are
  set on the root span only.
- **Flush** exports the spans that have ended. **Shutdown** ends the spans
  still open (level `ERROR`, status message `unterminated`), flushes, waits for
  pending scores and stops the provider: call it on SIGTERM, otherwise the
  last batch is lost.
