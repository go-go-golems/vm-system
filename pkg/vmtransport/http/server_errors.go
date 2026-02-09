package vmhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	stdhttp "net/http"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

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
