package vmhttp

import (
	stdhttp "net/http"
	"strconv"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

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
