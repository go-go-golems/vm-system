package vmcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmmodules"
)

// TemplateService owns template CRUD and policy metadata operations.
type TemplateService struct {
	store TemplateStorePort
}

func NewTemplateService(store TemplateStorePort) *TemplateService {
	return &TemplateService{store: store}
}

func (s *TemplateService) Create(_ context.Context, input CreateTemplateInput) (*vmmodels.VM, error) {
	engine := input.Engine
	if engine == "" {
		engine = "goja"
	}

	now := time.Now()
	vm := &vmmodels.VM{
		ID:        uuid.NewString(),
		Name:      input.Name,
		Engine:    engine,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.CreateVM(vm); err != nil {
		return nil, err
	}

	settings := &vmmodels.VMSettings{
		VMID: vm.ID,
		Limits: vmmodels.MarshalJSONWithFallback(vmmodels.LimitsConfig{
			CPUMs:       2000,
			WallMs:      5000,
			MemMB:       128,
			MaxEvents:   50000,
			MaxOutputKB: 256,
		}, json.RawMessage("{}")),
		Resolver: vmmodels.MarshalJSONWithFallback(vmmodels.ResolverConfig{
			Roots:                    []string{"."},
			Extensions:               []string{".js", ".mjs"},
			AllowAbsoluteRepoImports: true,
		}, json.RawMessage("{}")),
		Runtime: vmmodels.MarshalJSONWithFallback(vmmodels.RuntimeConfig{
			ESM:     true,
			Strict:  true,
			Console: true,
		}, json.RawMessage("{}")),
	}
	if err := s.store.SetVMSettings(settings); err != nil {
		return nil, err
	}

	return vm, nil
}

func (s *TemplateService) List(_ context.Context) ([]*vmmodels.VM, error) {
	return s.store.ListVMs()
}

func (s *TemplateService) Get(_ context.Context, templateID string) (*vmmodels.VM, error) {
	return s.store.GetVM(templateID)
}

func (s *TemplateService) Delete(_ context.Context, templateID string) error {
	return s.store.DeleteVM(templateID)
}

func (s *TemplateService) SetSettings(_ context.Context, settings *vmmodels.VMSettings) error {
	return s.store.SetVMSettings(settings)
}

func (s *TemplateService) GetSettings(_ context.Context, templateID string) (*vmmodels.VMSettings, error) {
	return s.store.GetVMSettings(templateID)
}

func (s *TemplateService) AddCapability(_ context.Context, cap *vmmodels.VMCapability) error {
	return s.store.AddCapability(cap)
}

func (s *TemplateService) ListCapabilities(_ context.Context, templateID string) ([]*vmmodels.VMCapability, error) {
	return s.store.ListCapabilities(templateID)
}

func (s *TemplateService) AddStartupFile(_ context.Context, file *vmmodels.VMStartupFile) error {
	mode := strings.ToLower(strings.TrimSpace(file.Mode))
	if mode == "" {
		mode = "eval"
	}
	if mode != "eval" {
		return fmt.Errorf("%w: %s", vmmodels.ErrStartupModeUnsupported, mode)
	}
	file.Mode = mode
	return s.store.AddStartupFile(file)
}

func (s *TemplateService) ListStartupFiles(_ context.Context, templateID string) ([]*vmmodels.VMStartupFile, error) {
	return s.store.ListStartupFiles(templateID)
}

func (s *TemplateService) ListModules(_ context.Context, templateID string) ([]string, error) {
	template, err := s.store.GetVM(templateID)
	if err != nil {
		return nil, err
	}
	modules := make([]string, len(template.ExposedModules))
	copy(modules, template.ExposedModules)
	return modules, nil
}

func (s *TemplateService) AddModule(_ context.Context, templateID, moduleName string) error {
	moduleName, err := vmmodules.ValidateConfiguredModuleName(moduleName)
	if err != nil {
		return err
	}

	template, err := s.store.GetVM(templateID)
	if err != nil {
		return err
	}

	for _, existing := range template.ExposedModules {
		if existing == moduleName {
			return nil
		}
	}

	template.ExposedModules = append(template.ExposedModules, moduleName)
	return s.store.UpdateVM(template)
}

func (s *TemplateService) RemoveModule(_ context.Context, templateID, moduleName string) error {
	template, err := s.store.GetVM(templateID)
	if err != nil {
		return err
	}

	filtered := make([]string, 0, len(template.ExposedModules))
	changed := false
	for _, existing := range template.ExposedModules {
		if existing == moduleName {
			changed = true
			continue
		}
		filtered = append(filtered, existing)
	}
	if !changed {
		return nil
	}

	template.ExposedModules = filtered
	return s.store.UpdateVM(template)
}

func (s *TemplateService) ListLibraries(_ context.Context, templateID string) ([]string, error) {
	template, err := s.store.GetVM(templateID)
	if err != nil {
		return nil, err
	}
	libraries := make([]string, len(template.Libraries))
	copy(libraries, template.Libraries)
	return libraries, nil
}

func (s *TemplateService) AddLibrary(_ context.Context, templateID, libraryName string) error {
	template, err := s.store.GetVM(templateID)
	if err != nil {
		return err
	}

	for _, existing := range template.Libraries {
		if existing == libraryName {
			return nil
		}
	}

	template.Libraries = append(template.Libraries, libraryName)
	return s.store.UpdateVM(template)
}

func (s *TemplateService) RemoveLibrary(_ context.Context, templateID, libraryName string) error {
	template, err := s.store.GetVM(templateID)
	if err != nil {
		return err
	}

	filtered := make([]string, 0, len(template.Libraries))
	changed := false
	for _, existing := range template.Libraries {
		if existing == libraryName {
			changed = true
			continue
		}
		filtered = append(filtered, existing)
	}
	if !changed {
		return nil
	}

	template.Libraries = filtered
	return s.store.UpdateVM(template)
}
