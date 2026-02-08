package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
)

func newVMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Manage VM profiles",
		Long:  `Create, list, update, and delete VM profiles (configuration templates).`,
	}

	cmd.AddCommand(
		newVMCreateCommand(),
		newVMListCommand(),
		newVMGetCommand(),
		newVMDeleteCommand(),
		newVMSetSettingsCommand(),
		newVMAddCapabilityCommand(),
		newVMListCapabilitiesCommand(),
		newVMAddStartupFileCommand(),
		newVMListStartupFilesCommand(),
	)

	return cmd
}

func newVMCreateCommand() *cobra.Command {
	var name, engine string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new VM profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			vm := &vmmodels.VM{
				ID:        uuid.New().String(),
				Name:      name,
				Engine:    engine,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := store.CreateVM(vm); err != nil {
				return err
			}

			// Set default settings
			defaultLimits := vmmodels.LimitsConfig{
				CPUMs:       2000,
				WallMs:      5000,
				MemMB:       128,
				MaxEvents:   50000,
				MaxOutputKB: 256,
			}
			defaultResolver := vmmodels.ResolverConfig{
				Roots:                   []string{"."},
				Extensions:              []string{".js", ".mjs"},
				AllowAbsoluteRepoImports: true,
			}
			defaultRuntime := vmmodels.RuntimeConfig{
				ESM:     true,
				Strict:  true,
				Console: true,
			}

			limitsJSON, _ := json.Marshal(defaultLimits)
			resolverJSON, _ := json.Marshal(defaultResolver)
			runtimeJSON, _ := json.Marshal(defaultRuntime)

			settings := &vmmodels.VMSettings{
				VMID:     vm.ID,
				Limits:   limitsJSON,
				Resolver: resolverJSON,
				Runtime:  runtimeJSON,
			}

			if err := store.SetVMSettings(settings); err != nil {
				return err
			}

			fmt.Printf("Created VM profile: %s (ID: %s)\n", vm.Name, vm.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "VM profile name (required)")
	cmd.Flags().StringVar(&engine, "engine", "goja", "Engine type (goja, quickjs, node, custom)")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newVMListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all VM profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			vms, err := store.ListVMs()
			if err != nil {
				return err
			}

			if len(vms) == 0 {
				fmt.Println("No VM profiles found")
				return nil
			}

			fmt.Printf("%-36s %-20s %-10s %-10s\n", "ID", "Name", "Engine", "Active")
			fmt.Println("------------------------------------------------------------------------------------")
			for _, vm := range vms {
				active := "no"
				if vm.IsActive {
					active = "yes"
				}
				fmt.Printf("%-36s %-20s %-10s %-10s\n", vm.ID, vm.Name, vm.Engine, active)
			}

			return nil
		},
	}
}

func newVMGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [vm-id]",
		Short: "Get VM profile details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			vmID := args[0]

			vm, err := store.GetVM(vmID)
			if err != nil {
				return err
			}

			settings, err := store.GetVMSettings(vmID)
			if err != nil && err != vmmodels.ErrVMNotFound {
				return err
			}

			capabilities, err := store.ListCapabilities(vmID)
			if err != nil {
				return err
			}

			startupFiles, err := store.ListStartupFiles(vmID)
			if err != nil {
				return err
			}

			// Print VM details
			fmt.Printf("VM Profile: %s\n", vm.Name)
			fmt.Printf("ID: %s\n", vm.ID)
			fmt.Printf("Engine: %s\n", vm.Engine)
			fmt.Printf("Active: %v\n", vm.IsActive)
			fmt.Printf("Created: %s\n", vm.CreatedAt.Format(time.RFC3339))
			fmt.Printf("Updated: %s\n", vm.UpdatedAt.Format(time.RFC3339))

			if settings != nil {
				fmt.Println("\nSettings:")
				fmt.Printf("  Limits: %s\n", string(settings.Limits))
				fmt.Printf("  Resolver: %s\n", string(settings.Resolver))
				fmt.Printf("  Runtime: %s\n", string(settings.Runtime))
			}

			if len(capabilities) > 0 {
				fmt.Println("\nCapabilities:")
				for _, cap := range capabilities {
					enabled := "disabled"
					if cap.Enabled {
						enabled = "enabled"
					}
					fmt.Printf("  [%s] %s:%s (config: %s)\n", enabled, cap.Kind, cap.Name, string(cap.Config))
				}
			}

			if len(startupFiles) > 0 {
				fmt.Println("\nStartup Files:")
				for _, file := range startupFiles {
					fmt.Printf("  [%d] %s (mode: %s)\n", file.OrderIndex, file.Path, file.Mode)
				}
			}

			// Display exposed modules
			if len(vm.ExposedModules) > 0 {
				fmt.Println("\nExposed Modules:")
				for _, module := range vm.ExposedModules {
					fmt.Printf("  - %s\n", module)
				}
			}

			// Display loaded libraries
			if len(vm.Libraries) > 0 {
				fmt.Println("\nLoaded Libraries:")
				for _, lib := range vm.Libraries {
					fmt.Printf("  - %s\n", lib)
				}
			}

			return nil
		},
	}
}

func newVMDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [vm-id]",
		Short: "Delete a VM profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			vmID := args[0]

			if err := store.DeleteVM(vmID); err != nil {
				return err
			}

			fmt.Printf("Deleted VM profile: %s\n", vmID)
			return nil
		},
	}
}

func newVMSetSettingsCommand() *cobra.Command {
	var limitsJSON, resolverJSON, runtimeJSON string

	cmd := &cobra.Command{
		Use:   "set-settings [vm-id]",
		Short: "Set VM settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			vmID := args[0]

			// Verify VM exists
			if _, err := store.GetVM(vmID); err != nil {
				return err
			}

			settings := &vmmodels.VMSettings{
				VMID:     vmID,
				Limits:   json.RawMessage(limitsJSON),
				Resolver: json.RawMessage(resolverJSON),
				Runtime:  json.RawMessage(runtimeJSON),
			}

			if err := store.SetVMSettings(settings); err != nil {
				return err
			}

			fmt.Printf("Updated settings for VM: %s\n", vmID)
			return nil
		},
	}

	cmd.Flags().StringVar(&limitsJSON, "limits", `{"cpu_ms":2000,"wall_ms":5000,"mem_mb":128,"max_events":50000,"max_output_kb":256}`, "Limits config (JSON)")
	cmd.Flags().StringVar(&resolverJSON, "resolver", `{"roots":["."],"extensions":[".js",".mjs"],"allow_absolute_repo_imports":true}`, "Resolver config (JSON)")
	cmd.Flags().StringVar(&runtimeJSON, "runtime", `{"esm":true,"strict":true,"console":true}`, "Runtime config (JSON)")

	return cmd
}

func newVMAddCapabilityCommand() *cobra.Command {
	var kind, name, configJSON string
	var enabled bool

	cmd := &cobra.Command{
		Use:   "add-capability [vm-id]",
		Short: "Add a capability to a VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			vmID := args[0]

			// Verify VM exists
			if _, err := store.GetVM(vmID); err != nil {
				return err
			}

			cap := &vmmodels.VMCapability{
				ID:      uuid.New().String(),
				VMID:    vmID,
				Kind:    kind,
				Name:    name,
				Enabled: enabled,
				Config:  json.RawMessage(configJSON),
			}

			if err := store.AddCapability(cap); err != nil {
				return err
			}

			fmt.Printf("Added capability: %s:%s to VM %s\n", kind, name, vmID)
			return nil
		},
	}

	cmd.Flags().StringVar(&kind, "kind", "module", "Capability kind (module, global, fs, net, env)")
	cmd.Flags().StringVar(&name, "name", "", "Capability name (required)")
	cmd.Flags().StringVar(&configJSON, "config", "{}", "Capability config (JSON)")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable capability")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newVMListCapabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-capabilities [vm-id]",
		Short: "List capabilities for a VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			vmID := args[0]

			capabilities, err := store.ListCapabilities(vmID)
			if err != nil {
				return err
			}

			if len(capabilities) == 0 {
				fmt.Println("No capabilities found")
				return nil
			}

			fmt.Printf("%-10s %-20s %-10s %-30s\n", "Kind", "Name", "Enabled", "Config")
			fmt.Println("--------------------------------------------------------------------------------")
			for _, cap := range capabilities {
				enabled := "no"
				if cap.Enabled {
					enabled = "yes"
				}
				configStr := string(cap.Config)
				if len(configStr) > 30 {
					configStr = configStr[:27] + "..."
				}
				fmt.Printf("%-10s %-20s %-10s %-30s\n", cap.Kind, cap.Name, enabled, configStr)
			}

			return nil
		},
	}
}

func newVMAddStartupFileCommand() *cobra.Command {
	var path, mode string
	var orderIndex int

	cmd := &cobra.Command{
		Use:   "add-startup [vm-id]",
		Short: "Add a startup file to a VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			vmID := args[0]

			// Verify VM exists
			if _, err := store.GetVM(vmID); err != nil {
				return err
			}

			file := &vmmodels.VMStartupFile{
				ID:         uuid.New().String(),
				VMID:       vmID,
				Path:       path,
				OrderIndex: orderIndex,
				Mode:       mode,
			}

			if err := store.AddStartupFile(file); err != nil {
				return err
			}

			fmt.Printf("Added startup file: %s (order: %d) to VM %s\n", path, orderIndex, vmID)
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "File path (required)")
	cmd.Flags().StringVar(&mode, "mode", "eval", "Mode (eval or import)")
	cmd.Flags().IntVar(&orderIndex, "order", 10, "Order index")
	cmd.MarkFlagRequired("path")

	return cmd
}

func newVMListStartupFilesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-startup [vm-id]",
		Short: "List startup files for a VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			vmID := args[0]

			files, err := store.ListStartupFiles(vmID)
			if err != nil {
				return err
			}

			if len(files) == 0 {
				fmt.Println("No startup files found")
				return nil
			}

			fmt.Printf("%-5s %-10s %-50s\n", "Order", "Mode", "Path")
			fmt.Println("----------------------------------------------------------------------")
			for _, file := range files {
				fmt.Printf("%-5d %-10s %-50s\n", file.OrderIndex, file.Mode, file.Path)
			}

			return nil
		},
	}
}
