package vmhttp

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmpath"
)

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
