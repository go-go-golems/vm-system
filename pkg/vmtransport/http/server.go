package vmhttp

import (
	stdhttp "net/http"

	"github.com/google/uuid"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
)

type Server struct {
	core *vmcontrol.Core
}

func NewHandler(core *vmcontrol.Core) stdhttp.Handler {
	s := &Server{core: core}
	mux := stdhttp.NewServeMux()

	// Health/ops.
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/runtime/summary", s.handleRuntimeSummary)

	// Template APIs.
	mux.HandleFunc("GET /api/v1/templates", s.handleTemplateList)
	mux.HandleFunc("POST /api/v1/templates", s.handleTemplateCreate)
	mux.HandleFunc("GET /api/v1/templates/{template_id}", s.handleTemplateGet)
	mux.HandleFunc("DELETE /api/v1/templates/{template_id}", s.handleTemplateDelete)
	mux.HandleFunc("GET /api/v1/templates/{template_id}/capabilities", s.handleTemplateListCapabilities)
	mux.HandleFunc("POST /api/v1/templates/{template_id}/capabilities", s.handleTemplateAddCapability)
	mux.HandleFunc("GET /api/v1/templates/{template_id}/modules", s.handleTemplateListModules)
	mux.HandleFunc("POST /api/v1/templates/{template_id}/modules", s.handleTemplateAddModule)
	mux.HandleFunc("DELETE /api/v1/templates/{template_id}/modules/{module_name}", s.handleTemplateRemoveModule)
	mux.HandleFunc("GET /api/v1/templates/{template_id}/libraries", s.handleTemplateListLibraries)
	mux.HandleFunc("POST /api/v1/templates/{template_id}/libraries", s.handleTemplateAddLibrary)
	mux.HandleFunc("DELETE /api/v1/templates/{template_id}/libraries/{library_name}", s.handleTemplateRemoveLibrary)
	mux.HandleFunc("GET /api/v1/templates/{template_id}/startup-files", s.handleTemplateListStartupFiles)
	mux.HandleFunc("POST /api/v1/templates/{template_id}/startup-files", s.handleTemplateAddStartupFile)

	// Session APIs.
	mux.HandleFunc("GET /api/v1/sessions", s.handleSessionList)
	mux.HandleFunc("POST /api/v1/sessions", s.handleSessionCreate)
	mux.HandleFunc("GET /api/v1/sessions/{session_id}", s.handleSessionGet)
	mux.HandleFunc("POST /api/v1/sessions/{session_id}/close", s.handleSessionClose)
	mux.HandleFunc("DELETE /api/v1/sessions/{session_id}", s.handleSessionDelete)

	// Execution APIs.
	mux.HandleFunc("GET /api/v1/executions", s.handleExecutionList)
	mux.HandleFunc("POST /api/v1/executions/repl", s.handleExecutionREPL)
	mux.HandleFunc("POST /api/v1/executions/run-file", s.handleExecutionRunFile)
	mux.HandleFunc("GET /api/v1/executions/{execution_id}", s.handleExecutionGet)
	mux.HandleFunc("GET /api/v1/executions/{execution_id}/events", s.handleExecutionEvents)

	return withRequestID(mux)
}

func withRequestID(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		requestID := uuid.NewString()
		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) handleRuntimeSummary(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, s.core.Registry.Summary(r.Context()))
}
