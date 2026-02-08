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
