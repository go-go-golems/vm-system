package vmstore

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

// CreateVM creates a new VM profile.
func (s *VMStore) CreateVM(vm *vmmodels.VM) error {
	_, err := s.db.Exec(`
		INSERT INTO vm (id, name, engine, is_active, exposed_modules_json, libraries_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, vm.ID, vm.Name, vm.Engine, vm.IsActive, string(vmmodels.MarshalJSONWithFallback(vm.ExposedModules, json.RawMessage("[]"))), string(vmmodels.MarshalJSONWithFallback(vm.Libraries, json.RawMessage("[]"))), vm.CreatedAt.Unix(), vm.UpdatedAt.Unix())
	return err
}

// GetVM retrieves a VM by ID.
func (s *VMStore) GetVM(id string) (*vmmodels.VM, error) {
	var vm vmmodels.VM
	var createdAt, updatedAt int64
	var exposedModulesJSON, librariesJSON string

	err := s.db.QueryRow(`
		SELECT id, name, engine, is_active, exposed_modules_json, libraries_json, created_at, updated_at
		FROM vm WHERE id = ?
	`, id).Scan(&vm.ID, &vm.Name, &vm.Engine, &vm.IsActive, &exposedModulesJSON, &librariesJSON, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, vmmodels.ErrVMNotFound
	}
	if err != nil {
		return nil, err
	}

	vm.CreatedAt = time.Unix(createdAt, 0)
	vm.UpdatedAt = time.Unix(updatedAt, 0)

	// Unmarshal JSON arrays.
	_ = json.Unmarshal([]byte(exposedModulesJSON), &vm.ExposedModules)
	_ = json.Unmarshal([]byte(librariesJSON), &vm.Libraries)

	return &vm, nil
}

// ListVMs lists all VMs.
func (s *VMStore) ListVMs() ([]*vmmodels.VM, error) {
	rows, err := s.db.Query(`
		SELECT id, name, engine, is_active, exposed_modules_json, libraries_json, created_at, updated_at
		FROM vm ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vms []*vmmodels.VM
	for rows.Next() {
		var vm vmmodels.VM
		var createdAt, updatedAt int64

		var exposedModulesJSON, librariesJSON string
		if err := rows.Scan(&vm.ID, &vm.Name, &vm.Engine, &vm.IsActive, &exposedModulesJSON, &librariesJSON, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		vm.CreatedAt = time.Unix(createdAt, 0)
		vm.UpdatedAt = time.Unix(updatedAt, 0)
		_ = json.Unmarshal([]byte(exposedModulesJSON), &vm.ExposedModules)
		_ = json.Unmarshal([]byte(librariesJSON), &vm.Libraries)
		vms = append(vms, &vm)
	}

	return vms, rows.Err()
}

// UpdateVM updates a VM profile.
func (s *VMStore) UpdateVM(vm *vmmodels.VM) error {
	vm.UpdatedAt = time.Now()
	_, err := s.db.Exec(`
		UPDATE vm SET name = ?, engine = ?, is_active = ?, exposed_modules_json = ?, libraries_json = ?, updated_at = ?
		WHERE id = ?
	`, vm.Name, vm.Engine, vm.IsActive, string(vmmodels.MarshalJSONWithFallback(vm.ExposedModules, json.RawMessage("[]"))), string(vmmodels.MarshalJSONWithFallback(vm.Libraries, json.RawMessage("[]"))), vm.UpdatedAt.Unix(), vm.ID)
	return err
}

// DeleteVM deletes a VM profile.
func (s *VMStore) DeleteVM(id string) error {
	_, err := s.db.Exec("DELETE FROM vm WHERE id = ?", id)
	return err
}

// SetVMSettings sets VM settings.
func (s *VMStore) SetVMSettings(settings *vmmodels.VMSettings) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO vm_settings (vm_id, limits_json, resolver_json, runtime_json)
		VALUES (?, ?, ?, ?)
	`, settings.VMID, settings.Limits, settings.Resolver, settings.Runtime)
	return err
}

// GetVMSettings retrieves VM settings.
func (s *VMStore) GetVMSettings(vmID string) (*vmmodels.VMSettings, error) {
	var settings vmmodels.VMSettings
	err := s.db.QueryRow(`
		SELECT vm_id, limits_json, resolver_json, runtime_json
		FROM vm_settings WHERE vm_id = ?
	`, vmID).Scan(&settings.VMID, &settings.Limits, &settings.Resolver, &settings.Runtime)

	if err == sql.ErrNoRows {
		return nil, vmmodels.ErrVMNotFound
	}
	return &settings, err
}

// AddCapability adds a capability to a VM.
func (s *VMStore) AddCapability(cap *vmmodels.VMCapability) error {
	_, err := s.db.Exec(`
		INSERT INTO vm_capability (id, vm_id, kind, name, enabled, config_json)
		VALUES (?, ?, ?, ?, ?, ?)
	`, cap.ID, cap.VMID, cap.Kind, cap.Name, cap.Enabled, cap.Config)
	return err
}

// ListCapabilities lists all capabilities for a VM.
func (s *VMStore) ListCapabilities(vmID string) ([]*vmmodels.VMCapability, error) {
	rows, err := s.db.Query(`
		SELECT id, vm_id, kind, name, enabled, config_json
		FROM vm_capability WHERE vm_id = ?
	`, vmID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caps []*vmmodels.VMCapability
	for rows.Next() {
		var cap vmmodels.VMCapability
		if err := rows.Scan(&cap.ID, &cap.VMID, &cap.Kind, &cap.Name, &cap.Enabled, &cap.Config); err != nil {
			return nil, err
		}
		caps = append(caps, &cap)
	}

	return caps, rows.Err()
}

// GetCapability retrieves a specific capability.
func (s *VMStore) GetCapability(vmID, kind, name string) (*vmmodels.VMCapability, error) {
	var cap vmmodels.VMCapability
	err := s.db.QueryRow(`
		SELECT id, vm_id, kind, name, enabled, config_json
		FROM vm_capability WHERE vm_id = ? AND kind = ? AND name = ?
	`, vmID, kind, name).Scan(&cap.ID, &cap.VMID, &cap.Kind, &cap.Name, &cap.Enabled, &cap.Config)

	if err == sql.ErrNoRows {
		return nil, vmmodels.ErrModuleNotAllowed
	}
	return &cap, err
}

// DeleteCapability deletes a capability.
func (s *VMStore) DeleteCapability(id string) error {
	_, err := s.db.Exec("DELETE FROM vm_capability WHERE id = ?", id)
	return err
}

// AddStartupFile adds a startup file to a VM.
func (s *VMStore) AddStartupFile(file *vmmodels.VMStartupFile) error {
	_, err := s.db.Exec(`
		INSERT INTO vm_startup_file (id, vm_id, path, order_index, mode)
		VALUES (?, ?, ?, ?, ?)
	`, file.ID, file.VMID, file.Path, file.OrderIndex, file.Mode)
	return err
}

// ListStartupFiles lists all startup files for a VM.
func (s *VMStore) ListStartupFiles(vmID string) ([]*vmmodels.VMStartupFile, error) {
	rows, err := s.db.Query(`
		SELECT id, vm_id, path, order_index, mode
		FROM vm_startup_file WHERE vm_id = ? ORDER BY order_index
	`, vmID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*vmmodels.VMStartupFile
	for rows.Next() {
		var file vmmodels.VMStartupFile
		if err := rows.Scan(&file.ID, &file.VMID, &file.Path, &file.OrderIndex, &file.Mode); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}

	return files, rows.Err()
}

// DeleteStartupFile deletes a startup file.
func (s *VMStore) DeleteStartupFile(id string) error {
	_, err := s.db.Exec("DELETE FROM vm_startup_file WHERE id = ?", id)
	return err
}
