package vmhttp

import (
	stdhttp "net/http"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

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
