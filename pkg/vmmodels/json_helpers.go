package vmmodels

import "encoding/json"

// MarshalJSONWithFallback marshals v into JSON.
// If marshalling fails, fallback is returned verbatim.
func MarshalJSONWithFallback(v interface{}, fallback json.RawMessage) json.RawMessage {
	data, err := json.Marshal(v)
	if err == nil {
		return data
	}
	if len(fallback) == 0 {
		return json.RawMessage("null")
	}
	out := make(json.RawMessage, len(fallback))
	copy(out, fallback)
	return out
}

// MarshalJSONStringWithFallback marshals v into a JSON string value.
// If marshalling fails, fallback is returned.
func MarshalJSONStringWithFallback(v interface{}, fallback json.RawMessage) string {
	return string(MarshalJSONWithFallback(v, fallback))
}
