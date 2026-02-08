package vmcontrol

import (
	"context"
	"sort"
)

// RuntimeRegistry exposes active in-memory runtime visibility for ops/health.
type RuntimeRegistry struct {
	runtime SessionRuntimePort
}

func NewRuntimeRegistry(runtime SessionRuntimePort) *RuntimeRegistry {
	return &RuntimeRegistry{runtime: runtime}
}

func (r *RuntimeRegistry) Summary(_ context.Context) RuntimeSummary {
	active := r.runtime.ListSessions()
	sessionIDs := make([]string, 0, len(active))
	for _, session := range active {
		sessionIDs = append(sessionIDs, session.ID)
	}
	sort.Strings(sessionIDs)
	return RuntimeSummary{
		ActiveSessions:  len(active),
		ActiveSessionID: sessionIDs,
	}
}
