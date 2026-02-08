package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/spf13/cobra"
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
		templateID, _ := cmd.Flags().GetString("vm-id")
		moduleID, _ := cmd.Flags().GetString("module-id")

		if templateID == "" || moduleID == "" {
			return fmt.Errorf("--vm-id and --module-id are required")
		}

		client := vmclient.New(serverURL, nil)
		if _, err := client.AddTemplateModule(context.Background(), templateID, vmclient.AddTemplateModuleRequest{
			Name: moduleID,
		}); err != nil {
			return err
		}

		fmt.Printf("Added module '%s' to template '%s'\n", moduleID, templateID)
		return nil
	},
}

var addLibraryCmd = &cobra.Command{
	Use:   "add-library",
	Short: "Add a library to a VM",
	RunE: func(cmd *cobra.Command, args []string) error {
		templateID, _ := cmd.Flags().GetString("vm-id")
		libraryID, _ := cmd.Flags().GetString("library-id")

		if templateID == "" || libraryID == "" {
			return fmt.Errorf("--vm-id and --library-id are required")
		}

		client := vmclient.New(serverURL, nil)
		if _, err := client.AddTemplateLibrary(context.Background(), templateID, vmclient.AddTemplateLibraryRequest{
			Name: libraryID,
		}); err != nil {
			return err
		}

		fmt.Printf("Added library '%s' to template '%s'\n", libraryID, templateID)
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
