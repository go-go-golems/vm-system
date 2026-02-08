package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
)

var modulesCmd = &cobra.Command{
	Use:   "modules",
	Short: "Manage VM modules and libraries",
}

var listModulesCmd = &cobra.Command{
	Use:   "list-available",
	Short: "List available built-in modules",
	RunE: func(cmd *cobra.Command, args []string) error {
		modules := vmmodels.BuiltinModules()
		data, err := json.MarshalIndent(modules, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var listLibrariesCmd = &cobra.Command{
	Use:   "list-libraries",
	Short: "List available built-in libraries",
	RunE: func(cmd *cobra.Command, args []string) error {
		libraries := vmmodels.BuiltinLibraries()
		data, err := json.MarshalIndent(libraries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var addModuleCmd = &cobra.Command{
	Use:   "add-module",
	Short: "Add an exposed module to a VM",
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID, _ := cmd.Flags().GetString("vm-id")
		moduleID, _ := cmd.Flags().GetString("module-id")

		if vmID == "" || moduleID == "" {
			return fmt.Errorf("--vm-id and --module-id are required")
		}

		store, err := vmstore.NewVMStore(dbPath)
		if err != nil {
			return err
		}
		defer store.Close()

		// Get VM
		vm, err := store.GetVM(vmID)
		if err != nil {
			return err
		}

		// Check if module is already added
		for _, m := range vm.ExposedModules {
			if m == moduleID {
				fmt.Printf("Module '%s' is already added to VM '%s'\n", moduleID, vm.Name)
				return nil
			}
		}

		// Add module
		vm.ExposedModules = append(vm.ExposedModules, moduleID)
		if err := store.UpdateVM(vm); err != nil {
			return err
		}

		fmt.Printf("Added module '%s' to VM '%s'\n", moduleID, vm.Name)
		return nil
	},
}

var addLibraryCmd = &cobra.Command{
	Use:   "add-library",
	Short: "Add a library to a VM",
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID, _ := cmd.Flags().GetString("vm-id")
		libraryID, _ := cmd.Flags().GetString("library-id")

		if vmID == "" || libraryID == "" {
			return fmt.Errorf("--vm-id and --library-id are required")
		}

		store, err := vmstore.NewVMStore(dbPath)
		if err != nil {
			return err
		}
		defer store.Close()

		// Get VM
		vm, err := store.GetVM(vmID)
		if err != nil {
			return err
		}

		// Check if library is already added
		for _, l := range vm.Libraries {
			if l == libraryID {
				fmt.Printf("Library '%s' is already added to VM '%s'\n", libraryID, vm.Name)
				return nil
			}
		}

		// Add library
		vm.Libraries = append(vm.Libraries, libraryID)
		if err := store.UpdateVM(vm); err != nil {
			return err
		}

		fmt.Printf("Added library '%s' to VM '%s'\n", libraryID, vm.Name)
		return nil
	},
}

func init() {
	modulesCmd.AddCommand(listModulesCmd)
	modulesCmd.AddCommand(listLibrariesCmd)
	modulesCmd.AddCommand(addModuleCmd)
	modulesCmd.AddCommand(addLibraryCmd)

	addModuleCmd.Flags().String("vm-id", "", "VM ID")
	addModuleCmd.Flags().String("module-id", "", "Module ID")

	addLibraryCmd.Flags().String("vm-id", "", "VM ID")
	addLibraryCmd.Flags().String("library-id", "", "Library ID")
}
