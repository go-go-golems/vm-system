package vmhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	stdhttp "net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmpath"
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

type createTemplateRequest struct {
	Name   string `json:"name"`
	Engine string `json:"engine"`
}

func (s *Server) handleTemplateCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req createTemplateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if req.Name == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "name is required", nil)
		return
	}

	template, err := s.core.Templates.Create(r.Context(), vmcontrol.CreateTemplateInput{
		Name:   req.Name,
		Engine: req.Engine,
	})
	if err != nil {
		writeCoreError(w, err, nil)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, template)
}

func (s *Server) handleTemplateList(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templates, err := s.core.Templates.List(r.Context())
	if err != nil {
		writeCoreError(w, err, nil)
		return
	}
	writeJSON(w, stdhttp.StatusOK, templates)
}

type templateDetailResponse struct {
	Template     *vmmodels.VM              `json:"template"`
	Settings     *vmmodels.VMSettings      `json:"settings,omitempty"`
	Capabilities []*vmmodels.VMCapability  `json:"capabilities"`
	StartupFiles []*vmmodels.VMStartupFile `json:"startup_files"`
}

func (s *Server) handleTemplateGet(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}

	template, err := s.core.Templates.Get(r.Context(), templateID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}

	settings, err := s.core.Templates.GetSettings(r.Context(), templateID.String())
	if err != nil && !errors.Is(err, vmmodels.ErrVMNotFound) {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}

	caps, err := s.core.Templates.ListCapabilities(r.Context(), templateID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}

	startup, err := s.core.Templates.ListStartupFiles(r.Context(), templateID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}

	writeJSON(w, stdhttp.StatusOK, templateDetailResponse{
		Template:     template,
		Settings:     settings,
		Capabilities: caps,
		StartupFiles: startup,
	})
}

func (s *Server) handleTemplateDelete(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}
	if err := s.core.Templates.Delete(r.Context(), templateID.String()); err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]string{
		"status":      "ok",
		"template_id": templateID.String(),
	})
}

type addCapabilityRequest struct {
	Kind    string          `json:"kind"`
	Name    string          `json:"name"`
	Enabled bool            `json:"enabled"`
	Config  json.RawMessage `json:"config"`
}

func (s *Server) handleTemplateAddCapability(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}

	var req addCapabilityRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if req.Kind == "" || req.Name == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "kind and name are required", nil)
		return
	}
	if len(req.Config) == 0 {
		req.Config = json.RawMessage("{}")
	}

	cap := &vmmodels.VMCapability{
		ID:      uuid.NewString(),
		VMID:    templateID.String(),
		Kind:    req.Kind,
		Name:    req.Name,
		Enabled: req.Enabled,
		Config:  req.Config,
	}
	if err := s.core.Templates.AddCapability(r.Context(), cap); err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusCreated, cap)
}

func (s *Server) handleTemplateListCapabilities(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}
	caps, err := s.core.Templates.ListCapabilities(r.Context(), templateID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, caps)
}

type addTemplateModuleRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleTemplateListModules(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}
	modules, err := s.core.Templates.ListModules(r.Context(), templateID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, modules)
}

func (s *Server) handleTemplateAddModule(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}

	var req addTemplateModuleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if req.Name == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "name is required", nil)
		return
	}

	if err := s.core.Templates.AddModule(r.Context(), templateID.String(), req.Name); err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}

	writeJSON(w, stdhttp.StatusCreated, map[string]string{
		"template_id": templateID.String(),
		"name":        req.Name,
	})
}

func (s *Server) handleTemplateRemoveModule(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}
	moduleName := r.PathValue("module_name")
	if moduleName == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "module_name is required", nil)
		return
	}

	if err := s.core.Templates.RemoveModule(r.Context(), templateID.String(), moduleName); err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]string{
		"status":      "ok",
		"template_id": templateID.String(),
		"name":        moduleName,
	})
}

type addTemplateLibraryRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleTemplateListLibraries(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}
	libraries, err := s.core.Templates.ListLibraries(r.Context(), templateID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, libraries)
}

func (s *Server) handleTemplateAddLibrary(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}

	var req addTemplateLibraryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if req.Name == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "name is required", nil)
		return
	}

	if err := s.core.Templates.AddLibrary(r.Context(), templateID.String(), req.Name); err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}

	writeJSON(w, stdhttp.StatusCreated, map[string]string{
		"template_id": templateID.String(),
		"name":        req.Name,
	})
}

func (s *Server) handleTemplateRemoveLibrary(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}
	libraryName := r.PathValue("library_name")
	if libraryName == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "library_name is required", nil)
		return
	}

	if err := s.core.Templates.RemoveLibrary(r.Context(), templateID.String(), libraryName); err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]string{
		"status":      "ok",
		"template_id": templateID.String(),
		"name":        libraryName,
	})
}

type addStartupFileRequest struct {
	Path       string `json:"path"`
	OrderIndex int    `json:"order_index"`
	Mode       string `json:"mode"`
}

func (s *Server) handleTemplateAddStartupFile(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}

	var req addStartupFileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if req.Path == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "path is required", nil)
		return
	}
	parsedPath, err := vmpath.ParseRelWorktreePath(req.Path)
	if err != nil {
		switch {
		case errors.Is(err, vmpath.ErrAbsoluteRelativePath), errors.Is(err, vmpath.ErrTraversalRelativePath), errors.Is(err, vmpath.ErrEmptyRelativePath):
			writeError(w, stdhttp.StatusUnprocessableEntity, "INVALID_PATH", "Path escapes allowed worktree", map[string]string{"template_id": templateID.String()})
		default:
			writeError(w, stdhttp.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		}
		return
	}
	if req.Mode == "" {
		req.Mode = "eval"
	}
	req.Mode = strings.ToLower(strings.TrimSpace(req.Mode))
	if req.Mode != "eval" {
		writeError(w, stdhttp.StatusUnprocessableEntity, "STARTUP_MODE_UNSUPPORTED", "Only startup mode 'eval' is currently supported", map[string]interface{}{
			"template_id":      templateID.String(),
			"requested_mode":   req.Mode,
			"supported_modes":  []string{"eval"},
			"migration_option": "Use mode=eval until import support is implemented",
		})
		return
	}

	startup := &vmmodels.VMStartupFile{
		ID:         uuid.NewString(),
		VMID:       templateID.String(),
		Path:       parsedPath.String(),
		OrderIndex: req.OrderIndex,
		Mode:       req.Mode,
	}
	if err := s.core.Templates.AddStartupFile(r.Context(), startup); err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}

	writeJSON(w, stdhttp.StatusCreated, startup)
}

func (s *Server) handleTemplateListStartupFiles(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	templateID, ok := parseTemplateIDOrWriteValidationError(w, r.PathValue("template_id"))
	if !ok {
		return
	}
	files, err := s.core.Templates.ListStartupFiles(r.Context(), templateID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, files)
}

type createSessionRequest struct {
	TemplateID    string `json:"template_id"`
	WorkspaceID   string `json:"workspace_id"`
	BaseCommitOID string `json:"base_commit_oid"`
	WorktreePath  string `json:"worktree_path"`
}

func (s *Server) handleSessionCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req createSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if req.TemplateID == "" || req.WorkspaceID == "" || req.BaseCommitOID == "" || req.WorktreePath == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "template_id, workspace_id, base_commit_oid, and worktree_path are required", nil)
		return
	}
	templateID, err := vmmodels.ParseTemplateID(req.TemplateID)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "template_id must be a valid UUID", nil)
		return
	}

	session, err := s.core.Sessions.Create(r.Context(), vmcontrol.CreateSessionInput{
		TemplateID:    templateID.String(),
		WorkspaceID:   req.WorkspaceID,
		BaseCommitOID: req.BaseCommitOID,
		WorktreePath:  req.WorktreePath,
	})
	if err != nil {
		writeCoreError(w, err, map[string]string{"template_id": templateID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusCreated, session)
}

func (s *Server) handleSessionList(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	status := r.URL.Query().Get("status")
	sessions, err := s.core.Sessions.List(r.Context(), status)
	if err != nil {
		writeCoreError(w, err, nil)
		return
	}
	writeJSON(w, stdhttp.StatusOK, sessions)
}

func (s *Server) handleSessionGet(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	sessionID, ok := parseSessionIDOrWriteValidationError(w, r.PathValue("session_id"))
	if !ok {
		return
	}
	session, err := s.core.Sessions.Get(r.Context(), sessionID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"session_id": sessionID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, session)
}

func (s *Server) handleSessionClose(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	sessionID, ok := parseSessionIDOrWriteValidationError(w, r.PathValue("session_id"))
	if !ok {
		return
	}
	session, err := s.core.Sessions.Close(r.Context(), sessionID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"session_id": sessionID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, session)
}

func (s *Server) handleSessionDelete(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.handleSessionClose(w, r)
}

type executeREPLRequest struct {
	SessionID string `json:"session_id"`
	Input     string `json:"input"`
}

func (s *Server) handleExecutionREPL(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req executeREPLRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if req.SessionID == "" || req.Input == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "session_id and input are required", nil)
		return
	}
	sessionID, err := vmmodels.ParseSessionID(req.SessionID)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "session_id must be a valid UUID", nil)
		return
	}

	exec, err := s.core.Executions.ExecuteREPL(r.Context(), vmcontrol.ExecuteREPLInput{
		SessionID: sessionID.String(),
		Input:     req.Input,
	})
	if err != nil {
		writeCoreError(w, err, map[string]string{"session_id": sessionID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusCreated, exec)
}

type executeRunFileRequest struct {
	SessionID string                 `json:"session_id"`
	Path      string                 `json:"path"`
	Args      map[string]interface{} `json:"args"`
	Env       map[string]interface{} `json:"env"`
}

func (s *Server) handleExecutionRunFile(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req executeRunFileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if req.SessionID == "" || req.Path == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "session_id and path are required", nil)
		return
	}
	sessionID, err := vmmodels.ParseSessionID(req.SessionID)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "session_id must be a valid UUID", nil)
		return
	}

	exec, err := s.core.Executions.ExecuteRunFile(r.Context(), vmcontrol.ExecuteRunFileInput{
		SessionID: sessionID.String(),
		Path:      req.Path,
		Args:      req.Args,
		Env:       req.Env,
	})
	if err != nil {
		writeCoreError(w, err, map[string]string{"session_id": sessionID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusCreated, exec)
}

func (s *Server) handleExecutionGet(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	executionID, ok := parseExecutionIDOrWriteValidationError(w, r.PathValue("execution_id"))
	if !ok {
		return
	}
	exec, err := s.core.Executions.Get(r.Context(), executionID.String())
	if err != nil {
		writeCoreError(w, err, map[string]string{"execution_id": executionID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, exec)
}

func (s *Server) handleExecutionList(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "session_id query param is required", nil)
		return
	}
	parsedSessionID, err := vmmodels.ParseSessionID(sessionID)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "session_id must be a valid UUID", nil)
		return
	}

	limit := 50
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "limit must be a positive integer", nil)
			return
		}
		limit = parsed
	}

	execs, err := s.core.Executions.List(r.Context(), parsedSessionID.String(), limit)
	if err != nil {
		writeCoreError(w, err, map[string]string{"session_id": parsedSessionID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, execs)
}

func (s *Server) handleExecutionEvents(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	executionID, ok := parseExecutionIDOrWriteValidationError(w, r.PathValue("execution_id"))
	if !ok {
		return
	}
	afterSeq := 0
	if rawAfter := r.URL.Query().Get("after_seq"); rawAfter != "" {
		parsed, err := strconv.Atoi(rawAfter)
		if err != nil || parsed < 0 {
			writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "after_seq must be a non-negative integer", nil)
			return
		}
		afterSeq = parsed
	}

	events, err := s.core.Executions.Events(r.Context(), executionID.String(), afterSeq)
	if err != nil {
		writeCoreError(w, err, map[string]string{"execution_id": executionID.String()})
		return
	}
	writeJSON(w, stdhttp.StatusOK, events)
}

func parseTemplateIDOrWriteValidationError(w stdhttp.ResponseWriter, raw string) (vmmodels.TemplateID, bool) {
	id, err := vmmodels.ParseTemplateID(raw)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "template_id must be a valid UUID", nil)
		return "", false
	}
	return id, true
}

func parseSessionIDOrWriteValidationError(w stdhttp.ResponseWriter, raw string) (vmmodels.SessionID, bool) {
	id, err := vmmodels.ParseSessionID(raw)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "session_id must be a valid UUID", nil)
		return "", false
	}
	return id, true
}

func parseExecutionIDOrWriteValidationError(w stdhttp.ResponseWriter, raw string) (vmmodels.ExecutionID, bool) {
	id, err := vmmodels.ParseExecutionID(raw)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, "VALIDATION_ERROR", "execution_id must be a valid UUID", nil)
		return "", false
	}
	return id, true
}

type errorEnvelope struct {
	Error struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error"`
}

func writeCoreError(w stdhttp.ResponseWriter, err error, details interface{}) {
	switch {
	case errors.Is(err, vmmodels.ErrVMNotFound):
		writeError(w, stdhttp.StatusNotFound, "TEMPLATE_NOT_FOUND", "Template not found", details)
	case errors.Is(err, vmmodels.ErrSessionNotFound):
		writeError(w, stdhttp.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", details)
	case errors.Is(err, vmmodels.ErrExecutionNotFound):
		writeError(w, stdhttp.StatusNotFound, "EXECUTION_NOT_FOUND", "Execution not found", details)
	case errors.Is(err, vmmodels.ErrSessionNotReady):
		writeError(w, stdhttp.StatusConflict, "SESSION_NOT_READY", "Session is not ready", details)
	case errors.Is(err, vmmodels.ErrSessionBusy):
		writeError(w, stdhttp.StatusConflict, "SESSION_BUSY", "Session is busy", details)
	case errors.Is(err, vmmodels.ErrPathTraversal):
		writeError(w, stdhttp.StatusUnprocessableEntity, "INVALID_PATH", "Path escapes allowed worktree", details)
	case errors.Is(err, vmmodels.ErrOutputLimitExceeded):
		writeError(w, stdhttp.StatusUnprocessableEntity, "OUTPUT_LIMIT_EXCEEDED", "Execution exceeded configured output/event limits", details)
	case errors.Is(err, vmmodels.ErrStartupModeUnsupported):
		writeError(w, stdhttp.StatusUnprocessableEntity, "STARTUP_MODE_UNSUPPORTED", "Only startup mode 'eval' is currently supported", details)
	case errors.Is(err, vmmodels.ErrModuleNotAllowed):
		writeError(w, stdhttp.StatusUnprocessableEntity, "MODULE_NOT_ALLOWED", "Module is not allowed for template configuration", details)
	case errors.Is(err, vmmodels.ErrFileNotFound):
		writeError(w, stdhttp.StatusNotFound, "FILE_NOT_FOUND", "File not found", details)
	default:
		writeError(w, stdhttp.StatusInternalServerError, "INTERNAL", err.Error(), details)
	}
}

func writeError(w stdhttp.ResponseWriter, status int, code, message string, details interface{}) {
	env := errorEnvelope{}
	env.Error.Code = code
	env.Error.Message = message
	env.Error.Details = details
	writeJSON(w, status, env)
}

func writeJSON(w stdhttp.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		stdhttp.Error(w, fmt.Sprintf("encode response: %v", err), stdhttp.StatusInternalServerError)
	}
}

func decodeJSON(r *stdhttp.Request, out interface{}) error {
	if r.Body == nil {
		return errors.New("empty request body")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}
