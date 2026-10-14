package langfuse

import (
	"encoding/json"
	"fmt"
	"sort"

	"go.opentelemetry.io/otel/attribute"

	"github.com/4books-sparta/utils/chatbots/langfuse/model"
)

// Langfuse OTel attribute names, see
// https://langfuse.com/integrations/native/opentelemetry
const (
	attrTraceName     = "langfuse.trace.name"
	attrTraceInput    = "langfuse.trace.input"
	attrTraceOutput   = "langfuse.trace.output"
	attrTraceTags     = "langfuse.trace.tags"
	attrTraceMetadata = "langfuse.trace.metadata"
	attrTracePublic   = "langfuse.trace.public"
	attrUserID        = "langfuse.user.id"
	attrSessionID     = "langfuse.session.id"
	attrRelease       = "langfuse.release"
	attrVersion       = "langfuse.version"
	attrEnvironment   = "langfuse.environment"

	attrObservationType                = "langfuse.observation.type"
	attrObservationInput               = "langfuse.observation.input"
	attrObservationOutput              = "langfuse.observation.output"
	attrObservationMetadata            = "langfuse.observation.metadata"
	attrObservationLevel               = "langfuse.observation.level"
	attrObservationStatusMessage       = "langfuse.observation.status_message"
	attrObservationModelName           = "langfuse.observation.model.name"
	attrObservationModelParameters     = "langfuse.observation.model.parameters"
	attrObservationUsageDetails        = "langfuse.observation.usage_details"
	attrObservationCostDetails         = "langfuse.observation.cost_details"
	attrObservationCompletionStartTime = "langfuse.observation.completion_start_time"
	attrObservationPromptName          = "langfuse.observation.prompt.name"
	attrObservationPromptVersion       = "langfuse.observation.prompt.version"
)

// Values of langfuse.observation.type.
const (
	observationTypeSpan       = "span"
	observationTypeGeneration = "generation"
	observationTypeEvent      = "event"
)

// jsonAttr serializes v as a string attribute. Strings are used as they are,
// other values are JSON-encoded (falling back to %v if encoding fails).
// It returns false when there is nothing to record (nil, "" or JSON null).
func jsonAttr(key string, v any) (attribute.KeyValue, bool) {
	s := stringify(v)
	if s == "" || s == "null" {
		return attribute.KeyValue{}, false
	}
	return attribute.String(key, s), true
}

// metadataAttrs flattens string-keyed maps into prefix.<key> attributes, in
// key order. Any other value is recorded as a single JSON attribute on prefix.
func metadataAttrs(prefix string, meta any) []attribute.KeyValue {
	switch m := meta.(type) {
	case nil:
		return nil
	case map[string]string:
		keys := sortedKeys(m)
		out := make([]attribute.KeyValue, 0, len(keys))
		for _, k := range keys {
			out = append(out, attribute.String(prefix+"."+k, m[k]))
		}
		return out
	case model.M:
		return metadataAttrs(prefix, map[string]any(m))
	case map[string]any:
		keys := sortedKeys(m)
		out := make([]attribute.KeyValue, 0, len(keys))
		for _, k := range keys {
			if kv, ok := jsonAttr(prefix+"."+k, m[k]); ok {
				out = append(out, kv)
			}
		}
		return out
	default:
		if kv, ok := jsonAttr(prefix, meta); ok {
			return []attribute.KeyValue{kv}
		}
		return nil
	}
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case *string:
		if t == nil {
			return ""
		}
		return *t
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
