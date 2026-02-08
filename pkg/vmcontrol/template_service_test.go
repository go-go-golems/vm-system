package vmcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

func TestTemplateServiceCreateMarshalsDefaultConfigJSON(t *testing.T) {
	store := &templateStoreStub{}
	service := NewTemplateService(store)

	vm, err := service.Create(context.Background(), CreateTemplateInput{Name: "defaults-template"})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	if vm == nil || vm.ID == "" {
		t.Fatalf("expected created template with id")
	}
	if store.settings == nil {
		t.Fatalf("expected VM settings to be persisted")
	}
	if store.settings.VMID != vm.ID {
		t.Fatalf("expected settings vm_id %q, got %q", vm.ID, store.settings.VMID)
	}

	limits := vmmodels.LimitsConfig{}
	if err := json.Unmarshal(store.settings.Limits, &limits); err != nil {
		t.Fatalf("unmarshal limits: %v", err)
	}
	if limits.CPUMs != 2000 || limits.WallMs != 5000 || limits.MemMB != 128 || limits.MaxEvents != 50000 || limits.MaxOutputKB != 256 {
		t.Fatalf("unexpected limits defaults: %+v", limits)
	}

	resolver := vmmodels.ResolverConfig{}
	if err := json.Unmarshal(store.settings.Resolver, &resolver); err != nil {
		t.Fatalf("unmarshal resolver: %v", err)
	}
	if len(resolver.Roots) != 1 || resolver.Roots[0] != "." {
		t.Fatalf("unexpected resolver roots: %+v", resolver.Roots)
	}
	if len(resolver.Extensions) != 2 || resolver.Extensions[0] != ".js" || resolver.Extensions[1] != ".mjs" {
		t.Fatalf("unexpected resolver extensions: %+v", resolver.Extensions)
	}
	if !resolver.AllowAbsoluteRepoImports {
		t.Fatalf("expected allow_absolute_repo_imports=true")
	}

	runtime := vmmodels.RuntimeConfig{}
	if err := json.Unmarshal(store.settings.Runtime, &runtime); err != nil {
		t.Fatalf("unmarshal runtime: %v", err)
	}
	if !runtime.ESM || !runtime.Strict || !runtime.Console {
		t.Fatalf("unexpected runtime defaults: %+v", runtime)
	}
}

type templateStoreStub struct {
	vm       *vmmodels.VM
	settings *vmmodels.VMSettings
}

func (s *templateStoreStub) CreateVM(vm *vmmodels.VM) error {
	s.vm = vm
	return nil
}

func (s *templateStoreStub) GetVM(id string) (*vmmodels.VM, error) {
	return nil, errors.New("not implemented")
}

func (s *templateStoreStub) ListVMs() ([]*vmmodels.VM, error) {
	return nil, errors.New("not implemented")
}

func (s *templateStoreStub) DeleteVM(id string) error {
	return errors.New("not implemented")
}

func (s *templateStoreStub) SetVMSettings(settings *vmmodels.VMSettings) error {
	s.settings = settings
	return nil
}

func (s *templateStoreStub) GetVMSettings(vmID string) (*vmmodels.VMSettings, error) {
	return nil, errors.New("not implemented")
}

func (s *templateStoreStub) AddCapability(cap *vmmodels.VMCapability) error {
	return errors.New("not implemented")
}

func (s *templateStoreStub) ListCapabilities(vmID string) ([]*vmmodels.VMCapability, error) {
	return nil, errors.New("not implemented")
}

func (s *templateStoreStub) AddStartupFile(file *vmmodels.VMStartupFile) error {
	return errors.New("not implemented")
}

func (s *templateStoreStub) ListStartupFiles(vmID string) ([]*vmmodels.VMStartupFile, error) {
	return nil, errors.New("not implemented")
}
